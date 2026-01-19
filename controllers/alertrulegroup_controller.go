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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend/gtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionAlertGroupSynchronized = "AlertGroupSynchronized"
	conditionReasonInvalidDuration  = "InvalidDuration"

	ErrMsgMissingFolderReference = "folder uid not found, AlertRuleGroup must include a folder reference (folderUID/folderRef)"
)

// GrafanaAlertRuleGroupReconciler reconciles a GrafanaAlertRuleGroup object
type GrafanaAlertRuleGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *GrafanaAlertRuleGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaAlertRuleGroupReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaAlertRuleGroup{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting GrafanaAlertRuleGroup: %w", err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaAlertRuleGroup: %w", err)
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

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionAlertGroupSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionAlertGroupSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.Info("found matching Grafana instances for group", "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, ErrMsgResolvingFolderUID)
		return ctrl.Result{}, fmt.Errorf("%s: %w", ErrMsgResolvingFolderUID, err)
	}

	if folderUID == "" {
		log.Error(err, ErrMsgMissingFolderReference)
		return ctrl.Result{}, fmt.Errorf("%s: %w", ErrMsgMissingFolderReference, err)
	}

	var disableProvenance *string

	if cr.Spec.Editable != nil && *cr.Spec.Editable {
		trueStr := "true"
		disableProvenance = &trueStr
	}

	mGroup, err := crToModel(cr, folderUID)
	if err != nil {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, conditionReasonInvalidDuration, err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionAlertGroupSynchronized)

		return ctrl.Result{}, fmt.Errorf("converting alert rule group to model: %w", err)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr, &mGroup, disableProvenance)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Alert Rule Group", conditionAlertGroupSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func crToModel(cr *v1beta1.GrafanaAlertRuleGroup, folderUID string) (models.AlertRuleGroup, error) {
	groupName := cr.GroupName()

	mRules := make(models.ProvisionedAlertRules, 0, len(cr.Spec.Rules))

	for _, r := range cr.Spec.Rules {
		apiRule := &models.ProvisionedAlertRule{
			Annotations:  r.Annotations,
			Condition:    &r.Condition,
			Data:         make([]*models.AlertQuery, len(r.Data)),
			ExecErrState: &r.ExecErrState,
			FolderUID:    &folderUID,
			IsPaused:     r.IsPaused,
			Labels:       r.Labels,
			NoDataState:  r.NoDataState,
			RuleGroup:    &groupName,
			Title:        &r.Title,
			UID:          r.UID,
		}

		if r.For != nil {
			duration, err := gtime.ParseDuration(*r.For)
			if err != nil {
				return models.AlertRuleGroup{}, fmt.Errorf("invalid 'for' duration %s: %w", *r.For, err)
			}

			result := strfmt.Duration(duration)
			apiRule.For = &result
		}

		if r.NotificationSettings != nil {
			apiRule.NotificationSettings = &models.AlertRuleNotificationSettings{
				Receiver:          &r.NotificationSettings.Receiver,
				GroupBy:           r.NotificationSettings.GroupBy,
				GroupWait:         r.NotificationSettings.GroupWait,
				MuteTimeIntervals: r.NotificationSettings.MuteTimeIntervals,
				GroupInterval:     r.NotificationSettings.GroupInterval,
				RepeatInterval:    r.NotificationSettings.RepeatInterval,
			}
		}

		if r.Record != nil {
			apiRule.Record = &models.Record{
				From:   &r.Record.From,
				Metric: &r.Record.Metric,
			}
		}

		if r.MissingSeriesEvalsToResolve != nil {
			apiRule.MissingSeriesEvalsToResolve = *r.MissingSeriesEvalsToResolve
		}

		for idx, q := range r.Data {
			apiRule.Data[idx] = &models.AlertQuery{
				DatasourceUID:     q.DatasourceUID,
				Model:             q.Model,
				QueryType:         q.QueryType,
				RefID:             q.RefID,
				RelativeTimeRange: q.RelativeTimeRange,
			}
		}

		if r.KeepFiringFor != nil {
			apiRule.KeepFiringFor = strfmt.Duration(r.KeepFiringFor.Duration)
		}

		mRules = append(mRules, apiRule)
	}

	modelAlertGroup := models.AlertRuleGroup{
		FolderUID: folderUID,
		Interval:  int64(cr.Spec.Interval.Seconds()),
		Rules:     mRules,
		Title:     groupName,
	}

	return modelAlertGroup, nil
}

func (r *GrafanaAlertRuleGroupReconciler) matchesStateInGrafana(exists bool, model *models.AlertRuleGroup, remoteARG *provisioning.GetAlertRuleGroupOK) bool {
	if !exists {
		return false
	}

	if model == nil || remoteARG == nil {
		return false
	}

	remoteModel := remoteARG.GetPayload()
	if remoteModel == nil {
		return false
	}

	matchesRemoteState := remoteModel.FolderUID == model.FolderUID &&
		remoteModel.Interval == model.Interval &&
		reflect.DeepEqual(remoteModel.Rules, model.Rules) &&
		remoteModel.Title == model.Title

	return matchesRemoteState
}

func (r *GrafanaAlertRuleGroupReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaAlertRuleGroup, mGroup *models.AlertRuleGroup, disableProvenance *string) error {
	log := logf.FromContext(ctx)

	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	folderUID := mGroup.FolderUID

	_, err = gClient.Folders.GetFolderByUID(folderUID) //nolint:errcheck
	if err != nil {
		var folderNotFound *folders.GetFolderByUIDNotFound
		if errors.As(err, &folderNotFound) {
			return fmt.Errorf("folder with uid %s not found", folderUID)
		}

		return fmt.Errorf("fetching folder: %w", err)
	}

	applied, err := gClient.Provisioning.GetAlertRuleGroup(mGroup.Title, folderUID)
	if err != nil {
		var ruleNotFound *provisioning.GetAlertRuleGroupNotFound
		if !errors.As(err, &ruleNotFound) {
			return fmt.Errorf("fetching existing alert rule group: %w", err)
		}
	}

	exists := applied != nil

	matchesStateInGrafana := r.matchesStateInGrafana(exists, mGroup, applied)

	if matchesStateInGrafana {
		log.V(1).Info("alert rule group hasn't changed, skipping update")
		return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
	}

	log.Info("updating alert rule group", "title", mGroup.Title)

	// Create an empty collection to loop over if group does not exist on remote
	remoteRules := models.ProvisionedAlertRules{}
	if applied != nil && applied.Payload != nil {
		remoteRules = applied.Payload.Rules
	}

	// TODO: workaround for < 11.6.0 where a rule group and rules cannot be created at once.
	//       Otherwise, Grafana will fail to calculate rule diff and respond with 500. Deprecate it later.
	// Rules must be created individually
	// Find rules missing on the instance and create them
	for _, mRule := range mGroup.Rules {
		ruleExists := false

		for _, remoteRule := range remoteRules {
			if mRule.UID == remoteRule.UID {
				ruleExists = true
				break
			}
		}

		if !ruleExists {
			params := provisioning.NewPostAlertRuleParams().
				WithBody(mRule).
				WithXDisableProvenance(disableProvenance)

			_, err = gClient.Provisioning.PostAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("creating rule: %w", err)
			}
		}
	}

	// Update whole group and all existing rules at once
	// Will delete rules not present in the body
	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(mGroup).
		WithGroup(mGroup.Title).
		WithFolderUID(folderUID).
		WithXDisableProvenance(disableProvenance)

	_, err = gClient.Provisioning.PutAlertRuleGroup(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("updating alert rule group: %s", err.Error())
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaAlertRuleGroupReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaAlertRuleGroup) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaAlertRuleGroup")

	isCleanupInGrafanaRequired := true

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		log.V(1).Info("Skipping Grafana finalize as folder no longer exists")

		isCleanupInGrafanaRequired = false
	}

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		// Skip cleanup in instances
		if isCleanupInGrafanaRequired {
			gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, &instance)
			if err != nil {
				return fmt.Errorf("building grafana client: %w", err)
			}

			_, err = gClient.Provisioning.DeleteAlertRuleGroup(cr.GroupName(), folderUID) //nolint:errcheck
			if err != nil {
				var notFound *provisioning.DeleteAlertRuleGroupNotFound
				if !errors.As(err, &notFound) {
					return fmt.Errorf("deleting alert rule group: %w", err)
				}
			}
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaAlertRuleGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaAlertRuleGroup{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
