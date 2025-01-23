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
	"time"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *GrafanaAlertRuleGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaAlertRuleGroupReconciler")
	logf.IntoContext(ctx, log)

	group := &grafanav1beta1.GrafanaAlertRuleGroup{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, group)
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

	defer func() {
		group.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Client.Status().Update(ctx, group); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(group.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, group); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, group); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	instances, err := GetScopedMatchingInstances(ctx, r.Client, group)
	if err != nil {
		setNoMatchingInstancesCondition(&group.Status.Conditions, group.Generation, err)
		meta.RemoveStatusCondition(&group.Status.Conditions, conditionAlertGroupSynchronized)
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&group.Status.Conditions, group.Generation, err)
		meta.RemoveStatusCondition(&group.Status.Conditions, conditionAlertGroupSynchronized)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&group.Status.Conditions)
	log.Info("found matching Grafana instances for group", "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, group)
	if err != nil || folderUID == "" {
		return ctrl.Result{}, fmt.Errorf("folder uid not found: %w", err)
	}

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		// can be removed in go 1.22+
		grafana := grafana

		err := r.reconcileWithInstance(ctx, &grafana, group, folderUID)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	condition := buildSynchronizedCondition("Alert Rule Group", conditionAlertGroupSynchronized, group.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&group.Status.Conditions, condition)

	return ctrl.Result{RequeueAfter: group.Spec.ResyncPeriod.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaAlertRuleGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaAlertRuleGroup{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

func (r *GrafanaAlertRuleGroupReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, group *grafanav1beta1.GrafanaAlertRuleGroup, folderUID string) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	trueRef := "true" //nolint:goconst
	editable := true
	if group.Spec.Editable != nil && !*group.Spec.Editable {
		editable = false
	}

	_, err = cl.Folders.GetFolderByUID(folderUID) //nolint:errcheck
	if err != nil {
		var folderNotFound *folders.GetFolderByUIDNotFound
		if errors.As(err, &folderNotFound) {
			return fmt.Errorf("folder with uid %s not found", folderUID)
		}
		return fmt.Errorf("fetching folder: %w", err)
	}

	groupName := group.GroupName()
	applied, err := cl.Provisioning.GetAlertRuleGroup(groupName, folderUID)
	var ruleNotFound *provisioning.GetAlertRuleGroupNotFound
	if err != nil && !errors.As(err, &ruleNotFound) {
		return fmt.Errorf("fetching existing alert rule group: %w", err)
	}

	currentRules := make(map[string]bool)
	if applied != nil {
		for _, rule := range applied.Payload.Rules {
			currentRules[rule.UID] = false
		}
	}

	for _, rule := range group.Spec.Rules {
		rule := rule
		apiRule := &models.ProvisionedAlertRule{
			Annotations:  rule.Annotations,
			Condition:    &rule.Condition,
			Data:         make([]*models.AlertQuery, len(rule.Data)),
			ExecErrState: &rule.ExecErrState,
			FolderUID:    &folderUID,
			For:          (*strfmt.Duration)(&rule.For.Duration),
			IsPaused:     rule.IsPaused,
			Labels:       rule.Labels,
			NoDataState:  rule.NoDataState,
			RuleGroup:    &groupName,
			Title:        &rule.Title,
			UID:          rule.UID,
		}
		if rule.NotificationSettings != nil {
			apiRule.NotificationSettings = &models.AlertRuleNotificationSettings{
				Receiver:       &rule.NotificationSettings.Receiver,
				GroupBy:        rule.NotificationSettings.GroupBy,
				GroupWait:      rule.NotificationSettings.GroupWait,
				GroupInterval:  rule.NotificationSettings.GroupInterval,
				RepeatInterval: rule.NotificationSettings.RepeatInterval,
			}
		}
		for idx, q := range rule.Data {
			apiRule.Data[idx] = &models.AlertQuery{
				DatasourceUID:     q.DatasourceUID,
				Model:             q.Model,
				QueryType:         q.QueryType,
				RefID:             q.RefID,
				RelativeTimeRange: q.RelativeTimeRange,
			}
		}

		if _, ok := currentRules[rule.UID]; ok {
			params := provisioning.NewPutAlertRuleParams().
				WithBody(apiRule).
				WithUID(rule.UID)
			if editable {
				params.SetXDisableProvenance(&trueRef)
			}
			_, err := cl.Provisioning.PutAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("updating rule: %w", err)
			}
		} else {
			params := provisioning.NewPostAlertRuleParams().
				WithBody(apiRule)
			if editable {
				params.SetXDisableProvenance(&trueRef)
			}
			_, err = cl.Provisioning.PostAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("creating rule: %w", err)
			}
		}

		currentRules[rule.UID] = true
	}

	for uid, present := range currentRules {
		if !present {
			params := provisioning.NewDeleteAlertRuleParams().
				WithUID(uid)
			if editable {
				params.SetXDisableProvenance(&trueRef)
			}
			_, err := cl.Provisioning.DeleteAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("deleting old alert rule %s: %w", uid, err)
			}
		}
	}

	mGroup := &models.AlertRuleGroup{
		FolderUID: folderUID,
		Interval:  int64(group.Spec.Interval.Seconds()),
		Rules:     []*models.ProvisionedAlertRule{},
		Title:     "",
	}
	params := provisioning.NewPutAlertRuleGroupParams().
		WithBody(mGroup).
		WithGroup(groupName).
		WithFolderUID(folderUID)
	if editable {
		params.SetXDisableProvenance(&trueRef)
	}
	_, err = cl.Provisioning.PutAlertRuleGroup(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("updating group: %s", err.Error())
	}
	return nil
}

func (r *GrafanaAlertRuleGroupReconciler) finalize(ctx context.Context, group *grafanav1beta1.GrafanaAlertRuleGroup) error {
	log := logf.FromContext(ctx)
	folderUID, err := getFolderUID(ctx, r.Client, group)
	if err != nil {
		log.Info("ignoring finalization logic as folder no longer exists")
		return nil //nolint:nilerr
	}

	instances, err := GetScopedMatchingInstances(ctx, r.Client, group)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, i := range instances {
		instance := i
		if err := r.removeFromInstance(ctx, &instance, group, folderUID); err != nil {
			return fmt.Errorf("removing from instance")
		}
	}
	return nil
}

func (r *GrafanaAlertRuleGroupReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, group *grafanav1beta1.GrafanaAlertRuleGroup, folderUID string) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}
	remote, err := cl.Provisioning.GetAlertRuleGroup(group.GroupName(), folderUID)
	if err != nil {
		var notFound *provisioning.GetAlertRuleGroupNotFound
		if errors.As(err, &notFound) {
			// nothing to do
			return nil
		}
		return fmt.Errorf("fetching alert rule group from instance %s: %w", instance.Status.AdminUrl, err)
	}
	for _, rule := range remote.Payload.Rules {
		rule := rule
		params := provisioning.NewDeleteAlertRuleParams().WithUID(rule.UID)
		_, err := cl.Provisioning.DeleteAlertRule(params) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("deleting alert rule %s: %w", rule.UID, err)
		}
	}
	return nil
}
