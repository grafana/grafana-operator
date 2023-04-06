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
	"fmt"
	"reflect"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	client2 "github.com/grafana-operator/grafana-operator/v5/controllers/client"
	gapi "github.com/grafana/grafana-api-golang-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

// GrafanaDatasourceReconciler reconciles a GrafanaDatasource object
type GrafanaDatasourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

const datasourceFinalizer = "datasource.grafana.integreatly.org/finalizer"

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/finalizers,verbs=update

func (r *GrafanaDatasourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("datasource", req.NamespacedName)

	datasource := &v1beta1.GrafanaDatasource{}
	if err := r.Get(ctx, req.NamespacedName, datasource); err != nil {
		log.Error(err, "unable to fetch Datasource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if datasource.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(datasource, datasourceFinalizer) {
			controllerutil.AddFinalizer(datasource, datasourceFinalizer)
			if err := r.Update(ctx, datasource); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(datasource, datasourceFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, datasource); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(datasource, datasourceFinalizer)
			if err := r.Update(ctx, datasource); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	content, err := r.getDatasourceContent(ctx, datasource)
	if err != nil {
		r.setReadyCondition(ctx, datasource, metav1.ConditionFalse, v1beta1.ContentUnavailableReason, fmt.Sprintf("failed to get datasource content: %s", err))
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	instances, err := getMatchingInstances(ctx, r.Client, datasource.Spec.InstanceSelector)
	if err != nil {
		log.Error(err, "could not find matching instances")
		r.setReadyCondition(ctx, datasource, metav1.ConditionFalse, v1beta1.NoMatchingInstancesReason, err.Error())
		return ctrl.Result{}, err
	}

	newInstanceStatuses := map[string]v1beta1.GrafanaDatasourceInstanceStatus{}
	for _, grafana := range instances.Items {
		grafana := &grafana
		log := log.WithValues("grafana", client.ObjectKeyFromObject(grafana))

		// check if this is a cross namespace import
		if grafana.Namespace != datasource.Namespace && !datasource.IsAllowCrossNamespaceImport() {
			continue
		}

		if !grafana.Ready() {
			log.V(1).Info("skipping grafana instance that is not ready")
			continue
		}

		if grafana.IsInternal() {
			err = updateGrafanaStatusPlugins(ctx, r.Client, grafana, datasource.Spec.Plugins)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to reconcile plugins: %w", err)
			}
		} else if datasource.Spec.Plugins != nil {
			log.Error(nil, "plugin availability not ensured for external grafana instance")
		}

		instanceStatus, err := r.syncDatasourceContent(ctx, grafana, datasource, content)
		if err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("error reconciling datasource: %w", err)
		}
		newInstanceStatuses[v1beta1.InstanceKeyFor(grafana)] = *instanceStatus
	}

	if !reflect.DeepEqual(datasource.Status.Instances, newInstanceStatuses) {
		datasource.Status.Instances = newInstanceStatuses
		if err := r.Client.Status().Update(ctx, datasource); err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to update status for datasource instance: %w", err)
		}
	}

	r.setReadyCondition(ctx, datasource, metav1.ConditionTrue, v1beta1.DatasourceSyncedReason, "Datasource synced")

	return ctrl.Result{RequeueAfter: datasource.GetResyncPeriod()}, nil
}

func (r *GrafanaDatasourceReconciler) deleteExternalResources(ctx context.Context, datasource *v1beta1.GrafanaDatasource) error {
	for grafanaKey, instanceStatus := range datasource.Status.Instances {
		var grafana v1beta1.Grafana
		err := r.Client.Get(ctx, v1beta1.NamespacedNameFor(grafanaKey), &grafana)
		if err != nil {
			return err
		}

		grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		datasource, err := grafanaClient.DataSourceByUID(instanceStatus.UID)
		if err != nil {
			return err
		}

		err = grafanaClient.DeleteDataSource(datasource.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) getDatasourceContent(ctx context.Context, cr *v1beta1.GrafanaDatasource) (*gapi.DataSource, error) {
	variables, err := r.collectVariablesFromSecrets(ctx, cr)
	if err != nil {
		return nil, err
	}

	return cr.ExpandVariables(variables)
}

func (r *GrafanaDatasourceReconciler) syncDatasourceContent(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource, datasource *gapi.DataSource) (*v1beta1.GrafanaDatasourceInstanceStatus, error) {
	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return nil, fmt.Errorf("failed to create grafana client: %w", err)
	}

	existing, err := r.getExistingDatasource(grafanaClient, grafana, cr)
	if err != nil {
		return nil, fmt.Errorf("failed to getexisting datasource: %w", err)
	}

	var instanceStatus v1beta1.GrafanaDatasourceInstanceStatus
	if existing != nil {
		datasource.ID = existing.ID
		datasource.UID = existing.UID

		if !reflect.DeepEqual(*existing, *datasource) {
			err := grafanaClient.UpdateDataSource(datasource)
			if err != nil {
				return nil, fmt.Errorf("failed to update datasource: %w", err)
			}
		}

		instanceStatus = v1beta1.GrafanaDatasourceInstanceStatus{
			ID:  existing.ID,
			UID: existing.UID,
		}
	} else {
		id, err := grafanaClient.NewDataSource(datasource)
		if err != nil && !strings.Contains(err.Error(), "status: 409") {
			return nil, fmt.Errorf("failed to create datasource: %w", err)
		}

		uid := cr.Spec.DataSource.UID
		if uid == "" {
			ds, err := grafanaClient.DataSource(id)
			if err != nil {
				return nil, fmt.Errorf("failed to get created datasource: %w", err)
			}
			uid = ds.UID
		}

		instanceStatus = v1beta1.GrafanaDatasourceInstanceStatus{
			ID:  id,
			UID: uid,
		}
	}

	return &instanceStatus, nil
}

func (r *GrafanaDatasourceReconciler) getExistingDatasource(client *gapi.Client, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource) (*gapi.DataSource, error) {
	if instanceStatus, ok := cr.Status.Instances[v1beta1.InstanceKeyFor(grafana)]; ok {
		return client.DataSourceByUID(instanceStatus.UID)
	}
	return nil, nil
}

func (r *GrafanaDatasourceReconciler) collectVariablesFromSecrets(ctx context.Context, cr *v1beta1.GrafanaDatasource) (map[string][]byte, error) {
	result := map[string][]byte{}
	for _, secret := range cr.Spec.Secrets {
		// secrets must be in the same namespace as the datasource
		selector := client.ObjectKey{
			Namespace: cr.Namespace,
			Name:      strings.TrimSpace(secret),
		}

		s := &v1.Secret{}
		err := r.Client.Get(ctx, selector, s)
		if err != nil {
			return nil, err
		}

		for key, value := range s.Data {
			result[key] = value
		}
	}
	return result, nil
}

func (r *GrafanaDatasourceReconciler) setReadyCondition(ctx context.Context, datasource *v1beta1.GrafanaDatasource, status metav1.ConditionStatus, reason string, message string) error {
	changed := datasource.SetReadyCondition(status, reason, message)
	if !changed {
		return nil
	}

	if err := r.Client.Status().Update(ctx, datasource); err != nil {
		r.Log.WithValues("datasource", client.ObjectKeyFromObject(datasource)).Error(err, "failed to update status")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDatasourceReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDatasource{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForDeployment),
			builder.WithPredicates(
				predicate.NewPredicateFuncs(grafanaOwnedResources),
				predicate.NewPredicateFuncs(deploymentReady),
			),
		).
		Complete(r)

	return err
}

func (r *GrafanaDatasourceReconciler) findObjectsForDeployment(o client.Object) []reconcile.Request {
	ctx := context.Background()
	gafanaRef := types.NamespacedName{Namespace: o.GetNamespace(), Name: strings.TrimSuffix(o.GetName(), "-grafana")}
	grafana := &v1beta1.Grafana{}
	err := r.Client.Get(ctx, gafanaRef, grafana)
	if err != nil {
		return []reconcile.Request{}
	}

	var datasources v1beta1.GrafanaDatasourceList
	err = r.Client.List(ctx, &datasources)
	if err != nil {
		return []reconcile.Request{}
	}

	reqs := []reconcile.Request{}
	for _, datasource := range datasources.Items {
		selector, err := metav1.LabelSelectorAsSelector(datasource.Spec.InstanceSelector)

		if err != nil {
			return []reconcile.Request{}
		}

		if selector.Matches(labels.Set(grafana.ObjectMeta.Labels)) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&datasource),
			})
		}
	}

	return reqs
}
