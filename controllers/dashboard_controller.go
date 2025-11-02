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

	"k8s.io/utils/strings/slices"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	conditionDashboardSynchronized        = "DashboardSynchronized"
	conditionReasonInvalidModelResolution = "InvalidModelResolution"
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) { //nolint:gocyclo
	log := logf.FromContext(ctx).WithName("GrafanaDashboardReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaDashboard{}

	err := r.Get(ctx, req.NamespacedName, cr)
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

	defer UpdateStatus(ctx, r.Client, cr)

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	// Retrieving the model before the loop ensures to exit early in case of failure and not fail once per matching instance
	resolver := content.NewContentResolver(cr, r.Client)

	dashboardModel, hash, err := resolver.Resolve(ctx)
	if err != nil {
		// Resolve has a lot of failure cases.
		// fetch content errors could be a temporary network issue but would result in an InvalidSpec condition
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, conditionReasonInvalidModelResolution, err.Error())
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

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	cr.Status.NoMatchingInstances = false

	log.Info("found matching Grafana instances for dashboard", "count", len(instances))

	uid := fmt.Sprintf("%s", dashboardModel["uid"])
	log = log.WithValues("uid", uid)
	ctx = logf.IntoContext(ctx, log)

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

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf(ErrFetchingFolder, err)
	}

	applyHomeErrors := make(map[string]string)
	pluginErrors := make(map[string]string)
	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested dashboards to a configmap from where the
			// grafana reconciler will pick them up
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, cr.Spec.Plugins, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
			if err != nil {
				pluginErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			}
		}

		// then import the dashboard into the matching grafana instances
		err = r.onDashboardCreated(ctx, &grafana, cr, dashboardModel, hash, folderUID)
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

	if len(pluginErrors) > 0 {
		err := fmt.Errorf("%v", pluginErrors)
		log.Error(err, "failed to apply plugins to all instances")
	}

	if len(applyHomeErrors) > 0 {
		err := fmt.Errorf("%v", applyHomeErrors)
		log.Error(err, "failed to apply home dashboards to all instances")
	}

	allApplyErrors := mergeReconcileErrors(applyErrors, pluginErrors, applyHomeErrors)

	condition := buildSynchronizedCondition("Dashboard", conditionDashboardSynchronized, cr.Generation, allApplyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(allApplyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", allApplyErrors)
	}

	cr.Status.Hash = hash
	cr.Status.UID = uid

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaDashboardReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaDashboard) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaDashboard")

	uid := content.CustomUIDOrUID(cr, cr.Status.UID)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("creating grafana http client: %w", err)
		}

		isCleanupInGrafanaRequired := true

		resp, err := grafanaClient.Dashboards.GetDashboardByUID(uid)
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

			_, err = grafanaClient.Dashboards.DeleteDashboardByUID(uid) //nolint:errcheck
			if err != nil {
				var notFound *dashboards.DeleteDashboardByUIDNotFound
				if !errors.As(err, &notFound) {
					return fmt.Errorf("deleting dashboard from instance: %w", err)
				}
			}

			if dash != nil && dash.Meta != nil && dash.Meta.FolderUID != "" && cr.Spec.FolderRef == "" && cr.Spec.FolderUID == "" {
				log.V(1).Info("Folder qualifies for deletion, checking if empty")

				resp, err := r.DeleteFolderIfEmpty(grafanaClient, dash.Meta.FolderUID)
				if err != nil {
					return fmt.Errorf("deleting empty parent folder from instance: %w", err)
				}

				if resp.StatusCode == http.StatusOK {
					log.Info("unused folder successfully removed")
				}

				if resp.StatusCode == 432 {
					log.Info("folder still in use by other dashboards, libraryPanels, or alertrules")
				}
			}
		}

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
			if err != nil {
				return fmt.Errorf("reconciling plugins: %w", err)
			}
		}

		// Update grafana instance Status
		err = grafana.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return fmt.Errorf("removing dashboard from grafana cr: %w", err)
		}
	}

	return nil
}

func (r *GrafanaDashboardReconciler) onDashboardCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDashboard, dashboardModel map[string]any, hash, folderUID string) error {
	log := logf.FromContext(ctx)

	if grafana.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your dashboard cr")
	}

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return fmt.Errorf("creating grafana http client: %w", err)
	}

	if folderUID == "" {
		folderUID, err = r.GetOrCreateFolder(grafanaClient, cr)
		if err != nil {
			return err
		}
	}

	uid := fmt.Sprintf("%s", dashboardModel["uid"])
	title := fmt.Sprintf("%s", dashboardModel["title"])
	remoteUID := uid

	if cr.Spec.CustomUID == "" {
		log.V(1).Info(".spec.uid empty, verifying uid has not changed using search")

		remoteUID, err = r.Exists(grafanaClient, uid, title, folderUID)
		if err != nil {
			return err
		}
	}

	dashWithMeta, err := grafanaClient.Dashboards.GetDashboardByUID(remoteUID)
	if err != nil {
		var notFound *dashboards.GetDashboardByUIDNotFound
		if !errors.As(err, &notFound) {
			return err
		}
	}

	exists := dashWithMeta != nil
	if exists && (remoteUID != uid || dashWithMeta.Payload.Meta.FolderUID != folderUID) {
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

	// Update when missing or the CR is updated
	if exists && content.Unchanged(cr, hash) {
		log.V(1).Info("dashboard model unchanged. skipping remaining requests")
		return nil
	}

	remoteChanged, err := r.hasRemoteChange(exists, dashboardModel, dashWithMeta)
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

	// Update grafana instance Status
	return grafana.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource(uid))
}

func (r *GrafanaDashboardReconciler) Exists(client *genapi.GrafanaHTTPAPI, uid string, title string, folderUID string) (string, error) {
	tvar := "dash-db"

	page := int64(1)

	limit := int64(1000)
	for {
		params := search.NewSearchParams().WithType(&tvar).WithLimit(&limit).WithPage(&page)

		resp, err := client.Search.Search(params)
		if err != nil {
			return "", err
		}

		hits := resp.GetPayload()

		for _, hit := range hits {
			if hit.UID == uid || (hit.Title == title && hit.FolderUID == folderUID) {
				return hit.UID, err
			}
		}

		if len(hits) < int(limit) {
			break
		}

		page++
	}

	return "", nil
}

// HasRemoteChange checks if a dashboard in Grafana is different to the model defined in the custom resources
func (r *GrafanaDashboardReconciler) hasRemoteChange(exists bool, model map[string]any, remoteDashboard *dashboards.GetDashboardByUIDOK) (bool, error) {
	if !exists {
		// if the dashboard doesn't exist, don't even request
		return true, nil
	}

	remoteModel, ok := (remoteDashboard.GetPayload().Dashboard).(map[string]any)
	if !ok {
		return true, fmt.Errorf("remote dashboard is not a valid object")
	}

	keys := make([]string, 0, len(model))
	for key := range model {
		keys = append(keys, key)
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
			StatusCode: http.StatusInternalServerError,
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
				StatusCode: http.StatusInternalServerError,
			}, err
		}
	}

	return http.Response{
		Status:     "grafana folder deleted",
		StatusCode: http.StatusOK,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	const (
		configMapIndexKey string = ".metadata.configMap"
	)

	// Index the dashboards by the ConfigMap references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &v1beta1.GrafanaDashboard{}, configMapIndexKey,
		r.indexConfigMapSource()); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDashboard{}, builder.WithPredicates(
			ignoreStatusUpdates(),
		)).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(configMapIndexKey)),
		).
		Complete(r)
}

func (r *GrafanaDashboardReconciler) indexConfigMapSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		dashboard, ok := o.(*v1beta1.GrafanaDashboard)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaDashboard, got %T", o))
		}

		if dashboard.Spec.ConfigMapRef != nil {
			return []string{fmt.Sprintf("%s/%s", dashboard.Namespace, dashboard.Spec.ConfigMapRef.Name)}
		}

		return nil
	}
}

func (r *GrafanaDashboardReconciler) requestsForChangeByField(indexKey string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		var list v1beta1.GrafanaDashboardList
		if err := r.List(ctx, &list, client.MatchingFields{
			indexKey: fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
		}); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for _, dashboard := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: dashboard.Namespace,
				Name:      dashboard.Name,
			}})
		}

		return reqs
	}
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
