/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/grafana/grafana-operator/v5/controllers/metrics"

	"github.com/go-logr/logr"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

// GrafanaDatasourceReconciler reconciles a GrafanaDatasource object
type GrafanaDatasourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/finalizers,verbs=update

func (r *GrafanaDatasourceReconciler) syncDatasources(ctx context.Context) (ctrl.Result, error) {
	syncLog := log.FromContext(ctx).WithName("GrafanaDatasourceReconciler")
	datasourcesSynced := 0

	// get all grafana instances
	grafanas := &v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, grafanas, opts...)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	// no instances, no need to sync
	if len(grafanas.Items) == 0 {
		return ctrl.Result{Requeue: false}, nil
	}

	// get all datasources
	allDatasources := &v1beta1.GrafanaDatasourceList{}
	err = r.Client.List(ctx, allDatasources, opts...)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	// sync datasources, delete datasources from grafana that do no longer have a cr
	datasourcesToDelete := map[*v1beta1.Grafana][]v1beta1.NamespacedResource{}
	for _, grafana := range grafanas.Items {
		grafana := grafana
		for _, datasource := range grafana.Status.Datasources {
			if allDatasources.Find(datasource.Namespace(), datasource.Name()) == nil {
				datasourcesToDelete[&grafana] = append(datasourcesToDelete[&grafana], datasource)
			}
		}
	}

	// delete all datasources that no longer have a cr
	for grafana, existingDatasources := range datasourcesToDelete {
		grafana := grafana
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		for _, datasource := range existingDatasources {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to ten datasources per cycle. This means that it will take longer to sync
			// a large number of deleted datasource crs, but that should be an edge case.
			if datasourcesSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := datasource.Split()
			instanceDatasource, err := grafanaClient.Datasources.GetDataSourceByUID(uid)
			if err != nil {
				var notFound *datasources.GetDataSourceByUIDNotFound
				if errors.As(err, &notFound) {
					return ctrl.Result{Requeue: false}, err
				}
				syncLog.Info("datasource no longer exists", "namespace", namespace, "name", name)
			} else {
				_, err = grafanaClient.Datasources.DeleteDataSourceByUID(instanceDatasource.Payload.UID) //nolint
				if err != nil {
					var notFound *datasources.DeleteDataSourceByUIDNotFound
					if errors.As(err, &notFound) {
						return ctrl.Result{Requeue: false}, err
					}
				}
			}

			grafana.Status.Datasources = grafana.Status.Datasources.Remove(namespace, name)
			datasourcesSynced += 1
		}

		// one update per grafana - this will trigger a reconcile of the grafana controller
		// so we should minimize those updates
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}
	}

	if datasourcesSynced > 0 {
		syncLog.Info("successfully synced datasources", "datasources", datasourcesSynced)
	}
	return ctrl.Result{Requeue: false}, nil
}

func (r *GrafanaDatasourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx).WithName("GrafanaDatasourceReconciler")
	r.Log = controllerLog

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncDatasources(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialDatasourceSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	cr := &v1beta1.GrafanaDatasource{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			err = r.onDatasourceDeleted(ctx, req.Namespace, req.Name)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, err
			}
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana datasource cr")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if cr.Spec.Datasource == nil {
		controllerLog.Info("skipped datasource with empty spec", cr.Name, cr.Namespace)
		// TODO: add a custom status around that?
		return ctrl.Result{}, nil
	}

	instances, err := r.GetMatchingDatasourceInstances(ctx, cr, r.Client)
	if err != nil {
		controllerLog.Error(err, "could not find matching instances", "name", cr.Name, "namespace", cr.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	controllerLog.Info("found matching Grafana instances for datasource", "count", len(instances.Items))

	datasource, hash, err := r.getDatasourceContent(ctx, cr)
	if err != nil {
		controllerLog.Error(err, "could not retrieve datasource contents", "name", cr.Name, "namespace", cr.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if cr.IsUpdatedUID(datasource.UID) {
		controllerLog.Info("datasource uid got updated, deleting datasources with the old uid")
		err = r.onDatasourceDeleted(ctx, req.Namespace, req.Name)
		if err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, err
		}

		// Clean up uid, so further reconcilications can track changes there
		cr.Status.UID = ""

		err = r.Client.Status().Update(ctx, cr)
		if err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, err
		}

		// Status update should trigger the next reconciliation right away, no need to requeue for dashboard creation
		return ctrl.Result{}, nil
	}

	success := true
	for _, grafana := range instances.Items {
		// check if this is a cross namespace import
		if grafana.Namespace != cr.Namespace && !cr.IsAllowCrossNamespaceImport() {
			continue
		}

		grafana := grafana
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		if grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			success = false
			continue
		}

		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested datasources to a configmap from where the
			// grafana reconciler will pick them upi
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, cr.Spec.Plugins, fmt.Sprintf("%v-datasource", cr.Name))
			if err != nil {
				success = false
				controllerLog.Error(err, "error reconciling plugins", "datasource", cr.Name, "grafana", grafana.Name)
			}
		}

		// then import the datasource into the matching grafana instances
		err = r.onDatasourceCreated(ctx, &grafana, cr, datasource, hash)
		if err != nil {
			success = false
			cr.Status.LastMessage = err.Error()
			controllerLog.Error(err, "error reconciling datasource", "datasource", cr.Name, "grafana", grafana.Name)
		}
	}

	// if the datasource was successfully synced in all instances, wait for its re-sync period
	if success {
		cr.Status.LastMessage = ""
		cr.Status.Hash = hash
		if cr.ResyncPeriodHasElapsed() {
			cr.Status.LastResync = metav1.Time{Time: time.Now()}
		}
		cr.Status.UID = datasource.UID
		return ctrl.Result{RequeueAfter: cr.GetResyncPeriod()}, r.Client.Status().Update(ctx, cr)
	} else {
		// if there was an issue with the datasource, update the status
		return ctrl.Result{RequeueAfter: RequeueDelay}, r.Client.Status().Update(ctx, cr)
	}
}

func (r *GrafanaDatasourceReconciler) onDatasourceDeleted(ctx context.Context, namespace string, name string) error {
	list := v1beta1.GrafanaList{}
	opts := []client.ListOption{}
	err := r.Client.List(ctx, &list, opts...)
	if err != nil {
		return err
	}

	for _, grafana := range list.Items {
		grafana := grafana
		if found, uid := grafana.Status.Datasources.Find(namespace, name); found {
			grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			datasource, err := grafanaClient.Datasources.GetDataSourceByUID(*uid)
			if err != nil {
				var notFound *datasources.GetDataSourceByUIDNotFound
				if errors.As(err, &notFound) {
					return err
				}
			} else {
				_, err = grafanaClient.Datasources.DeleteDataSourceByUID(datasource.Payload.UID) //nolint
				if err != nil {
					var notFound *datasources.DeleteDataSourceByUIDNotFound
					if errors.As(err, &notFound) {
						return err
					}
				}
			}

			if grafana.IsInternal() {
				err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-datasource", name))
				if err != nil {
					return err
				}
			}

			grafana.Status.Datasources = grafana.Status.Datasources.Remove(namespace, name)
			return r.Client.Status().Update(ctx, &grafana)
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) onDatasourceCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource, datasource *models.UpdateDataSourceCommand, hash string) error {
	if grafana.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your datasource cr")
	}

	if cr.Spec.Datasource == nil {
		return nil
	}

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	exists, uid, err := r.Exists(grafanaClient, datasource.UID, datasource.Name)
	if err != nil {
		return err
	}

	if exists && cr.Unchanged(hash) && !cr.ResyncPeriodHasElapsed() {
		return nil
	}

	encoded, err := json.Marshal(datasource)
	if err != nil {
		return fmt.Errorf("representing datasource as JSON: %w", err)
	}
	if exists {
		var body models.UpdateDataSourceCommand
		if err := json.Unmarshal(encoded, &body); err != nil {
			return fmt.Errorf("representing data source as update command: %w", err)
		}
		datasource.UID = uid
		_, err := grafanaClient.Datasources.UpdateDataSourceByUID(datasource.UID, &body) //nolint
		if err != nil {
			return err
		}
	} else {
		var body models.AddDataSourceCommand
		if err := json.Unmarshal(encoded, &body); err != nil {
			return fmt.Errorf("representing data source as create command: %w", err)
		}
		_, err = grafanaClient.Datasources.AddDataSource(&body) //nolint
		if err != nil {
			return err
		}
	}

	grafana.Status.Datasources = grafana.Status.Datasources.Add(cr.Namespace, cr.Name, datasource.UID)
	return r.Client.Status().Update(ctx, grafana)
}

func (r *GrafanaDatasourceReconciler) Exists(client *genapi.GrafanaHTTPAPI, uid, name string) (bool, string, error) {
	datasources, err := client.Datasources.GetDataSources()
	if err != nil {
		return false, "", fmt.Errorf("fetching data sources: %w", err)
	}

	for _, datasource := range datasources.Payload {
		if datasource.UID == uid || datasource.Name == name {
			return true, datasource.UID, nil
		}
	}

	return false, "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDatasourceReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDatasource{}).
		Complete(r)

	if err == nil {
		d, err := time.ParseDuration(initialSyncDelay)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d):
					result, err := r.Reconcile(ctx, ctrl.Request{})
					if err != nil {
						r.Log.Error(err, "error synchronizing datasources")
						continue
					}
					if result.Requeue {
						r.Log.Info("more datasources left to synchronize")
						continue
					}
					r.Log.Info("datasources sync complete")
					return
				}
			}
		}()
	}

	return err
}

func (r *GrafanaDatasourceReconciler) GetMatchingDatasourceInstances(ctx context.Context, datasource *v1beta1.GrafanaDatasource, k8sClient client.Client) (v1beta1.GrafanaList, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, datasource.Spec.InstanceSelector)
	if err != nil || len(instances.Items) == 0 {
		datasource.Status.NoMatchingInstances = true
		if err := r.Client.Status().Update(ctx, datasource); err != nil {
			r.Log.Info("unable to update the status of %v, in %v", datasource.Name, datasource.Namespace)
		}
		return v1beta1.GrafanaList{}, err
	}
	datasource.Status.NoMatchingInstances = false
	if err := r.Client.Status().Update(ctx, datasource); err != nil {
		r.Log.Info("unable to update the status of %v, in %v", datasource.Name, datasource.Namespace)
	}

	return instances, err
}

func (r *GrafanaDatasourceReconciler) getDatasourceContent(ctx context.Context, cr *v1beta1.GrafanaDatasource) (*models.UpdateDataSourceCommand, string, error) {
	initialBytes, err := json.Marshal(cr.Spec.Datasource)
	if err != nil {
		return nil, "", err
	}

	simpleContent, err := simplejson.NewJson(initialBytes)
	if err != nil {
		return nil, "", err
	}

	if cr.Spec.Datasource.UID == "" {
		simpleContent.Set("uid", string(cr.UID))
	}

	for _, ref := range cr.Spec.ValuesFrom {
		ref := ref
		val, key, err := r.getReferencedValue(ctx, cr, &ref.ValueFrom)
		if err != nil {
			return nil, "", err
		}

		patternToReplace := simpleContent.GetPath(strings.Split(ref.TargetPath, ".")...)
		patternString, err := patternToReplace.String()
		if err != nil {
			return nil, "", fmt.Errorf("pattern must be a string")
		}

		patternString = strings.ReplaceAll(patternString, fmt.Sprintf("${%v}", key), val)
		patternString = strings.ReplaceAll(patternString, fmt.Sprintf("$%v", key), val)
		simpleContent.SetPath(strings.Split(ref.TargetPath, "."), patternString)
	}

	newBytes, err := simpleContent.MarshalJSON()
	if err != nil {
		return nil, "", err
	}

	// We use UpdateDataSourceCommand here because models.DataSource lacks the SecureJsonData field
	var res models.UpdateDataSourceCommand
	if err = json.Unmarshal(newBytes, &res); err != nil {
		return nil, "", err
	}

	hash := sha256.New()
	hash.Write(newBytes)

	return &res, fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (r *GrafanaDatasourceReconciler) getReferencedValue(ctx context.Context, cr *v1beta1.GrafanaDatasource, source *v1beta1.GrafanaDatasourceValueFromSource) (string, string, error) {
	if source.SecretKeyRef != nil {
		s := &v1.Secret{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return string(val), source.SecretKeyRef.Key, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in secret %s", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
		}
	} else {
		s := &v1.ConfigMap{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: source.ConfigMapKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.ConfigMapKeyRef.Key]; ok {
			return val, source.ConfigMapKeyRef.Key, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in configmap %s", source.ConfigMapKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	}
}
