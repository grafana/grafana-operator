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

	simplejson "github.com/bitly/go-simplejson"
	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionDatasourceSynchronized = "DatasourceSynchronized"
)

// GrafanaDatasourceReconciler reconciles a GrafanaDatasource object
type GrafanaDatasourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/finalizers,verbs=update

func (r *GrafanaDatasourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaDatasourceReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaDatasource{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("error getting grafana datasource cr: %w", err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaDatasource: %w", err)
			}
			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, cr)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	cr.Status.NoMatchingInstances = false
	log.Info("found matching Grafana instances for datasource", "count", len(instances))

	if cr.IsUpdatedUID() {
		log.Info("datasource uid got updated, deleting datasources with the old uid")
		if err = r.deleteOldDatasource(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}

		// Clean up uid, so further reconcilications can track changes there
		cr.Status.UID = ""

		// Force requeue for datasource creation
		return ctrl.Result{Requeue: true}, nil
	}

	datasource, hash, err := r.buildDatasourceModel(ctx, cr)
	if err != nil {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidModel", err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)
		return ctrl.Result{}, fmt.Errorf("could not build datasource model: %w", err)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	pluginErrors := make(map[string]string)
	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested datasources to a configmap from where the
			// grafana reconciler will pick them upi
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, cr.Spec.Plugins, fmt.Sprintf("%v-datasource", cr.Name))
			if err != nil {
				pluginErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			}
		}

		// then import the datasource into the matching grafana instances
		err = r.onDatasourceCreated(ctx, &grafana, cr, datasource, hash)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	if len(pluginErrors) > 0 {
		err := fmt.Errorf("%v", pluginErrors)
		log.Error(err, "failed to apply plugins to all instances")
	}

	allApplyErrors := mergeReconcileErrors(applyErrors, pluginErrors)

	condition := buildSynchronizedCondition("Datasource", conditionDatasourceSynchronized, cr.Generation, allApplyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(allApplyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", allApplyErrors)
	}

	cr.Status.Hash = hash
	cr.Status.LastMessage = "" // nolint:staticcheck
	cr.Status.UID = cr.CustomUIDOrUID()

	return ctrl.Result{RequeueAfter: cr.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaDatasourceReconciler) deleteOldDatasource(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		found, uid := grafana.Status.Datasources.Find(cr.Namespace, cr.Name)
		if !found {
			continue
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		datasource, err := grafanaClient.Datasources.GetDataSourceByUID(*uid)
		var notFound *datasources.GetDataSourceByUIDNotFound
		if errors.As(err, &notFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("fetching datasource: %w", err)
		}

		_, err = grafanaClient.Datasources.DeleteDataSourceByUID(datasource.Payload.UID) //nolint
		if err != nil {
			return fmt.Errorf("deleting datasource to update uid %s: %w", *uid, err)
		}

		// Update grafana instance Status
		err = removeNamespacedResource(ctx, r.Client, &grafana, cr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		found, uid := grafana.Status.Datasources.Find(cr.Namespace, cr.Name)
		if !found {
			continue
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		_, err = grafanaClient.Datasources.DeleteDataSourceByUID(*uid) // nolint:errcheck
		var notFound *datasources.DeleteDataSourceByUIDNotFound
		if errors.As(err, &notFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("deleting datasource %s: %w", *uid, err)
		}

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-datasource", cr.Name))
			if err != nil {
				return fmt.Errorf("reconciling plugins: %w", err)
			}
		}

		// Update grafana instance Status
		err = removeNamespacedResource(ctx, r.Client, &grafana, cr)
		if err != nil {
			return err
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

	// Update grafana instance Status
	return addNamespacedResource(ctx, r.Client, grafana, cr, cr.NamespacedResource())
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
func (r *GrafanaDatasourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDatasource{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

func (r *GrafanaDatasourceReconciler) buildDatasourceModel(ctx context.Context, cr *v1beta1.GrafanaDatasource) (*models.UpdateDataSourceCommand, string, error) {
	log := logf.FromContext(ctx)
	// Overwrite OrgID to ensure the field is useless
	cr.Spec.Datasource.OrgID = nil

	initialBytes, err := json.Marshal(cr.Spec.Datasource)
	if err != nil {
		return nil, "", fmt.Errorf("encoding existing datasource model as json: %w", err)
	}

	// Unstructured object for mutating target paths
	simpleContent, err := simplejson.NewJson(initialBytes)
	if err != nil {
		return nil, "", fmt.Errorf("parsing marshaled json as simplejson: %w", err)
	}

	simpleContent.Set("uid", cr.CustomUIDOrUID())

	for _, override := range cr.Spec.ValuesFrom {
		val, key, err := getReferencedValue(ctx, r.Client, cr, override.ValueFrom)
		if err != nil {
			return nil, "", fmt.Errorf("getting referenced value: %w", err)
		}

		patternToReplace := simpleContent.GetPath(strings.Split(override.TargetPath, ".")...)
		patternString, err := patternToReplace.String()
		if err != nil {
			return nil, "", fmt.Errorf("pattern must be a string")
		}

		patternString = strings.ReplaceAll(patternString, fmt.Sprintf("${%v}", key), val)
		patternString = strings.ReplaceAll(patternString, fmt.Sprintf("$%v", key), val)

		log.V(1).Info("overriding value", "key", override.TargetPath, "value", val)
		simpleContent.SetPath(strings.Split(override.TargetPath, "."), patternString)
	}

	newBytes, err := simpleContent.MarshalJSON()
	if err != nil {
		return nil, "", fmt.Errorf("encoding expanded datasource model as json: %w", err)
	}

	// TODO models.DataSource has SecureJsonData field now, verify if below is still true
	// We use UpdateDataSourceCommand here because models.DataSource lacks the SecureJsonData field
	var res models.UpdateDataSourceCommand
	if err = json.Unmarshal(newBytes, &res); err != nil {
		return nil, "", fmt.Errorf("deserializing expanded datasource model from json: %w", err)
	}

	// TODO Remove hashing along with the Status.Hash field
	hash := sha256.New()
	hash.Write(newBytes)

	return &res, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
