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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana-operator/grafana-operator-experimental/controllers/metrics"

	"github.com/go-logr/logr"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
	gapi "github.com/grafana/grafana-api-golang-client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
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
	syncLog := log.FromContext(ctx)
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

	// sync datasources, delete dashboards from grafana that do no longer have a cr
	datasourcesToDelete := map[*v1beta1.Grafana][]v1beta1.NamespacedResource{}
	for _, grafana := range grafanas.Items {
		for _, datasource := range grafana.Status.Datasources {
			if allDatasources.Find(datasource.Namespace(), datasource.Name()) == nil {
				datasourcesToDelete[&grafana] = append(datasourcesToDelete[&grafana], datasource)
			}
		}
	}

	// delete all dashboards that no longer have a cr
	for grafana, datasources := range datasourcesToDelete {
		grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		for _, datasource := range datasources {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to ten dashboards per cycle. This means that it will take longer to sync
			// a large number of deleted dashboard crs, but that should be an edge case.
			if datasourcesSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := datasource.Split()
			instanceDatasource, err := grafanaClient.DataSourceByUID(uid)
			if err != nil {
				if strings.Contains(err.Error(), "status: 404") {
					syncLog.Info("datasource no longer exists", "namespace", namespace, "name", name)
				} else {
					return ctrl.Result{Requeue: false}, err
				}
			}

			err = grafanaClient.DeleteDataSource(instanceDatasource.ID)
			if err != nil {
				return ctrl.Result{Requeue: false}, err
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
	controllerLog := log.FromContext(ctx)
	r.Log = controllerLog

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncDatasources(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialDatasourceSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	datasource := &v1beta1.GrafanaDatasource{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, datasource)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.onDatasourceDeleted(ctx, req.Namespace, req.Name)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelayError}, err
			}
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana dashboard cr")
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	if datasource.Spec.InstanceSelector == nil {
		controllerLog.Info("no instance selector found for datasource, nothing to do", "name", datasource.Name, "namespace", datasource.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	instances, err := GetMatchingInstances(ctx, r.Client, datasource.Spec.InstanceSelector)
	if err != nil {
		controllerLog.Error(err, "could not find matching instance", "name", datasource.Name)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	if len(instances.Items) == 0 {
		controllerLog.Info("no matching instances found for datasource", "datasource", datasource.Name, "namespace", datasource.Namespace)
	}

	controllerLog.Info("found matching Grafana instances for datasource", "count", len(instances.Items))

	success := true
	for _, grafana := range instances.Items {
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		//if grafana.Status.AdminUrl == "" || grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
		if grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		// first reconcile the plugins
		// append the requested dashboards to a configmap from where the
		// grafana reconciler will pick them upi
		err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, datasource.Spec.Plugins, fmt.Sprintf("%v-datasource", datasource.Name))
		if err != nil {
			success = false
			controllerLog.Error(err, "error reconciling plugins", "datasource", datasource.Name, "grafana", grafana.Name)
		}

		// then import the dashboard into the matching grafana instances
		err = r.onDatasourceCreated(ctx, &grafana, datasource)
		if err != nil {
			success = false
			controllerLog.Error(err, "error reconciling dashboard", "datasource", datasource.Name, "grafana", grafana.Name)
		}
	}

	// if the datasource was successfully synced in all instances, wait for its re-sync period
	if success {
		return ctrl.Result{RequeueAfter: datasource.GetResyncPeriod()}, nil
	}

	return ctrl.Result{RequeueAfter: RequeueDelayError}, nil
}

func (r *GrafanaDatasourceReconciler) onDatasourceDeleted(ctx context.Context, namespace string, name string) error {
	list := v1beta1.GrafanaList{}
	opts := []client.ListOption{}
	err := r.Client.List(ctx, &list, opts...)
	if err != nil {
		return err
	}

	for _, grafana := range list.Items {
		if found, uid := grafana.Status.Datasources.Find(namespace, name); found {
			grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			datasource, err := grafanaClient.DataSourceByUID(*uid)
			if err != nil {
				return err
			}

			err = grafanaClient.DeleteDataSource(datasource.ID)
			if err != nil {
				if !strings.Contains(err.Error(), "status: 404") {
					return err
				}
			}

			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-datasource", name))
			if err != nil {
				return err
			}

			grafana.Status.Datasources = grafana.Status.Datasources.Remove(namespace, name)
			return r.Client.Status().Update(ctx, &grafana)
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) onDatasourceCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource) error {
	if cr.Spec.Datasource == nil {
		return nil
	}

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	id, err := r.ExistingId(grafanaClient, cr)
	if err != nil {
		return err
	}

	// always use the same uid for CR and datasource
	cr.Spec.Datasource.UID = string(cr.UID)
	datasourceBytes, err := json.Marshal(cr.Spec.Datasource)
	if err != nil {
		return err
	}

	if id == nil {
		_, err = grafanaClient.NewDataSourceFromRawData(datasourceBytes)
		// already exists error?
		if err != nil && !strings.Contains(err.Error(), "status: 409") {
			return err
		}
	} else if !cr.Unchanged() {
		err := grafanaClient.UpdateDataSourceFromRawData(*id, datasourceBytes)
		if err != nil {
			return err
		}
	} else {
		// datasource exists and is unchanged, nothing to do
		return nil
	}

	err = r.UpdateStatus(ctx, cr)
	if err != nil {
		return err
	}

	grafana.Status.Datasources = grafana.Status.Datasources.Add(cr.Namespace, cr.Name, string(cr.UID))
	return r.Client.Status().Update(ctx, grafana)
}

func (r *GrafanaDatasourceReconciler) UpdateStatus(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	cr.Status.Hash = cr.Hash()
	return r.Client.Status().Update(ctx, cr)
}

func (r *GrafanaDatasourceReconciler) ExistingId(client *gapi.Client, cr *v1beta1.GrafanaDatasource) (*int64, error) {
	datasources, err := client.DataSources()
	if err != nil {
		return nil, err
	}
	for _, datasource := range datasources {
		if datasource.UID == string(cr.UID) {
			return &datasource.ID, nil
		}
	}
	return nil, nil
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
