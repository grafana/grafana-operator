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
	"time"

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

const (
	dashboardFinalizer      = "dashboard.grafana.integreatly.org/finalizer"
	dashboardConfigMapField = ".spec.source.configMap.name"
)

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dashboard", req.NamespacedName)

	dashboard := &v1beta1.GrafanaDashboard{}
	if err := r.Get(ctx, req.NamespacedName, dashboard); err != nil {
		log.Error(err, "ignoring deleted dashboard")
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
		if !controllerutil.ContainsFinalizer(dashboard, dashboardFinalizer) {
			controllerutil.AddFinalizer(dashboard, dashboardFinalizer)
			if err := r.Update(ctx, dashboard); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(dashboard, dashboardFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, dashboard); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(dashboard, dashboardFinalizer)
			if err := r.Update(ctx, dashboard); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	nextDashboard := dashboard.DeepCopy()

	manifest, err := r.getDashboardManifest(ctx, nextDashboard)
	if err != nil {
		err = fmt.Errorf("failed to get dashboard: %w", err)
		nextDashboard.SetReadyCondition(metav1.ConditionFalse, v1beta1.ContentUnavailableReason, err.Error())
		return r.reconcileResult(ctx, dashboard, nextDashboard, &errorRequeueDelay, err)
	}

	instances, err := getMatchingInstances(ctx, r.Client, dashboard.Spec.InstanceSelector)
	if err != nil {
		nextDashboard.SetReadyCondition(metav1.ConditionFalse, v1beta1.NoMatchingInstancesReason, err.Error())
		return r.reconcileResult(ctx, dashboard, nextDashboard, nil, err)
	}

	if nextDashboard.Status.Instances == nil {
		nextDashboard.Status.Instances = map[string]v1beta1.GrafanaDashboardInstanceStatus{}
	}

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
				err = fmt.Errorf("failed to reconcile plugins: %w", err)
				return r.reconcileResult(ctx, dashboard, nextDashboard, &errorRequeueDelay, err)
			}
		} else if dashboard.Spec.Plugins != nil {
			log.Error(nil, "plugin availability not ensured for external grafana instance")
		}

		instanceStatus, err := r.syncDashboardContent(ctx, grafana, nextDashboard, manifest)
		if err != nil {
			err = fmt.Errorf("error reconciling dashboard: %w", err)
			return r.reconcileResult(ctx, dashboard, nextDashboard, &errorRequeueDelay, err)
		}
		nextDashboard.Status.Instances[v1beta1.InstanceKeyFor(grafana)] = *instanceStatus
	}

	nextDashboard.SetReadyCondition(metav1.ConditionTrue, v1beta1.DashboardSyncedReason, "Dashboard synced")

	requeueAfter := dashboard.GetResyncPeriod()
	return r.reconcileResult(ctx, dashboard, nextDashboard, &requeueAfter, nil)
}

func (r *GrafanaDashboardReconciler) reconcileResult(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, nextDashboard *v1beta1.GrafanaDashboard, resyncDelay *time.Duration, err error) (reconcile.Result, error) {
	if !reflect.DeepEqual(dashboard.Status, nextDashboard.Status) {
		err := r.Client.Status().Update(context.Background(), nextDashboard)
		if err != nil {
			return ctrl.Result{RequeueAfter: errorRequeueDelay}, fmt.Errorf("failed to update status for dashboard %s", dashboard.Name)
		}
	}

	if resyncDelay == nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: *resyncDelay}, err
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

	folder, err := util.GetOrCreateFolder(grafanaClient, dashboard)
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
		existingMatches, err := util.ExistingDashboardVersionMatches(grafanaClient, instanceStatus)
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
		Message:   fmt.Sprintf("Updated by Grafana Operator. Generation %d, ResourceVersion: %s", dashboard.Generation, dashboard.ResourceVersion),
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
		if message := dashboard.ContentErrorBackoff(); message != "" {
			dashboard.SetReadyCondition(metav1.ConditionFalse, v1beta1.ErrorBackoffReason, message)
			return nil, fmt.Errorf("currently in error backoff: %s", message)
		}

		manifestBytes, err = r.getRemoteDashboardManifest(ctx, dashboard, dashboard.Spec.Source.Remote)
		dashboard.SetStatusContentError(err)
		if err != nil {
			if dashboard.Status.Content != nil && dashboard.Status.Content.Cache != nil {
				dashboard.SetCondition(metav1.Condition{
					Type:    "StaleContent",
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ContentUnavailableReason,
					Message: err.Error(),
				})
				manifestBytes, err = v1beta1.Gunzip(dashboard.Status.Content.Cache)
				if err != nil {
					return nil, fmt.Errorf("failed to gunzip dashboard content (stale): %w", err)
				}
			} else {
				dashboard.SetCondition(metav1.Condition{
					Type:    "StaleContent",
					Status:  metav1.ConditionFalse,
					Reason:  v1beta1.ContentUnavailableReason,
					Message: err.Error(),
				})
			}
		} else {
			dashboard.SetCondition(metav1.Condition{
				Type:   "StaleContent",
				Status: metav1.ConditionFalse,
				Reason: v1beta1.ContentAvailableReason,
			})
		}
	} else {
		return nil, fmt.Errorf("missing source for dashboard %s/%s", dashboard.Namespace, dashboard.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch dashboard content: %w", err)
	}

	var manifest map[string]interface{}
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal dashboard: %s", err)
		dashboard.SetReadyCondition(metav1.ConditionFalse, v1beta1.ContentUnavailableReason, err.Error())
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
		var err error
		url, err = util.GetGrafanaComDashboardUrl(ctx, *source.GrafanaCom)
		if err != nil {
			return nil, err
		}
	}

	cache := dashboard.GetContentCache(url)
	if len(cache) > 0 {
		return cache, nil
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w ", err)
	}

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.DashboardUrlRequests)
	response, err := client.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w ", err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status not ok: http status %d, body: %s ", response.StatusCode, content)
	}

	gz, err := v1beta1.Gzip(content)
	if err != nil {
		return nil, fmt.Errorf("failed to gzip dashboard: %w ", err)
	}

	dashboard.Status.Content = &v1beta1.GrafanaDashboardStatusContent{
		Cache:          gz,
		CacheTimestamp: metav1.Now(),
		Url:            url,
	}

	return content, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.GrafanaDashboard{}, dashboardConfigMapField, func(rawObj client.Object) []string {
		// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
		configDeployment := rawObj.(*v1beta1.GrafanaDashboard)
		if configDeployment.Spec.Source.ConfigMap == nil || configDeployment.Spec.Source.ConfigMap.Name == "" {
			return nil
		}
		return []string{configDeployment.Spec.Source.ConfigMap.Name}
	}); err != nil {
		return err
	}

	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDashboard{}).
		Watches(
			&source.Kind{Type: &v1beta1.Grafana{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForGrafana),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}), // api availability change should trigger dashboard reconcile
		).
		Watches(
			&source.Kind{Type: &v1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)

	return err
}

func (r *GrafanaDashboardReconciler) findObjectsForConfigMap(configMap client.Object) []reconcile.Request {
	configMapDashboards := &v1beta1.GrafanaDashboardList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(dashboardConfigMapField, configMap.GetName()),
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

func (r *GrafanaDashboardReconciler) findObjectsForGrafana(o client.Object) []reconcile.Request {
	grafanaLabels := o.GetLabels()
	ctx := context.Background()
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
