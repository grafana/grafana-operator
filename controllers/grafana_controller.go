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
	"os"
	"strings"
	"time"

	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers/grafana"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionTypeGrafanaReady         = "GrafanaReady"
	conditionReasonReconcileSuspended = "ReconcileSuspended"
)

// GrafanaReconciler reconciles a Grafana object
type GrafanaReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	IsOpenShift   bool
	ClusterDomain string
}

// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=get;list;create;update;delete;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts;services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch;

func (r *GrafanaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &grafanav1beta1.Grafana{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "error getting grafana cr")

		return ctrl.Result{}, err
	}

	metrics.GrafanaReconciles.WithLabelValues(cr.Namespace, cr.Name).Inc()

	defer func() {
		if err := r.Status().Update(ctx, cr); err != nil {
			log.Error(err, "updating status")
		}
	}()

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonReconcileSuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	var stages []grafanav1beta1.OperatorStageName
	if cr.IsExternal() {
		// Only reconcile the Completion stage for external instances
		stages = []grafanav1beta1.OperatorStageName{grafanav1beta1.OperatorStageComplete}
		// AdminURL is normally set during ingress/route stage.
		// External instances only use the complete stage
		cr.Status.AdminURL = cr.Spec.External.URL
	} else {
		stages = getInstallationStages()

		// set spec.version to the current default version to avoid accidental updates when we change the default.
		if cr.Spec.Version == "" {
			err := r.setDefaultGrafanaVersion(ctx, cr)
			if err != nil {
				meta.RemoveStatusCondition(&cr.Status.Conditions, conditionTypeGrafanaReady)
				return ctrl.Result{}, fmt.Errorf("patching grafana version in spec: %w", err)
			}
		}
	}

	vars := &grafanav1beta1.OperatorReconcileVars{}

	for _, stage := range stages {
		log.Info("running stage", "stage", stage)

		cr.Status.Stage = stage
		reconciler := r.getReconcilerForStage(stage)

		if reconciler == nil {
			log.Info("no reconciler known for stage", "stage", stage)
			continue
		}

		stageStatus, err := reconciler.Reconcile(ctx, cr, vars, r.Scheme)
		if err != nil {
			cr.Status.StageStatus = stageStatus // In progress or failed, both accompanied by Error
			cr.Status.LastMessage = err.Error()

			metrics.GrafanaFailedReconciles.WithLabelValues(cr.Namespace, cr.Name, string(stage)).Inc()
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionTypeGrafanaReady)

			return ctrl.Result{}, fmt.Errorf("reconciler error in stage '%s': %w", stage, err)
		}
	}

	cr.Status.StageStatus = grafanav1beta1.OperatorStageResultSuccess
	cr.Status.LastMessage = ""

	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:               conditionTypeGrafanaReady, // Maybe use Grafana instead to be consistent with other conditions
		Reason:             "GrafanaReady",
		Message:            "Grafana reconcile completed",
		ObservedGeneration: cr.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	})

	return ctrl.Result{}, nil
}

func (r *GrafanaReconciler) setDefaultGrafanaVersion(ctx context.Context, cr client.Object) error {
	// For clusters where RELATED_IMAGE_GRAFANA is set to an image hash,
	// we want to set version to the value of the variable to support airgapped clusters as well
	targetVersion := config.GrafanaVersion
	if envVersion := os.Getenv("RELATED_IMAGE_GRAFANA"); isImageSHA256(envVersion) {
		targetVersion = envVersion
	}

	// Create patch with the target version
	patch, err := json.Marshal(map[string]any{"spec": map[string]any{"version": targetVersion}})
	if err != nil {
		return err
	}

	return r.Patch(ctx, cr, client.RawPatch(types.MergePatchType, patch))
}

func removeMissingCRs(statusList *grafanav1beta1.NamespacedResourceList, crs grafanav1beta1.NamespacedResourceImpl, updateStatus *bool) {
	toRemove := grafanav1beta1.NamespacedResourceList{}

	for _, r := range *statusList {
		namespace, name, _ := r.Split()
		if !crs.Exists(namespace, name) {
			toRemove = append(toRemove, r)
		}
	}

	if len(toRemove) > 0 {
		*statusList = statusList.RemoveEntries(&toRemove)
		*updateStatus = true
	}
}

func (r *GrafanaReconciler) syncStatuses(ctx context.Context) error {
	log := logf.FromContext(ctx)

	// get all grafana instances
	grafanas := &grafanav1beta1.GrafanaList{}

	err := r.List(ctx, grafanas)
	if err != nil {
		return err
	}
	// no instances, skip sync
	if len(grafanas.Items) == 0 {
		return nil
	}

	// Fetch all resources
	alertRuleGroups := &grafanav1beta1.GrafanaAlertRuleGroupList{}

	err = r.List(ctx, alertRuleGroups)
	if err != nil {
		return err
	}

	contactPoints := &grafanav1beta1.GrafanaContactPointList{}

	err = r.List(ctx, contactPoints)
	if err != nil {
		return err
	}

	dashboards := &grafanav1beta1.GrafanaDashboardList{}

	err = r.List(ctx, dashboards)
	if err != nil {
		return err
	}

	datasources := &grafanav1beta1.GrafanaDatasourceList{}

	err = r.List(ctx, datasources)
	if err != nil {
		return err
	}

	folders := &grafanav1beta1.GrafanaFolderList{}

	err = r.List(ctx, folders)
	if err != nil {
		return err
	}

	libraryPanels := &grafanav1beta1.GrafanaLibraryPanelList{}

	err = r.List(ctx, libraryPanels)
	if err != nil {
		return err
	}

	muteTimings := &grafanav1beta1.GrafanaLibraryPanelList{}

	err = r.List(ctx, muteTimings)
	if err != nil {
		return err
	}

	notificationTemplates := &grafanav1beta1.GrafanaNotificationTemplateList{}

	err = r.List(ctx, notificationTemplates)
	if err != nil {
		return err
	}

	// delete resources from grafana statuses that no longer have a CR
	statusUpdates := 0

	for _, grafana := range grafanas.Items {
		updateStatus := false

		removeMissingCRs(&grafana.Status.AlertRuleGroups, alertRuleGroups, &updateStatus)
		removeMissingCRs(&grafana.Status.ContactPoints, contactPoints, &updateStatus)
		removeMissingCRs(&grafana.Status.Dashboards, dashboards, &updateStatus)
		removeMissingCRs(&grafana.Status.Datasources, datasources, &updateStatus)
		removeMissingCRs(&grafana.Status.Folders, folders, &updateStatus)
		removeMissingCRs(&grafana.Status.LibraryPanels, libraryPanels, &updateStatus)
		removeMissingCRs(&grafana.Status.MuteTimings, muteTimings, &updateStatus)
		removeMissingCRs(&grafana.Status.NotificationTemplates, notificationTemplates, &updateStatus)

		if updateStatus {
			statusUpdates += 1

			err = r.Client.Status().Update(ctx, &grafana)
			if err != nil {
				return err
			}
		}
	}

	if statusUpdates > 0 {
		log.Info("successfully synced grafana statuses", "update count", statusUpdates)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.Grafana{}, builder.WithPredicates(ignoreStatusUpdates())).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(ignoreStatusUpdates())).
		Owns(&corev1.ConfigMap{}).
		WithOptions(controller.Options{RateLimiter: defaultRateLimiter()}).
		Complete(r)
	if err != nil {
		return err
	}

	go func() {
		// Wait with sync until elected as leader
		select {
		case <-ctx.Done():
			return
		case <-mgr.Elected():
		}

		// periodic sync reconcile
		log := logf.FromContext(ctx).WithName("GrafanaReconciler")

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(initialSyncDelay):
				start := time.Now()
				err := r.syncStatuses(ctx)
				elapsed := time.Since(start).Milliseconds()
				metrics.InitialStatusSyncDuration.Set(float64(elapsed))

				if err != nil {
					log.Error(err, "error synchronizing grafana statuses")
					continue
				}

				log.Info("Grafana status sync complete")

				return
			}
		}
	}()

	return nil
}

func getInstallationStages() []grafanav1beta1.OperatorStageName {
	return []grafanav1beta1.OperatorStageName{
		grafanav1beta1.OperatorStageAdminUser,
		grafanav1beta1.OperatorStageGrafanaConfig,
		grafanav1beta1.OperatorStagePvc,
		grafanav1beta1.OperatorStageServiceAccount,
		grafanav1beta1.OperatorStageService,
		grafanav1beta1.OperatorStageIngress,
		grafanav1beta1.OperatorStageHTTPRoute,
		grafanav1beta1.OperatorStagePlugins,
		grafanav1beta1.OperatorStageDeployment,
		grafanav1beta1.OperatorStageComplete,
	}
}

func (r *GrafanaReconciler) getReconcilerForStage(stage grafanav1beta1.OperatorStageName) reconcilers.OperatorGrafanaReconciler {
	switch stage {
	case grafanav1beta1.OperatorStageGrafanaConfig:
		return grafana.NewConfigReconciler(r.Client)
	case grafanav1beta1.OperatorStageAdminUser:
		return grafana.NewAdminSecretReconciler(r.Client)
	case grafanav1beta1.OperatorStagePvc:
		return grafana.NewPvcReconciler(r.Client)
	case grafanav1beta1.OperatorStageServiceAccount:
		return grafana.NewServiceAccountReconciler(r.Client)
	case grafanav1beta1.OperatorStageHTTPRoute:
		return grafana.NewHTTPRouteReconciler(r.Client)
	case grafanav1beta1.OperatorStageService:
		return grafana.NewServiceReconciler(r.Client, r.ClusterDomain)
	case grafanav1beta1.OperatorStageIngress:
		return grafana.NewIngressReconciler(r.Client, r.IsOpenShift)
	case grafanav1beta1.OperatorStagePlugins:
		return grafana.NewPluginsReconciler(r.Client)
	case grafanav1beta1.OperatorStageDeployment:
		return grafana.NewDeploymentReconciler(r.Client, r.IsOpenShift)
	case grafanav1beta1.OperatorStageComplete:
		return grafana.NewCompleteReconciler(r.Client)
	default:
		return nil
	}
}

func isImageSHA256(image string) bool {
	return strings.Contains(image, "@sha256:")
}
