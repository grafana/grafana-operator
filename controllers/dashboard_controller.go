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
	"net/http"
	"strings"
	"time"

	"github.com/grafana-operator/grafana-operator/v5/embeds"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana-operator/grafana-operator/v5/controllers/client"
	"github.com/grafana-operator/grafana-operator/v5/controllers/fetchers"
	"github.com/grafana-operator/grafana-operator/v5/controllers/metrics"
	grapi "github.com/grafana/grafana-api-golang-client"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	initialSyncDelay = "10s"
	syncBatchSize    = 100
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	Client    client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Discovery discovery.DiscoveryInterface
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update

func (r *GrafanaDashboardReconciler) syncDashboards(ctx context.Context) (ctrl.Result, error) {
	syncLog := log.FromContext(ctx)
	dashboardsSynced := 0

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

	// get all dashboards
	allDashboards := &v1beta1.GrafanaDashboardList{}
	err = r.Client.List(ctx, allDashboards, opts...)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	dashboardsToDelete := getDashboardsToDelete(allDashboards, grafanas.Items)

	// delete all dashboards that no longer have a cr
	for grafana, dashboards := range dashboardsToDelete {
		grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		for _, dashboard := range dashboards {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to a certain number of dashboards per cycle. This means that it will take longer to sync
			// a large number of deleted dashboard crs, but that should be an edge case.
			if dashboardsSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := dashboard.Split()
			err = grafanaClient.DeleteDashboardByUID(uid)
			if err != nil {
				if strings.Contains(err.Error(), "status: 404") {
					syncLog.Info("dashboard no longer exists", "namespace", namespace, "name", name)
				} else {
					return ctrl.Result{Requeue: false}, err
				}
			}

			grafana.Status.Dashboards = grafana.Status.Dashboards.Remove(namespace, name)
			dashboardsSynced += 1
		}

		// one update per grafana - this will trigger a reconcile of the grafana controller
		// so we should minimize those updates
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}
	}

	if dashboardsSynced > 0 {
		syncLog.Info("successfully synced dashboards", "dashboards", dashboardsSynced)
	}
	return ctrl.Result{Requeue: false}, nil
}

// sync dashboards, delete dashboards from grafana that do no longer have a cr
func getDashboardsToDelete(allDashboards *v1beta1.GrafanaDashboardList, grafanas []v1beta1.Grafana) map[*v1beta1.Grafana][]v1beta1.NamespacedResource {
	dashboardsToDelete := map[*v1beta1.Grafana][]v1beta1.NamespacedResource{}
	for _, grafana := range grafanas {
		grafana := grafana
		for _, dashboard := range grafana.Status.Dashboards {
			if allDashboards.Find(dashboard.Namespace(), dashboard.Name()) == nil {
				dashboardsToDelete[&grafana] = append(dashboardsToDelete[&grafana], dashboard)
			}
		}
	}
	return dashboardsToDelete
}

func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx)
	r.Log = controllerLog

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncDashboards(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialDashboardSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	dashboard := &v1beta1.GrafanaDashboard{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, dashboard)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.onDashboardDeleted(ctx, req.Namespace, req.Name)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, err
			}
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana dashboard cr")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	instances, err := r.GetMatchingDashboardInstances(ctx, dashboard, r.Client)
	if err != nil {
		controllerLog.Error(err, "could not find matching instances", "name", dashboard.Name, "namespace", dashboard.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	controllerLog.Info("found matching Grafana instances for dashboard", "count", len(instances.Items))

	success := true
	for _, grafana := range instances.Items {
		// check if this is a cross namespace import
		if grafana.Namespace != dashboard.Namespace && !dashboard.IsAllowCrossNamespaceImport() {
			continue
		}

		grafana := grafana
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		// TODO: status.conditions for grafana CRD and implement grafana.Ready()
		if grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			success = false
			continue
		}

		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested dashboards to a configmap from where the
			// grafana reconciler will pick them up
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, dashboard.Spec.Plugins, fmt.Sprintf("%v-dashboard", dashboard.Name))
			if err != nil {
				controllerLog.Error(err, "error reconciling plugins", "dashboard", dashboard.Name, "grafana", grafana.Name)
				success = false
			}
		} else if dashboard.Spec.Plugins != nil {
			return ctrl.Result{}, fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your dashboard cr")
		}

		// then import the dashboard into the matching grafana instances
		err = r.onDashboardCreated(ctx, &grafana, dashboard)
		if err != nil {
			controllerLog.Error(err, "error reconciling dashboard", "dashboard", dashboard.Name, "grafana", grafana.Name)
			success = false
		}
	}

	// if the dashboard was successfully synced in all instances, wait for its re-sync period
	if success {
		return ctrl.Result{RequeueAfter: dashboard.GetResyncPeriod()}, nil
	}

	return ctrl.Result{RequeueAfter: RequeueDelay}, nil
}

func (r *GrafanaDashboardReconciler) onDashboardDeleted(ctx context.Context, namespace string, name string) error {
	list := v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, &list, opts...)
	if err != nil {
		return err
	}

	for _, grafana := range list.Items {
		if found, uid := grafana.Status.Dashboards.Find(namespace, name); found {
			grafana := grafana
			grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			dash, err := grafanaClient.DashboardByUID(*uid)
			if err != nil {
				if !strings.Contains(err.Error(), "status: 404") {
					return err
				}
			}

			err = grafanaClient.DeleteDashboardByUID(*uid)
			if err != nil {
				if !strings.Contains(err.Error(), "status: 404") {
					return err
				}
			}

			if dash != nil && dash.Meta.Folder > 0 {
				resp, err := r.DeleteFolderIfEmpty(grafanaClient, dash.Folder)
				if err != nil {
					return err
				}
				if resp.StatusCode == 200 {
					r.Log.Info("unused folder successfully removed")
				}
				if resp.StatusCode == 432 {
					r.Log.Info("folder still in use by other dashboards")
				}
			}

			if grafana.IsInternal() {
				err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, fmt.Sprintf("%v-dashboard", name))
				if err != nil {
					return err
				}
			}

			grafana.Status.Dashboards = grafana.Status.Dashboards.Remove(namespace, name)
			err = r.Client.Status().Update(ctx, &grafana)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *GrafanaDashboardReconciler) onDashboardCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDashboard) error {
	manifestBytes, err := r.fetchDashboardManifest(cr)
	if err != nil {
		r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.ContentUnavailableReason, fmt.Sprintf("failed to get dashboard: %s", err))
		return err
	}
	var manifest map[string]interface{}
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.ContentUnavailableReason, fmt.Sprintf("failed to unmarshal dashboard: %s", err))
		return err
	}

	if _, ok := manifest["uid"]; !ok {
		manifest["uid"] = string(cr.UID)
	}

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.GrafanaApiErrorReason, fmt.Sprintf("failed to create grafana client: %s", err))
		return err
	}

	folder, err := r.GetOrCreateFolder(grafanaClient, cr)
	if err != nil {
		r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.GrafanaApiErrorReason, fmt.Sprintf("failed get or create folder: %s", err))
		return errors.NewInternalError(err)
	}
	folderId := int64(0)
	if folder != nil {
		folderId = folder.ID
	}

	shouldCreate := true
	instanceStatus, ok := cr.Status.Instances[string(grafana.GetUID())]
	if ok {
		existingMatches, err := r.ExistingVersionMatchesStatus(grafanaClient, instanceStatus)
		if err != nil {
			r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.GrafanaApiErrorReason, fmt.Sprintf("failed check for existing dashboard: %s", err))
			return err
		}
		shouldCreate = !existingMatches
	} else {
		if cr.Status.Instances == nil {
			cr.Status.Instances = make(map[string]v1beta1.GrafanaDashboardInstanceStatus)
		}
		instanceStatus = v1beta1.GrafanaDashboardInstanceStatus{
			UID: manifest["uid"].(string),
		}
		cr.Status.Instances[string(grafana.GetUID())] = instanceStatus
	}
	if !shouldCreate {
		return nil // Updating status here will trigger another reconcile which is no bueno, perhaps this logic should happen within .Reconcile()?
	}

	resp, err := grafanaClient.NewDashboard(grapi.Dashboard{
		Meta: grapi.DashboardMeta{
			IsStarred: false,
			Slug:      cr.Name,
			Folder:    folderId,
		},
		Model:     manifest,
		Folder:    folderId,
		Overwrite: true,
		Message:   "",
	})
	if err != nil {
		return err
	}
	if resp.Status != "success" {
		r.SetReadyCondition(ctx, cr, v1.ConditionFalse, v1beta1.GrafanaApiErrorReason, fmt.Sprintf("failed create dashboard: %s", err))
		return errors.NewBadRequest(fmt.Sprintf("error creating dashboard, status was %v", resp.Status))
	}

	cr.Status.Instances[string(grafana.GetUID())] = v1beta1.GrafanaDashboardInstanceStatus{
		UID:     resp.UID,
		Version: resp.Version,
	}

	grafana.Status.Dashboards = grafana.Status.Dashboards.Add(cr.Namespace, cr.Name, resp.UID)
	err = r.Client.Status().Update(ctx, grafana)
	if err != nil {
		return err
	}

	return r.SetReadyCondition(ctx, cr, v1.ConditionTrue, v1beta1.DashboardSyncedReason, "Dashboard synced")
}

// fetchDashboardJson delegates obtaining the dashboard json definition to one of the known fetchers, for example
// from embedded raw json or from a url
func (r *GrafanaDashboardReconciler) fetchDashboardManifest(dashboard *v1beta1.GrafanaDashboard) ([]byte, error) {
	sourceTypes := dashboard.GetSourceTypes()

	if len(sourceTypes) == 0 {
		return nil, fmt.Errorf("no source type provided for dashboard %v", dashboard.Name)
	}

	if len(sourceTypes) > 1 {
		return nil, fmt.Errorf("more than one source types found for dashboard %v", dashboard.Name)
	}

	switch sourceTypes[0] {
	case v1beta1.DashboardSourceTypeRawJson:
		return []byte(dashboard.Spec.Json), nil
	case v1beta1.DashboardSourceTypeGzipJson:
		return v1beta1.Gunzip([]byte(dashboard.Spec.GzipJson))
	case v1beta1.DashboardSourceTypeUrl:
		return fetchers.FetchDashboardFromUrl(dashboard)
	case v1beta1.DashboardSourceTypeJsonnet:
		return fetchers.FetchJsonnet(dashboard, embeds.GrafonnetEmbed)
	case v1beta1.DashboardSourceTypeGrafanaCom:
		return fetchers.FetchDashboardFromGrafanaCom(dashboard)
	default:
		return nil, fmt.Errorf("unknown source type %v found in dashboard %v", sourceTypes[0], dashboard.Name)
	}
}

func (r *GrafanaDashboardReconciler) ExistingVersionMatchesStatus(client *grapi.Client, instanceStatus v1beta1.GrafanaDashboardInstanceStatus) (bool, error) {
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

func (r *GrafanaDashboardReconciler) GetOrCreateFolder(client *grapi.Client, cr *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
	if cr.Spec.FolderTitle == "" {
		return nil, nil
	}

	folder, err := r.GetFolder(client, cr)
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

func (r *GrafanaDashboardReconciler) GetFolder(client *grapi.Client, cr *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
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

func (r *GrafanaDashboardReconciler) DeleteFolderIfEmpty(client *grapi.Client, folderID int64) (http.Response, error) {
	dashboards, err := client.Dashboards()
	if err != nil {
		return http.Response{
			Status:     "internal grafana client error getting dashboards",
			StatusCode: 500,
		}, err
	}

	for _, dashboard := range dashboards {
		if int64(dashboard.FolderID) == folderID {
			return http.Response{
				Status:     "resource is still in use",
				StatusCode: 423, // Locked return code
			}, err
		}
		continue
	}

	folder, err := client.Folder(folderID)
	if err != nil {
		return http.Response{
			Status:     "internal grafana client error getting folder UID for folder",
			StatusCode: 500,
		}, err
	}

	if err = client.DeleteFolder(folder.UID); err != nil {
		return http.Response{
			Status:     "internal grafana client error deleting grafana folder",
			StatusCode: 500,
		}, err
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
						r.Log.Error(err, "error synchronizing dashboards")
						continue
					}
					if result.Requeue {
						r.Log.Info("more dashboards left to synchronize")
						continue
					}
					r.Log.Info("dashboard sync complete")
					return
				}
			}
		}()
	}

	return err
}

func (r *GrafanaDashboardReconciler) GetMatchingDashboardInstances(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, k8sClient client.Client) (v1beta1.GrafanaList, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, dashboard.Spec.InstanceSelector)
	if err != nil || len(instances.Items) == 0 {
		r.SetReadyCondition(ctx, dashboard, v1.ConditionFalse, v1beta1.NoMatchingInstancesReason, "No instances found matching .spec.instanceSelector")
		return v1beta1.GrafanaList{}, err
	}

	return instances, err
}

func (r *GrafanaDashboardReconciler) SetReadyCondition(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, status v1.ConditionStatus, reason string, message string) error {
	dashboard.SetReadyCondition(status, reason, message)
	r.Log.Info("updating status", "conditions", dashboard.Status.Conditions, "instances", dashboard.Status.Instances)
	if err := r.Client.Status().Update(ctx, dashboard); err != nil {
		r.Log.Info("unable to update the status of %v, in %v", dashboard.Name, dashboard.Namespace)
		return err
	}
	return nil
}
