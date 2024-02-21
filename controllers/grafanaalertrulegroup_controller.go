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
	"strings"
	"time"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

// GrafanaAlertRuleGroupReconciler reconciles a GrafanaAlertRuleGroup object
type GrafanaAlertRuleGroupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaalertrulegroups/finalizers,verbs=update

func (r *GrafanaAlertRuleGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx).WithName("GrafanaAlertRuleGroupReconciler")
	r.Log = log.FromContext(ctx)

	group := &grafanav1beta1.GrafanaAlertRuleGroup{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, group)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana folder cr")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if group.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(group, grafanaFinalizer) {
			// still need to clean up
			err := r.finalize(ctx, group)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("cleaning up alert rule group: %w", err)
			}
			controllerutil.RemoveFinalizer(group, grafanaFinalizer)
			if err := r.Update(ctx, group); err != nil {
				r.Log.Error(err, "failed to remove finalizer")
				return ctrl.Result{RequeueAfter: RequeueDelay}, err
			}
		}
		return ctrl.Result{}, nil
	}
	if !controllerutil.ContainsFinalizer(group, grafanaFinalizer) {
		controllerutil.AddFinalizer(group, grafanaFinalizer)
		if err := r.Update(ctx, group); err != nil {
			r.Log.Error(err, "failed to set finalizer")
			return ctrl.Result{RequeueAfter: RequeueDelay}, err
		}
	}

	defer func() {
		if err := r.Client.Status().Update(ctx, group); err != nil {
			r.Log.Error(err, "updating status")
		}
	}()

	instances, err := r.GetMatchingInstances(ctx, group.Spec.InstanceSelector, r.Client)
	if err != nil {
		setNoMatchingInstance(&group.Status.Conditions, group.Generation, "ErrFetchingInstances", fmt.Sprintf("error occurred during fetching of instances: %s", err.Error()))
		r.Log.Error(err, "could not find matching instances")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}
	if len(instances.Items) == 0 {
		setNoMatchingInstance(&group.Status.Conditions, group.Generation, "EmptyAPIReply", "Instances could not be fetched, reconciliation will be retried")
		return ctrl.Result{}, nil
	}

	removeNoMatchingInstance(&group.Status.Conditions)
	folderUID := r.GetFolderUID(ctx, group)
	if folderUID == "" {
		// error is already set in conditions
		return ctrl.Result{}, nil
	}

	applyErrors := make(map[string]string)
	for _, grafana := range instances.Items {
		// can be removed in go 1.22+
		grafana := grafana
		if grafana.Status.Stage != grafanav1beta1.OperatorStageComplete || grafana.Status.StageStatus != grafanav1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		err := r.reconcileWithInstance(ctx, &grafana, group, folderUID)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}
	condition := metav1.Condition{
		Type:               "AlertGroupSynchronized",
		ObservedGeneration: group.Generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	}

	if len(applyErrors) == 0 {
		condition.Status = "True"
		condition.Reason = "ApplySuccesfull"
		condition.Message = fmt.Sprintf("Alert Rule Group was successfully applied to %d instances", len(instances.Items))
	} else {
		condition.Status = "False"
		condition.Reason = "ApplyFailed"

		var sb strings.Builder
		for i, err := range applyErrors {
			sb.WriteString(fmt.Sprintf("\n- %s: %s", i, err))
		}

		condition.Message = fmt.Sprintf("Alert Rule Group failed to be applied for %d out of %d instances. Errors:%s", len(applyErrors), len(instances.Items), sb.String())
	}
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
	strue := "true"

	applied, err := cl.Provisioning.GetAlertRuleGroup(group.Name, folderUID)
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
			RuleGroup:    &group.Name,
			Title:        &rule.Title,
			UID:          rule.UID,
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
				WithXDisableProvenance(&strue).
				WithUID(rule.UID)
			_, err := cl.Provisioning.PutAlertRule(params) //nolint:errcheck
			if err != nil {
				return fmt.Errorf("updating rule: %w", err)
			}
		} else {
			params := provisioning.NewPostAlertRuleParams().
				WithBody(apiRule).
				WithXDisableProvenance(&strue)
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
				WithUID(uid).
				WithXDisableProvenance(&strue)
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
		WithGroup(group.Name).
		WithFolderUID(folderUID).
		WithXDisableProvenance(&strue)
	_, err = cl.Provisioning.PutAlertRuleGroup(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("updating group: %s", err.Error())
	}
	return nil
}

func (r *GrafanaAlertRuleGroupReconciler) finalize(ctx context.Context, group *grafanav1beta1.GrafanaAlertRuleGroup) error {
	folderUID := r.GetFolderUID(ctx, group)
	if folderUID == "" {
		r.Log.Info("ignoring finalization logic as folder no longer exists")
		return nil
	}
	instances, err := r.GetMatchingInstances(ctx, group.Spec.InstanceSelector, r.Client)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances.Items {
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
	remote, err := cl.Provisioning.GetAlertRuleGroup(group.Name, folderUID)
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

func (r *GrafanaAlertRuleGroupReconciler) GetMatchingInstances(ctx context.Context, selector *metav1.LabelSelector, k8sClient client.Client) (grafanav1beta1.GrafanaList, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, selector)
	if err != nil || len(instances.Items) == 0 {
		return grafanav1beta1.GrafanaList{}, err
	}
	return instances, err
}

func (r *GrafanaAlertRuleGroupReconciler) GetMatchingFolders(ctx context.Context, selector *metav1.LabelSelector) (grafanav1beta1.GrafanaFolderList, error) {
	var list grafanav1beta1.GrafanaFolderList
	opts := []client.ListOption{
		client.MatchingLabels(selector.MatchLabels),
	}
	err := r.Client.List(ctx, &list, opts...)
	return list, err
}

func (r *GrafanaAlertRuleGroupReconciler) GetFolderUID(ctx context.Context, group *grafanav1beta1.GrafanaAlertRuleGroup) string {
	if group.Spec.FolderUID != "" {
		return group.Spec.FolderUID
	}
	folders, err := r.GetMatchingFolders(ctx, group.Spec.FolderSelector)
	if err != nil {
		setNoMatchingFolder(&group.Status.Conditions, group.Generation, "ErrFetchingFolders", fmt.Sprintf("Failed to fetch folders: %s", err.Error()))
		return ""
	}
	if len(folders.Items) < 1 {
		setNoMatchingFolder(&group.Status.Conditions, group.Generation, "NoMatches", "Folder selector did not match any folders")
		return ""
	}
	if len(folders.Items) > 1 {
		setNoMatchingFolder(&group.Status.Conditions, group.Generation, "MultipleMatches", fmt.Sprintf("Folder selector matched %d folders. Only one folder can be matched", len(folders.Items)))
		return ""
	}
	return string(folders.Items[0].UID)
}
