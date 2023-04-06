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
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	grapi "github.com/grafana/grafana-api-golang-client"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana-operator/grafana-operator/v5/controllers/client"
	"github.com/grafana-operator/grafana-operator/v5/controllers/metrics"
	"github.com/grafana-operator/grafana-operator/v5/controllers/util"
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Discovery discovery.DiscoveryInterface
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update

func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dashboard", req.NamespacedName)

	dashboard := &v1beta1.GrafanaDashboard{}
	if err := r.Get(ctx, req.NamespacedName, dashboard); err != nil {
		log.Error(err, "unable to fetch Dashboard")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if dashboard.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(dashboard, datasourceFinalizer) {
			controllerutil.AddFinalizer(dashboard, datasourceFinalizer)
			if err := r.Update(ctx, dashboard); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(dashboard, datasourceFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, dashboard); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(dashboard, datasourceFinalizer)
			if err := r.Update(ctx, dashboard); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// TODO: get content once here subreconcile step 1

	manifest, err := r.getDashboardManifest(ctx, dashboard)
	if err != nil {
		r.setReadyCondition(ctx, dashboard, metav1.ConditionFalse, v1beta1.ContentUnavailableReason, fmt.Sprintf("failed to get dashboard: %s", err))
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	instances, err := getMatchingInstances(ctx, r.Client, dashboard.Spec.InstanceSelector)
	if err != nil {
		log.Error(err, "could not find matching instances")
		r.setReadyCondition(ctx, dashboard, metav1.ConditionFalse, v1beta1.NoMatchingInstancesReason, err.Error())
		return ctrl.Result{}, err
	}

	newInstanceStatuses := map[string]v1beta1.GrafanaDashboardInstanceStatus{}
	for _, grafana := range instances.Items {
		grafana := &grafana
		log := log.WithValues("grafana", client.ObjectKeyFromObject(grafana))

		// check if this is a cross namespace import
		if grafana.Namespace != dashboard.Namespace && !dashboard.IsAllowCrossNamespaceImport() {
			continue
		}

		if !grafana.Ready() {
			log.V(1).Info("skipping grafana instance that is not ready")
			continue
		}

		if grafana.IsInternal() {
			err = updateGrafanaStatusPlugins(ctx, r.Client, grafana, dashboard.Spec.Plugins)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to reconcile plugins: %w", err)
			}
		} else if dashboard.Spec.Plugins != nil {
			log.Error(nil, "plugin availability not ensured for external grafana instance")
		}

		instanceStatus, err := r.syncDashboardContent(ctx, grafana, dashboard, manifest)
		if err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("error reconciling dashboard: %w", err)
		}
		newInstanceStatuses[v1beta1.InstanceKeyFor(grafana)] = *instanceStatus
	}

	if !reflect.DeepEqual(dashboard.Status.Instances, newInstanceStatuses) {
		dashboard.Status.Instances = newInstanceStatuses
		if err := r.Client.Status().Update(ctx, dashboard); err != nil {
			return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to update status for dashboard instance: %w", err)
		}
	}

	r.setReadyCondition(ctx, dashboard, metav1.ConditionTrue, v1beta1.DashboardSyncedReason, "Dashboard synced")

	return ctrl.Result{RequeueAfter: dashboard.GetResyncPeriod()}, nil

}

func (r *GrafanaDashboardReconciler) deleteExternalResources(ctx context.Context, dashboard *v1beta1.GrafanaDashboard) error {
	for grafanaKey, instanceStatus := range dashboard.Status.Instances {
		var grafana v1beta1.Grafana
		err := r.Client.Get(ctx, v1beta1.NamespacedNameFor(grafanaKey), &grafana)
		if err != nil {
			return err
		}

		grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		err = grafanaClient.DeleteDashboardByUID(instanceStatus.UID)
		if err != nil {
			if !strings.Contains(err.Error(), "status: 404") {
				return err
			}
		}
	}

	return nil
}

func (r *GrafanaDashboardReconciler) syncDashboardContent(ctx context.Context, grafana *v1beta1.Grafana, dashboard *v1beta1.GrafanaDashboard, manifest map[string]interface{}) (*v1beta1.GrafanaDashboardInstanceStatus, error) {
	instanceKey := v1beta1.InstanceKeyFor(grafana)

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return nil, fmt.Errorf("failed to create grafana client: %w", err)
	}

	folder, err := r.getOrCreateFolder(grafanaClient, dashboard)
	if err != nil {
		return nil, fmt.Errorf("failed get or create folder: %w", err)
	}
	folderId := int64(0)
	if folder != nil {
		folderId = folder.ID
	}

	shouldCreate := true
	instanceStatus, ok := dashboard.Status.Instances[instanceKey]
	if ok {
		existingMatches, err := r.existingVersionMatchesStatus(grafanaClient, instanceStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing dashboard: %w", err)
		}
		shouldCreate = !existingMatches
	} else {
		if dashboard.Status.Instances == nil {
			dashboard.Status.Instances = make(map[string]v1beta1.GrafanaDashboardInstanceStatus)
		}
		instanceStatus = v1beta1.GrafanaDashboardInstanceStatus{
			UID:     manifest["uid"].(string),
			Version: -1, // todo
		}
		dashboard.Status.Instances[instanceKey] = instanceStatus
	}
	if !shouldCreate {
		return &instanceStatus, nil
	}

	resp, err := grafanaClient.NewDashboard(grapi.Dashboard{
		Meta: grapi.DashboardMeta{
			IsStarred: false,
			Slug:      dashboard.Name,
			Folder:    folderId,
		},
		Model:     manifest,
		Folder:    folderId,
		Overwrite: true,
		Message:   "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new dashboard: %w", err)
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("unsuccessful status when creating dashboard: %s", resp.Status)
	}

	return &v1beta1.GrafanaDashboardInstanceStatus{
		UID:     resp.UID,
		Version: resp.Version,
	}, nil
}

func (r *GrafanaDashboardReconciler) getDashboardManifest(ctx context.Context, dashboard *v1beta1.GrafanaDashboard) (map[string]interface{}, error) {

	var manifestBytes []byte
	var err error
	if dashboard.Spec.Source.Inline != nil {
		manifestBytes, err = r.getInlineDashboardManifest(dashboard, dashboard.Spec.Source.Inline)
	} else if dashboard.Spec.Source.ConfigMap != nil {
		manifestBytes, err = r.getConfigMapDashboardManifest(ctx, dashboard, dashboard.Spec.Source.ConfigMap)
	} else if dashboard.Spec.Source.Remote != nil {
		manifestBytes, err = r.getRemoteDashboardManifest(ctx, dashboard, dashboard.Spec.Source.Remote)
	} else {
		return nil, fmt.Errorf("missing source for dashboard %s/%s", dashboard.Namespace, dashboard.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch dashboard content: %w", err)
	}

	var manifest map[string]interface{}
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		r.setReadyCondition(ctx, dashboard, metav1.ConditionFalse, v1beta1.ContentUnavailableReason, fmt.Sprintf("failed to unmarshal dashboard: %s", err))
		return nil, err
	}

	if _, ok := manifest["uid"]; !ok {
		manifest["uid"] = string(dashboard.UID)
	}

	return manifest, nil
}

func (r *GrafanaDashboardReconciler) getInlineDashboardManifest(dashboard *v1beta1.GrafanaDashboard, source *v1beta1.GrafanaDashboardInlineSource) ([]byte, error) {
	if source.Json != nil {
		return []byte(*source.Json), nil
	} else if source.GzipJson != nil {
		return v1beta1.Gunzip(source.GzipJson)
	} else if source.Jsonnet != nil {
		return util.FetchJsonnet(*source.Jsonnet)
	} else {
		return nil, fmt.Errorf("missing inline source for dashboard %s/%s", dashboard.Namespace, dashboard.Name)
	}
}

func (r *GrafanaDashboardReconciler) getConfigMapDashboardManifest(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, source *v1.ConfigMapKeySelector) ([]byte, error) {
	var cm v1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{
		Namespace: dashboard.Namespace,
		Name:      source.Name,
	}, &cm)
	if err != nil {
		return nil, err
	}
	if data, ok := cm.Data[source.Key]; ok {
		return []byte(data), nil
	} else {
		return nil, fmt.Errorf("missing key %s in configmap %s", source.Key, source.Name)
	}
}

func (r *GrafanaDashboardReconciler) getRemoteDashboardManifest(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, source *v1beta1.GrafanaDashboardRemoteSource) ([]byte, error) {
	var url string
	if source.Url != nil {
		url = *source.Url
	} else if source.GrafanaCom != nil {
		if source.GrafanaCom.Revision == nil {
			var err error
			source.GrafanaCom.Revision, err = util.GetLatestGrafanaComRevision(ctx, source.GrafanaCom.Id)
			if err != nil {
				return nil, err
			}
		}
		url = util.GetGrafanaComDashboardUrl(*source.GrafanaCom)
	}

	cache := dashboard.GetContentCache(url)
	if len(cache) > 0 {
		return cache, nil
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.DashboardUrlRequests)
	response, err := client.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from dashboard url request, get %v for dashboard %v", response.StatusCode, dashboard.Name)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	gz, err := v1beta1.Gzip(content)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to gzip dashboard %v", dashboard.Name)
	}

	dashboard.Status.Content = &v1beta1.GrafanaDashboardStatusContent{
		Cache:     gz,
		Timestamp: metav1.Now(),
		Url:       url,
	}

	if err := r.Client.Status().Update(ctx, dashboard); err != nil {
		r.Log.Error(err, "failed to set content cache", dashboard.Name, dashboard.Namespace)
	}

	return content, nil
}

func (r *GrafanaDashboardReconciler) existingVersionMatchesStatus(client *grapi.Client, instanceStatus v1beta1.GrafanaDashboardInstanceStatus) (bool, error) {
	existing, err := client.DashboardByUID(instanceStatus.UID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	if float64(instanceStatus.Version) == existing.Model["version"].(float64) {
		return true, nil
	}

	return false, nil
}

func (r *GrafanaDashboardReconciler) getOrCreateFolder(client *grapi.Client, cr *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
	if cr.Spec.FolderTitle == "" {
		return nil, nil
	}

	folder, err := r.getFolder(client, cr)
	if err != nil {
		return nil, err
	}
	if folder != nil {
		return folder, nil
	}

	// Folder wasn't found, let's create it
	resp, err := client.NewFolder(cr.Spec.FolderTitle)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *GrafanaDashboardReconciler) getFolder(client *grapi.Client, cr *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
	folders, err := client.Folders()
	if err != nil {
		return nil, err
	}

	for _, folder := range folders {
		if folder.Title == cr.Spec.FolderTitle {
			return &folder, nil
		}
		continue
	}
	return nil, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDashboard{}).
		Watches(
			&source.Kind{Type: &v1beta1.Grafana{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForGrafana),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&source.Kind{Type: &v1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
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
func (r *GrafanaDashboardReconciler) setReadyCondition(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, status metav1.ConditionStatus, reason string, message string) error {
	changed := dashboard.SetReadyCondition(status, reason, message)
	if !changed {
		return nil
	}

	if err := r.Client.Status().Update(ctx, dashboard); err != nil {
		r.Log.WithValues("dashboard", client.ObjectKeyFromObject(dashboard)).Error(err, "failed to update status")
		return err
	}
	return nil
}

func (r *GrafanaDashboardReconciler) findObjectsForConfigMap(configMap client.Object) []reconcile.Request {
	configMapDashboards := &v1beta1.GrafanaDashboardList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(".spec.source.configMap.name", configMap.GetName()),
		Namespace:     configMap.GetNamespace(),
	}
	err := r.Client.List(context.TODO(), configMapDashboards, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(configMapDashboards.Items))
	for i, item := range configMapDashboards.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *GrafanaDashboardReconciler) findObjectsForDeployment(o client.Object) []reconcile.Request {
	owner := getGrafanaOwner(o.GetOwnerReferences())
	ctx := context.Background()
	var grafana v1beta1.Grafana
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: o.GetNamespace(), Name: owner.Name}, &grafana)
	if err != nil {
		return []reconcile.Request{}
	}

	return r.findObjectsForGrafanaLabels(ctx, grafana.GetLabels())
}

func (r *GrafanaDashboardReconciler) findObjectsForGrafana(o client.Object) []reconcile.Request {
	return r.findObjectsForGrafanaLabels(context.Background(), o.GetLabels())
}

func (r *GrafanaDashboardReconciler) findObjectsForGrafanaLabels(ctx context.Context, grafanaLabels map[string]string) []reconcile.Request {
	var dashboards v1beta1.GrafanaDashboardList
	err := r.Client.List(ctx, &dashboards)
	if err != nil {
		return []reconcile.Request{}
	}

	reqs := []reconcile.Request{}
	for _, dashboard := range dashboards.Items {
		selector, err := metav1.LabelSelectorAsSelector(dashboard.Spec.InstanceSelector)

		if err != nil {
			return []reconcile.Request{}
		}

		if selector.Matches(labels.Set(grafanaLabels)) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&dashboard),
			})
		}
	}

	return reqs
}
