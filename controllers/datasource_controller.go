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

	"github.com/grafana-operator/grafana-operator/v5/controllers/model"

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

	instances, err := r.GetMatchingDatasourceInstances(ctx, datasource, r.Client)
	if err != nil {
		log.Error(err, "could not find matching instances", "name", datasource.Name, "namespace", datasource.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	log.Info("found matching Grafana instances for datasource", "count", len(instances.Items))

	success := true
	for _, grafana := range instances.Items {
		// check if this is a cross namespace import
		if grafana.Namespace != datasource.Namespace && !datasource.IsAllowCrossNamespaceImport() {
			continue
		}

		grafana := grafana
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		if !grafana.Ready() {
			log.Info("grafana instance not ready", "grafana", grafana.Name)
			// TODO: set condition
			success = false
			continue
		}

		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested datasources to a configmap from where the
			// grafana reconciler will pick them upi
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, datasource.Spec.Plugins, fmt.Sprintf("%v-datasource", datasource.Name))
			if err != nil {
				success = false
				log.Error(err, "error reconciling plugins", "grafana", grafana.Name)
			}

			deploy := model.GetGrafanaDeployment(&grafana, r.Scheme)
			err := r.Client.Get(ctx, client.ObjectKeyFromObject(deploy), deploy)
			if err != nil || deploy.Status.ReadyReplicas == 0 {
				return ctrl.Result{}, nil
			}
		}

		// then import the datasource into the matching grafana instances
		err = r.onDatasourceCreated(ctx, &grafana, datasource)
		if err != nil {
			success = false
			// datasource.Status.LastMessage = err.Error()
			// todo: set condition
			log.Error(err, "error reconciling datasource", "grafana", grafana.Name)
		}
	}

	// if the datasource was successfully synced in all instances, wait for its re-sync period
	if success {
		// datasource.Status.LastMessage = ""
		// todo: set condition
		return ctrl.Result{RequeueAfter: datasource.GetResyncPeriod()}, r.UpdateStatus(ctx, datasource)
	} else {
		// if there was an issue with the datasource, update the status
		return ctrl.Result{RequeueAfter: RequeueDelay}, r.UpdateStatus(ctx, datasource)
	}
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

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-datasource", datasource.Name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) onDatasourceCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource) error {
	if grafana.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your datasource cr")
	}

	variables, err := r.CollectVariablesFromSecrets(ctx, cr)
	if err != nil {
		return err
	}

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	existing, err := r.Existing(grafanaClient, grafana, cr)
	if err != nil {
		return err
	}

	datasource, err := cr.ExpandVariables(variables)
	if err != nil {
		return err
	}

	if existing != nil {
		datasource.ID = existing.ID
		datasource.UID = existing.UID
		if !reflect.DeepEqual(*existing, *datasource) {
			err := grafanaClient.UpdateDataSource(datasource)
			if err != nil {
				return err
			}
		}
	} else {
		id, err := grafanaClient.NewDataSource(datasource)
		if err != nil && !strings.Contains(err.Error(), "status: 409") {
			return err
		}

		uid := cr.Spec.DataSource.UID
		if uid == "" {
			ds, err := grafanaClient.DataSource(id)
			if err != nil {
				return err
			}
			uid = ds.UID
		}

		cr.Status.Instances[v1beta1.InstanceKeyFor(grafana)] = v1beta1.GrafanaDatasourceInstanceStatus{
			ID:  id,
			UID: uid,
		}
	}

	return r.UpdateStatus(ctx, cr)
}

func (r *GrafanaDatasourceReconciler) UpdateStatus(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	return r.Client.Status().Update(ctx, cr)
}

func (r *GrafanaDatasourceReconciler) Existing(client *gapi.Client, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource) (*gapi.DataSource, error) {
	if instanceStatus, ok := cr.Status.Instances[v1beta1.InstanceKeyFor(grafana)]; ok {
		return client.DataSourceByUID(instanceStatus.UID)
	}
	return nil, nil
}

func (r *GrafanaDatasourceReconciler) CollectVariablesFromSecrets(ctx context.Context, cr *v1beta1.GrafanaDatasource) (map[string][]byte, error) {
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

func (r *GrafanaDatasourceReconciler) GetMatchingDatasourceInstances(ctx context.Context, datasource *v1beta1.GrafanaDatasource, k8sClient client.Client) (v1beta1.GrafanaList, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, datasource.Spec.InstanceSelector)
	if err != nil || len(instances.Items) == 0 {
		r.SetReadyCondition(ctx, datasource, metav1.ConditionFalse, v1beta1.NoMatchingInstancesReason, "No instances found matching .spec.instanceSelector")
		return v1beta1.GrafanaList{}, err
	}

	return instances, err
}

func (r *GrafanaDatasourceReconciler) SetReadyCondition(ctx context.Context, datasource *v1beta1.GrafanaDatasource, status metav1.ConditionStatus, reason string, message string) error {
	existingCond := datasource.GetReadyCondition()
	if existingCond != nil && existingCond.Status == status && existingCond.Reason == reason && existingCond.Message == message {
		return nil
	}

	datasource.SetReadyCondition(status, reason, message)
	if err := r.Client.Status().Update(ctx, datasource); err != nil {
		r.Log.Info("unable to update the status of %v, in %v", datasource.Name, datasource.Namespace)
		return err
	}
	return nil
}

func (r *GrafanaDatasourceReconciler) findObjectsForDeployment(o client.Object) []reconcile.Request {
	ctx := context.Background()
	gafanaRef := types.NamespacedName{Namespace: o.GetNamespace(), Name: strings.TrimSuffix(o.GetName(), "-deployment")}
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
