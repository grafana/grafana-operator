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
	"reflect"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	client2 "github.com/grafana-operator/grafana-operator/v5/controllers/client"
	"github.com/grafana-operator/grafana-operator/v5/controllers/client/github.com/grafana/grafana/pkg/components/simplejson"
	gapi "github.com/grafana/grafana-api-golang-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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
	Scheme        *runtime.Scheme
	Log           logr.Logger
	EventRecorder record.EventRecorder
}

const (
	datasourceFinalizer              = "datasource.grafana.integreatly.org/finalizer"
	datasourceSecretsRefIndexField   = ".spec.valuesFrom.secretKeyRef"
	datasourceConfigMapRefIndexField = ".spec.valuesFrom.configMapKeyRef"
)

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadatasources/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

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

	nextDatasource := datasource.DeepCopy()

	content, err := r.getDatasourceContent(ctx, nextDatasource)
	if err != nil {
		err = fmt.Errorf("failed to get datasource content: %s", err)
		nextDatasource.SetReadyCondition(metav1.ConditionFalse, v1beta1.ContentUnavailableReason, err.Error())
		r.EventRecorder.Event(datasource, v1.EventTypeWarning, "DatasourceSyncFailed", "failed to get datasource content")
		return r.reconcileResult(ctx, datasource, nextDatasource, &errorRequeueDelay, err)
	}

	instances, err := getMatchingInstances(ctx, r.Client, datasource.Spec.InstanceSelector)
	if err != nil {
		log.Error(err, "could not find matching instances")
		nextDatasource.SetReadyCondition(metav1.ConditionFalse, v1beta1.NoMatchingInstancesReason, err.Error())
		return r.reconcileResult(ctx, datasource, nextDatasource, nil, err)
	}

	if nextDatasource.Status.Instances == nil {
		nextDatasource.Status.Instances = map[string]v1beta1.GrafanaDatasourceInstanceStatus{}
	}

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
				err = fmt.Errorf("failed to reconcile plugins: %w", err)
				nextDatasource.SetReadyCondition(metav1.ConditionFalse, v1beta1.CreateResourceFailedReason, err.Error())
				return r.reconcileResult(ctx, datasource, nextDatasource, &errorRequeueDelay, err)
			}
		} else if datasource.Spec.Plugins != nil {
			log.Error(nil, "plugin availability not ensured for external grafana instance")
		}

		instanceStatus, err := r.syncDatasourceContent(ctx, grafana, nextDatasource, content)
		if err != nil {
			err = fmt.Errorf("error reconciling datasource: %w", err)
			nextDatasource.SetReadyCondition(metav1.ConditionFalse, v1beta1.CreateResourceFailedReason, err.Error())
			return r.reconcileResult(ctx, datasource, nextDatasource, &errorRequeueDelay, err)
		}
		nextDatasource.Status.Instances[v1beta1.InstanceKeyFor(grafana)] = *instanceStatus
		r.EventRecorder.Eventf(datasource, v1.EventTypeNormal, v1beta1.DatasourceSyncedReason, "Synced datasource with grafana instance %s/%s", grafana.Namespace, grafana.Name)
	}

	nextDatasource.SetReadyCondition(metav1.ConditionTrue, v1beta1.DatasourceSyncedReason, "Datasource synced")

	requeueDelay := datasource.GetResyncPeriod()
	return r.reconcileResult(ctx, datasource, nextDatasource, &requeueDelay, nil)
}

func (r *GrafanaDatasourceReconciler) reconcileResult(ctx context.Context, datasource *v1beta1.GrafanaDatasource, nextDatasource *v1beta1.GrafanaDatasource, resyncDelay *time.Duration, err error) (reconcile.Result, error) {
	if !reflect.DeepEqual(datasource.Status, nextDatasource.Status) {
		err := r.Client.Status().Update(context.Background(), nextDatasource)
		if err != nil {
			return ctrl.Result{RequeueAfter: errorRequeueDelay}, fmt.Errorf("failed to update status for datasource %s", datasource.Name)
		}
	}

	if resyncDelay == nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: *resyncDelay}, err
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
	initialBytes, err := json.Marshal(cr.Spec.DataSource)
	if err != nil {
		return nil, err
	}

	simpleContent, err := simplejson.NewJson(initialBytes)
	for _, ref := range cr.Spec.ValuesFrom {
		val, err := r.getReferencedValue(ctx, cr, &ref.ValueFrom)
		if err != nil {
			return nil, err
		}
		simpleContent.SetPath(strings.Split(ref.TargetPath, "."), val)
	}

	newBytes, err := simpleContent.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var res gapi.DataSource

	err = json.Unmarshal(newBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
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
		existing, err := client.DataSourceByUID(instanceStatus.UID)
		if err != nil && !strings.Contains(err.Error(), "404") {
			return nil, err
		}
		return existing, nil
	}
	return nil, nil
}

func (r *GrafanaDatasourceReconciler) getReferencedValue(ctx context.Context, cr *v1beta1.GrafanaDatasource, source *v1beta1.GrafanaDatasourceValueFromSource) (string, error) {
	if source.SecretKeyRef != nil {
		s := &v1.Secret{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", err
		}
		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return string(val), nil
		} else {
			return "", fmt.Errorf("missing key %s in secret %s", source.SecretKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	} else {
		s := &v1.ConfigMap{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", err
		}
		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return val, nil
		} else {
			return "", fmt.Errorf("missing key %s in configmap %s", source.SecretKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDatasourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.addValueSourceIndexField(mgr, datasourceSecretsRefIndexField, func(source v1beta1.GrafanaDatasourceValueFromSource) string {
		if source.SecretKeyRef != nil {
			return source.SecretKeyRef.Name
		}
		return ""
	}); err != nil {
		return err
	}

	if err := r.addValueSourceIndexField(mgr, datasourceConfigMapRefIndexField, func(source v1beta1.GrafanaDatasourceValueFromSource) string {
		if source.ConfigMapKeyRef != nil {
			return source.ConfigMapKeyRef.Name
		}
		return ""
	}); err != nil {
		return err
	}

	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDatasource{}).
		Watches(
			&source.Kind{Type: &v1beta1.Grafana{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForGrafana),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}), // api availability change should trigger dashboard reconcile
		).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForIndexField(datasourceSecretsRefIndexField)),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Watches(
			&source.Kind{Type: &v1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForIndexField(datasourceConfigMapRefIndexField)),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)

	return err
}

func (r *GrafanaDatasourceReconciler) findObjectsForGrafana(grafana client.Object) []reconcile.Request {
	grafanaLabels := grafana.GetLabels()
	ctx := context.Background()
	var datasources v1beta1.GrafanaDatasourceList
	err := r.Client.List(ctx, &datasources)
	if err != nil {
		return []reconcile.Request{}
	}

	reqs := []reconcile.Request{}
	for _, datasource := range datasources.Items {
		selector, err := metav1.LabelSelectorAsSelector(datasource.Spec.InstanceSelector)
		if err != nil {
			return []reconcile.Request{}
		}

		if selector.Matches(labels.Set(grafanaLabels)) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&datasource),
			})
		}
	}

	return reqs
}

func (r *GrafanaDatasourceReconciler) findObjectsForIndexField(indexField string) func(secret client.Object) []reconcile.Request {
	return func(object client.Object) []reconcile.Request {
		secretDatasources := &v1beta1.GrafanaDatasourceList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(indexField, object.GetName()),
			Namespace:     object.GetNamespace(),
		}
		err := r.Client.List(context.Background(), secretDatasources, listOps)
		if err != nil {
			return []reconcile.Request{}
		}

		requests := make([]reconcile.Request, len(secretDatasources.Items))
		for i, item := range secretDatasources.Items {
			requests[i] = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			}
		}
		return requests
	}
}

func (r *GrafanaDatasourceReconciler) addValueSourceIndexField(mgr ctrl.Manager, indexField string, valueSourceName func(v1beta1.GrafanaDatasourceValueFromSource) string) error {
	return mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.GrafanaDatasource{}, indexField, func(rawObj client.Object) []string {
		datasource := rawObj.(*v1beta1.GrafanaDatasource)
		var res []string
		for _, v := range datasource.Spec.ValuesFrom {
			if name := valueSourceName(v.ValueFrom); name != "" {
				res = append(res, name)
			}
		}
		return res
	})
}
