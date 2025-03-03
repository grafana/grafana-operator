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
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"k8s.io/utils/strings/slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	conditionDashboardSynchronized = "DashboardSynchronized"
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update

func (r *GrafanaDashboardReconciler) syncDashboards(ctx context.Context) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	dashboardsSynced := 0

	// get all grafana instances
	grafanas := &v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, grafanas, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// no instances, no need to sync
	if len(grafanas.Items) == 0 {
		return ctrl.Result{Requeue: false}, nil
	}

	// get all dashboards
	allDashboards := &v1beta1.GrafanaDashboardList{}
	err = r.Client.List(ctx, allDashboards, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	dashboardsToDelete := getDashboardsToDelete(allDashboards, grafanas.Items)

	// delete all dashboards that no longer have a cr
	for grafana, oldDashboards := range dashboardsToDelete {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{}, err
		}

		for _, dashboard := range oldDashboards {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to a certain number of dashboards per cycle. This means that it will take longer to sync
			// a large number of deleted dashboard crs, but that should be an edge case.
			if dashboardsSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := dashboard.Split()
			_, err = grafanaClient.Dashboards.DeleteDashboardByUID(uid) //nolint:errcheck
			if err != nil {
				var notFound *dashboards.DeleteDashboardByUIDNotFound
				if !errors.As(err, &notFound) {
					return ctrl.Result{}, err
				}
			}

			grafana.Status.Dashboards = grafana.Status.Dashboards.Remove(namespace, name)
			dashboardsSynced += 1
		}

		// one update per grafana - this will trigger a reconcile of the grafana controller
		// so we should minimize those updates
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if dashboardsSynced > 0 {
		log.Info("successfully synced dashboards", "dashboards", dashboardsSynced)
	}
	return ctrl.Result{Requeue: false}, nil
}

// sync dashboards, delete dashboards from grafana that do no longer have a cr
func getDashboardsToDelete(allDashboards *v1beta1.GrafanaDashboardList, grafanas []v1beta1.Grafana) map[*v1beta1.Grafana][]v1beta1.NamespacedResource {
	dashboardsToDelete := map[*v1beta1.Grafana][]v1beta1.NamespacedResource{}
	for _, grafana := range grafanas {
		for _, dashboard := range grafana.Status.Dashboards {
			if allDashboards.Find(dashboard.Namespace(), dashboard.Name()) == nil {
				dashboardsToDelete[&grafana] = append(dashboardsToDelete[&grafana], dashboard)
			}
		}
	}
	return dashboardsToDelete
}

func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) { //nolint:gocyclo
	log := logf.FromContext(ctx).WithName("GrafanaDashboardReconciler")
	ctx = logf.IntoContext(ctx, log)

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncDashboards(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialDashboardSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	cr := &v1beta1.GrafanaDashboard{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting grafana dashboard cr: %w", err)
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

	defer func() {
		cr.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, cr); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(cr.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, cr); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	// Retrieving the model before the loop ensures to exit early in case of failure and not fail once per matching instance
	resolver := content.NewContentResolver(cr, r.Client)
	dashboardModel, hash, err := resolver.Resolve(ctx)
	if err != nil {
		// Resolve has a lot of failure cases.
		// fetch content errors could be a temporary network issue but would result in an InvalidSpec condition
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidModelResolution", err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDashboardSynchronized)
		return ctrl.Result{}, fmt.Errorf("resolving dashboard contents: %w", err)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDashboardSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}
	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDashboardSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	cr.Status.NoMatchingInstances = false
	log.Info("found matching Grafana instances for dashboard", "count", len(instances))

	uid := fmt.Sprintf("%s", dashboardModel["uid"])

	// Garbage collection for a case where dashboard uid get changed, dashboard creation is expected to happen in a separate reconcilication cycle
	if content.IsUpdatedUID(cr, uid) {
		log.Info("dashboard uid got updated, deleting dashboards with the old uid")
		if err = r.finalize(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}

		// Clean up uid, so further reconciliations can track changes there
		cr.Status.UID = ""

		// Trigger the next reconciliation right away
		return ctrl.Result{Requeue: true}, nil
	}

	applyHomeErrors := make(map[string]string)
	pluginErrors := make(map[string]string)
	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested dashboards to a configmap from where the
			// grafana reconciler will pick them up
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, cr.Spec.Plugins, fmt.Sprintf("%v-dashboard", cr.Name))
			if err != nil {
				pluginErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
				log.Error(err, "error reconciling plugins", "dashboard", cr.Name, "grafana", grafana.Name)
			}
		}

		// then import the dashboard into the matching grafana instances
		err = r.onDashboardCreated(ctx, &grafana, cr, dashboardModel, hash)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}

		if grafana.Spec.Preferences != nil && uid == grafana.Spec.Preferences.HomeDashboardUID {
			err = r.UpdateHomeDashboard(ctx, grafana, uid, cr)
			if err != nil {
				applyHomeErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			}
		}
	}

	// TODO Add new Condition displaying plugin reconciliation errors
	if len(pluginErrors) > 0 {
		err := fmt.Errorf("%v", pluginErrors)
		log.Error(err, "failed to apply plugins to all instances")
	}

	// Specific to dashboards
	// NOTE Add new Condition displaying home dash sync errors?
	if len(applyHomeErrors) > 0 {
		err := fmt.Errorf("%v", pluginErrors)
		log.Error(err, "failed to apply home dashboards to all instances")
	}

	condition := buildSynchronizedCondition("Dashboard", conditionDashboardSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	cr.Status.Hash = hash
	cr.Status.UID = uid
	return ctrl.Result{RequeueAfter: cr.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaDashboardReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaDashboard) error {
	log := logf.FromContext(ctx)
	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		found, uid := grafana.Status.Dashboards.Find(cr.Namespace, cr.Name)
		if !found {
			continue
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("creating grafana http client: %w", err)
		}

		isCleanupInGrafanaRequired := true

		resp, err := grafanaClient.Dashboards.GetDashboardByUID(*uid)
		if err != nil {
			var notFound *dashboards.GetDashboardByUIDNotFound
			if !errors.As(err, &notFound) {
				return fmt.Errorf("fetching dashboard from instance: %w", err)
			}

			isCleanupInGrafanaRequired = false
		}

		if isCleanupInGrafanaRequired {
			var dash *models.DashboardFullWithMeta
			if resp != nil {
				dash = resp.GetPayload()
			}

			_, err = grafanaClient.Dashboards.DeleteDashboardByUID(*uid) //nolint:errcheck
			if err != nil {
				var notFound *dashboards.DeleteDashboardByUIDNotFound
				if !errors.As(err, &notFound) {
					return fmt.Errorf("deleting dashboard from instance: %w", err)
				}
			}

			if dash != nil && dash.Meta != nil && dash.Meta.FolderUID != "" {
				resp, err := r.DeleteFolderIfEmpty(grafanaClient, dash.Meta.FolderUID)
				if err != nil {
					return fmt.Errorf("deleting empty parent folder from instance: %w", err)
				}
				if resp.StatusCode == 200 {
					log.Info("unused folder successfully removed")
				}
				if resp.StatusCode == 432 {
					log.Info("folder still in use by other dashboards, libraryPanels, or alertrules")
				}
			}
		}

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-dashboard", cr.Name))
			if err != nil {
				return fmt.Errorf("reconciling plugins: %w", err)
			}
		}

		// Update status of Grafana instance
		grafana.Status.Dashboards = grafana.Status.Dashboards.Remove(cr.Namespace, cr.Name)
		err = r.Client.Status().Update(ctx, &grafana)
		if err != nil {
			return fmt.Errorf("updating grafana cr status %s/%s: %w", grafana.Namespace, grafana.Name, err)
		}
	}

	return nil
}

func (r *GrafanaDashboardReconciler) onDashboardCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDashboard, dashboardModel map[string]interface{}, hash string) error {
	log := logf.FromContext(ctx)
	if grafana.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your dashboard cr")
	}

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return fmt.Errorf("creating grafana http client: %w", err)
	}

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		return err
	}

	if folderUID == "" {
		folderUID, err = r.GetOrCreateFolder(grafanaClient, cr)
		if err != nil {
			return err
		}
	}

	uid := fmt.Sprintf("%s", dashboardModel["uid"])
	title := fmt.Sprintf("%s", dashboardModel["title"])

	// TODO Check if cr.Spec.CustomUID is defined and skip the search. Get the dashboard directly with the UID
	// The 'Exists' function is unreasonably expensive if reconciling multiple instances with a lot of :wqdashboards.
	exists, remoteUID, err := r.Exists(grafanaClient, uid, title, folderUID)
	if err != nil {
		return err
	}

	if exists && remoteUID != uid {
		// If there's already a dashboard with the same title in the same folder, grafana preserves that dashboard's uid, so we should remove it first
		log.Info("found dashboard with the same title (in the same folder) but different uid, removing the dashboard before recreating it with a new uid")
		_, err = grafanaClient.Dashboards.DeleteDashboardByUID(remoteUID) //nolint:errcheck
		if err != nil {
			var notFound *dashboards.DeleteDashboardByUIDNotFound
			if !errors.As(err, &notFound) {
				return err
			}
		}

		exists = false
	}

	if exists && content.Unchanged(cr, hash) && !cr.ResyncPeriodHasElapsed() {
		return nil
	}

	// NOTE Only the dashboard reconciler will detect if the cr and remote model differ by walking the structure
	// Most others instead apply the resource and outsource it to the instance or check a hash.
	remoteChanged, err := r.hasRemoteChange(exists, grafanaClient, uid, dashboardModel)
	if err != nil {
		return err
	}

	if !remoteChanged {
		return nil
	}

	resp, err := grafanaClient.Dashboards.PostDashboard(&models.SaveDashboardCommand{
		Dashboard: dashboardModel,
		FolderUID: folderUID,
		Overwrite: true,
	})
	if err != nil {
		return err
	}

	payload := resp.GetPayload()

	if payload.Status == nil || *payload.Status != "success" {
		return kuberr.NewBadRequest(fmt.Sprintf("error creating dashboard, status was %v", payload.Status))
	}

	grafana.Status.Dashboards = grafana.Status.Dashboards.Add(cr.Namespace, cr.Name, uid)
	return r.Client.Status().Update(ctx, grafana)
}

func (r *GrafanaDashboardReconciler) Exists(client *genapi.GrafanaHTTPAPI, uid string, title string, folderUID string) (bool, string, error) {
	tvar := "dash-db"

	page := int64(1)
	limit := int64(1000)
	for {
		params := search.NewSearchParams().WithType(&tvar).WithLimit(&limit).WithPage(&page)
		resp, err := client.Search.Search(params)
		if err != nil {
			return false, "", err
		}
		results := resp.GetPayload()

		for _, dashboard := range results {
			if dashboard.UID == uid || (dashboard.Title == title && dashboard.FolderUID == folderUID) {
				return true, dashboard.UID, nil
			}
		}
		if len(results) < int(limit) {
			break
		}
		page++
	}
	return false, "", nil
}

// HasRemoteChange checks if a dashboard in Grafana is different to the model defined in the custom resources
func (r *GrafanaDashboardReconciler) hasRemoteChange(exists bool, client *genapi.GrafanaHTTPAPI, uid string, model map[string]interface{}) (bool, error) {
	if !exists {
		// if the dashboard doesn't exist, don't even request
		return true, nil
	}

	remoteDashboard, err := client.Dashboards.GetDashboardByUID(uid)
	if err != nil {
		var notFound *dashboards.GetDashboardByUIDNotFound
		if !errors.As(err, &notFound) {
			return true, nil
		}
		return false, err
	}

	keys := make([]string, 0, len(model))
	for key := range model {
		keys = append(keys, key)
	}

	remoteModel, ok := (remoteDashboard.GetPayload().Dashboard).(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("remote dashboard is not an object")
	}

	skipKeys := []string{"id", "version"} //nolint
	for _, key := range keys {
		// we do not keep track of those keys in the custom resource
		if slices.Contains(skipKeys, key) {
			continue
		}
		localValue := model[key]
		remoteValue := remoteModel[key]
		if !reflect.DeepEqual(localValue, remoteValue) {
			return true, nil
		}
	}

	return false, nil
}

func (r *GrafanaDashboardReconciler) GetOrCreateFolder(client *genapi.GrafanaHTTPAPI, cr *v1beta1.GrafanaDashboard) (string, error) {
	title := cr.Namespace
	if cr.Spec.FolderTitle != "" {
		title = cr.Spec.FolderTitle
	}

	exists, folderUID, err := r.GetFolderUID(client, title)
	if err != nil {
		return "", err
	}

	if exists {
		return folderUID, nil
	}

	// Folder wasn't found, let's create it
	body := &models.CreateFolderCommand{
		Title: title,
	}
	resp, err := client.Folders.CreateFolder(body)
	if err != nil {
		return "", err
	}
	folder := resp.GetPayload()
	if folder == nil {
		return "", fmt.Errorf("invalid payload returned")
	}

	return folder.UID, nil
}

func (r *GrafanaDashboardReconciler) GetFolderUID(
	client *genapi.GrafanaHTTPAPI,
	title string,
) (bool, string, error) {
	// Pre-existing folder that is not returned in Folder API
	if strings.EqualFold(title, "General") {
		return true, "", nil
	}
	page := int64(1)
	limit := int64(1000)
	for {
		params := folders.NewGetFoldersParams().WithPage(&page).WithLimit(&limit)

		foldersResp, err := client.Folders.GetFolders(params)
		if err != nil {
			return false, "", err
		}
		folders := foldersResp.GetPayload()

		for _, remoteFolder := range folders {
			if strings.EqualFold(remoteFolder.Title, title) {
				return true, remoteFolder.UID, nil
			}
		}
		if len(folders) < int(limit) {
			break
		}
		page++
	}

	return false, "", nil
}

func (r *GrafanaDashboardReconciler) DeleteFolderIfEmpty(client *genapi.GrafanaHTTPAPI, folderUID string) (http.Response, error) {
	params := search.NewSearchParams().WithFolderUIDs([]string{folderUID})
	results, err := client.Search.Search(params)
	if err != nil {
		return http.Response{
			Status:     "internal grafana client error getting dashboards",
			StatusCode: 500,
		}, err
	}
	if len(results.GetPayload()) > 0 {
		return http.Response{
			Status:     "resource is still in use",
			StatusCode: http.StatusLocked,
		}, err
	}

	deleteParams := folders.NewDeleteFolderParams().WithFolderUID(folderUID)
	if _, err = client.Folders.DeleteFolder(deleteParams); err != nil { //nolint:errcheck
		var notFound *folders.DeleteFolderNotFound
		if !errors.As(err, &notFound) {
			return http.Response{
				Status:     "internal grafana client error deleting grafana folder",
				StatusCode: 500,
			}, err
		}
	}
	return http.Response{
		Status:     "grafana folder deleted",
		StatusCode: 200,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDashboard{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)

	if err == nil {
		go func() {
			log := logf.FromContext(ctx).WithName("GrafanaDashboardReconciler")
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(initialSyncDelay):
					result, err := r.Reconcile(ctx, ctrl.Request{})
					if err != nil {
						log.Error(err, "error synchronizing dashboards")
						continue
					}
					if result.Requeue {
						log.Info("more dashboards left to synchronize")
						continue
					}
					log.Info("dashboard sync complete")
					return
				}
			}
		}()
	}

	return err
}

func (r *GrafanaDashboardReconciler) UpdateHomeDashboard(ctx context.Context, grafana v1beta1.Grafana, uid string, dashboard *v1beta1.GrafanaDashboard) error {
	log := logf.FromContext(ctx)
	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
	if err != nil {
		return err
	}

	_, err = grafanaClient.OrgPreferences.UpdateOrgPreferences(&models.UpdatePrefsCmd{ //nolint:errcheck
		HomeDashboardUID: uid,
	})
	if err != nil {
		log.Error(err, "unable to update the home dashboard", "namespace", dashboard.Namespace, "name", dashboard.Name)
		return err
	}

	log.Info("home dashboard configured", "namespace", dashboard.Namespace, "name", dashboard.Name)
	return nil
}
