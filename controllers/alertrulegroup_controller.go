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

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionAlertGroupSynchronized = "AlertGroupSynchronized"
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

	group := &grafanav1beta1.GrafanaAlertRuleGroup{}

	err := r.Get(ctx, req.NamespacedName, group)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting GrafanaAlertRuleGroup: %w", err)
	}

	if group.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(group, grafanaFinalizer) {
			if err := r.finalize(ctx, group); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaAlertRuleGroup: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, group); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, group)

	if group.Spec.Suspend {
		setSuspended(&group.Status.Conditions, group.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&group.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, group)
	if err != nil {
		setNoMatchingInstancesCondition(&group.Status.Conditions, group.Generation, err)
		meta.RemoveStatusCondition(&group.Status.Conditions, conditionAlertGroupSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&group.Status.Conditions, group.Generation, err)
		meta.RemoveStatusCondition(&group.Status.Conditions, conditionAlertGroupSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&group.Status.Conditions)
	log.Info("found matching Grafana instances for group", "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, group)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf(ErrFetchingFolder, err)
	}

	if folderUID == "" {
		return ctrl.Result{}, fmt.Errorf("folder uid not found, alert rule must reference a folder")
	}

	editable := "true" //nolint:goconst
	if group.Spec.Editable != nil && !*group.Spec.Editable {
		editable = "false"
	}

	mGroup := crToModel(group, folderUID)

	log.V(1).Info("converted cr to api model")

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, group, &mGroup, editable)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Alert Rule Group", conditionAlertGroupSynchronized, group.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&group.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(group.Spec.ResyncPeriod)}, nil
}

func crToModel(cr *grafanav1beta1.GrafanaAlertRuleGroup, folderUID string) models.AlertRuleGroup {
	groupName := cr.GroupName()

	mRules := make(models.ProvisionedAlertRules, 0, len(cr.Spec.Rules))

	for _, r := range cr.Spec.Rules {
		apiRule := &models.ProvisionedAlertRule{
			Annotations:  r.Annotations,
			Condition:    &r.Condition,
			Data:         make([]*models.AlertQuery, len(r.Data)),
			ExecErrState: &r.ExecErrState,
			FolderUID:    &folderUID,
			For:          (*strfmt.Duration)(&r.For.Duration),
			IsPaused:     r.IsPaused,
			Labels:       r.Labels,
			NoDataState:  r.NoDataState,
			RuleGroup:    &groupName,
			Title:        &r.Title,
			UID:          r.UID,
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
			apiRule.KeepFiringFor = (strfmt.Duration)(r.KeepFiringFor.Duration)
		}

		mRules = append(mRules, apiRule)
	}

	return models.AlertRuleGroup{
		FolderUID: folderUID,
		Interval:  int64(cr.Spec.Interval.Seconds()),
		Rules:     mRules,
		Title:     groupName,
	}
}

func (r *GrafanaAlertRuleGroupReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, group *grafanav1beta1.GrafanaAlertRuleGroup, mGroup *models.AlertRuleGroup, disableProvenance string) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	folderUID := mGroup.FolderUID

	_, err = cl.Folders.GetFolderByUID(folderUID) //nolint:errcheck
	if err != nil {
		var folderNotFound *folders.GetFolderByUIDNotFound
		if errors.As(err, &folderNotFound) {
			return fmt.Errorf("folder with uid %s not found", folderUID)
		}

		return fmt.Errorf("fetching folder: %w", err)
	}

	applied, err := cl.Provisioning.GetAlertRuleGroup(mGroup.Title, folderUID)

	var ruleNotFound *provisioning.GetAlertRuleGroupNotFound
	if err != nil && !errors.As(err, &ruleNotFound) {
		return fmt.Errorf("fetching existing alert rule group: %w", err)
	}

	// Create an empty collection to loop over if group does not exist on remote
	remoteRules := models.ProvisionedAlertRules{}
	if applied != nil && applied.Payload != nil {
		remoteRules = applied.Payload.Rules
	}

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
				WithXDisableProvenance(&disableProvenance)

			_, err = cl.Provisioning.PostAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("creating rule: %w", err)
			}
		}
	}

	// Update whole group and all rules existing rules at once
	// Will delete rules not present in the body
	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(mGroup).
		WithGroup(mGroup.Title).
		WithFolderUID(folderUID).
		WithXDisableProvenance(&disableProvenance)

	_, err = cl.Provisioning.PutAlertRuleGroup(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("updating group: %s", err.Error())
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, group, group.NamespacedResource())
}

func (r *GrafanaAlertRuleGroupReconciler) finalize(ctx context.Context, group *grafanav1beta1.GrafanaAlertRuleGroup) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaAlertRuleGroup")

	isCleanupInGrafanaRequired := true

	folderUID, err := getFolderUID(ctx, r.Client, group)
	if err != nil {
		log.Info("Skipping Grafana finalize logic as folder no longer exists")

		isCleanupInGrafanaRequired = false
	}

	instances, err := GetScopedMatchingInstances(ctx, r.Client, group)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		// Skip cleanup in instances
		if isCleanupInGrafanaRequired {
			cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &instance)
			if err != nil {
				return fmt.Errorf("building grafana client: %w", err)
			}

			_, err = cl.Provisioning.DeleteAlertRuleGroup(group.GroupName(), folderUID) //nolint:errcheck
			if err != nil {
				var notFound *provisioning.DeleteAlertRuleGroupNotFound
				if !errors.As(err, &notFound) {
					return fmt.Errorf("deleting alert rule group: %w", err)
				}
			}
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, group)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaAlertRuleGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaAlertRuleGroup{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
