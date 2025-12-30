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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
	conditionPrometheusRuleGroupSynchronized = "PrometheusRuleGroupSynchronized"
	conditionReasonInvalidPrometheusDuration = "InvalidDuration"
)

// GrafanaPrometheusRuleGroupReconciler reconciles a GrafanaPrometheusRuleGroup object
type GrafanaPrometheusRuleGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *GrafanaPrometheusRuleGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaPrometheusRuleGroupReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaPrometheusRuleGroup{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting GrafanaPrometheusRuleGroup: %w", err)
	}

	if cr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaPrometheusRuleGroup: %w", err)
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
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionPrometheusRuleGroupSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionPrometheusRuleGroupSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.Info("found matching Grafana instances for prometheus rule group", "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf(ErrFetchingFolder, err)
	}

	if folderUID == "" {
		return ctrl.Result{}, fmt.Errorf("folder uid not found, prometheus rule group must reference a folder")
	}

	var disableProvenance *string

	if cr.Spec.Editable != nil && *cr.Spec.Editable {
		dp := disableProvenanceTrue
		disableProvenance = &dp
	}

	mGroup, err := prometheusRuleToModel(cr, folderUID)
	if err != nil {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, conditionReasonInvalidPrometheusDuration, err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionPrometheusRuleGroupSynchronized)

		return ctrl.Result{}, fmt.Errorf("converting prometheus rule group to model: %w", err)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr, &mGroup, disableProvenance)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Prometheus Rule Group", conditionPrometheusRuleGroupSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

// prometheusRuleToModel converts a GrafanaPrometheusRuleGroup to a Grafana AlertRuleGroup model
func prometheusRuleToModel(cr *v1beta1.GrafanaPrometheusRuleGroup, folderUID string) (models.AlertRuleGroup, error) {
	groupName := cr.GroupName()

	mRules := make(models.ProvisionedAlertRules, 0, len(cr.Spec.Rules))

	for idx, r := range cr.Spec.Rules {
		// Skip recording rules - they are not supported in Grafana alerting
		if r.Record != "" {
			continue
		}

		// Skip rules without an alert name
		if r.Alert == "" {
			continue
		}

		// Generate a unique UID for the alert rule based on the CR and rule index
		uid := generateRuleUID(cr.Namespace, cr.Name, r.Alert, idx)

		// Build the query model for Prometheus
		model, err := buildPrometheusQueryModel(r.Expr, cr.Spec.DatasourceUID)
		if err != nil {
			return models.AlertRuleGroup{}, fmt.Errorf("building query model for rule %s: %w", r.Alert, err)
		}

		// Build the reduce expression model for condition evaluation
		reduceModel, err := buildReduceExpressionModel()
		if err != nil {
			return models.AlertRuleGroup{}, fmt.Errorf("building reduce model for rule %s: %w", r.Alert, err)
		}

		// Build the threshold expression model
		thresholdModel, err := buildThresholdExpressionModel()
		if err != nil {
			return models.AlertRuleGroup{}, fmt.Errorf("building threshold model for rule %s: %w", r.Alert, err)
		}

		condition := "C"
		execErrState := "Error"
		noDataState := "NoData"

		apiRule := &models.ProvisionedAlertRule{
			Annotations:  r.Annotations,
			Condition:    &condition,
			ExecErrState: &execErrState,
			FolderUID:    &folderUID,
			IsPaused:     false,
			Labels:       r.Labels,
			NoDataState:  &noDataState,
			RuleGroup:    &groupName,
			Title:        &r.Alert,
			UID:          uid,
			Data: []*models.AlertQuery{
				{
					DatasourceUID: cr.Spec.DatasourceUID,
					Model:         model,
					QueryType:     "",
					RefID:         "A",
					RelativeTimeRange: &models.RelativeTimeRange{
						From: 600,
						To:   0,
					},
				},
				{
					DatasourceUID: "__expr__",
					Model:         reduceModel,
					QueryType:     "",
					RefID:         "B",
					RelativeTimeRange: &models.RelativeTimeRange{
						From: 0,
						To:   0,
					},
				},
				{
					DatasourceUID: "__expr__",
					Model:         thresholdModel,
					QueryType:     "",
					RefID:         "C",
					RelativeTimeRange: &models.RelativeTimeRange{
						From: 0,
						To:   0,
					},
				},
			},
		}

		// Handle 'for' duration
		if r.For != "" {
			duration, err := gtime.ParseDuration(r.For)
			if err != nil {
				return models.AlertRuleGroup{}, fmt.Errorf("invalid 'for' duration %s: %w", r.For, err)
			}

			result := strfmt.Duration(duration)
			apiRule.For = &result
		} else {
			// Default to 0s if not specified
			zeroDuration := strfmt.Duration(0)
			apiRule.For = &zeroDuration
		}

		// Handle keep_firing_for duration
		if r.KeepFiringFor != "" {
			duration, err := gtime.ParseDuration(r.KeepFiringFor)
			if err != nil {
				return models.AlertRuleGroup{}, fmt.Errorf("invalid 'keep_firing_for' duration %s: %w", r.KeepFiringFor, err)
			}

			apiRule.KeepFiringFor = strfmt.Duration(duration)
		}

		mRules = append(mRules, apiRule)
	}

	if len(mRules) == 0 {
		return models.AlertRuleGroup{}, fmt.Errorf("no valid alerting rules found (recording rules are not supported)")
	}

	modelAlertGroup := models.AlertRuleGroup{
		FolderUID: folderUID,
		Interval:  int64(cr.Spec.Interval.Seconds()),
		Rules:     mRules,
		Title:     groupName,
	}

	return modelAlertGroup, nil
}

// generateRuleUID creates a deterministic UID for an alert rule
func generateRuleUID(namespace, name, alertName string, idx int) string {
	input := fmt.Sprintf("%s-%s-%s-%d", namespace, name, alertName, idx)
	hash := sha256.Sum256([]byte(input))
	// Take first 20 bytes and convert to hex (40 chars)
	uid := fmt.Sprintf("%x", hash[:20])

	return uid
}

// buildPrometheusQueryModel creates the JSON model for a Prometheus query
func buildPrometheusQueryModel(expr, datasourceUID string) (*apiextensionsv1.JSON, error) {
	model := map[string]any{
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  datasourceUID,
		},
		"editorMode":    "code",
		"expr":          expr,
		"instant":       true,
		"intervalMs":    1000,
		"legendFormat":  "__auto",
		"maxDataPoints": 43200,
		"range":         false,
		"refId":         "A",
	}

	jsonBytes, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	return &apiextensionsv1.JSON{Raw: jsonBytes}, nil
}

// buildReduceExpressionModel creates the JSON model for a reduce expression
func buildReduceExpressionModel() (*apiextensionsv1.JSON, error) {
	model := map[string]any{
		"conditions": []map[string]any{
			{
				"evaluator": map[string]any{
					"params": []any{},
					"type":   "gt",
				},
				"operator": map[string]any{
					"type": "and",
				},
				"query": map[string]any{
					"params": []string{"B"},
				},
				"reducer": map[string]any{
					"params": []any{},
					"type":   "last",
				},
				"type": "query",
			},
		},
		"datasource": map[string]string{
			"type": "__expr__",
			"uid":  "__expr__",
		},
		"expression":  "A",
		"reducer":     "last",
		"refId":       "B",
		"type":        "reduce",
		"downsampler": "last",
	}

	jsonBytes, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	return &apiextensionsv1.JSON{Raw: jsonBytes}, nil
}

// buildThresholdExpressionModel creates the JSON model for a threshold expression
func buildThresholdExpressionModel() (*apiextensionsv1.JSON, error) {
	model := map[string]any{
		"conditions": []map[string]any{
			{
				"evaluator": map[string]any{
					"params": []any{0},
					"type":   "gt",
				},
				"operator": map[string]any{
					"type": "and",
				},
				"query": map[string]any{
					"params": []string{"C"},
				},
				"reducer": map[string]any{
					"params": []any{},
					"type":   "last",
				},
				"type": "query",
			},
		},
		"datasource": map[string]string{
			"type": "__expr__",
			"uid":  "__expr__",
		},
		"expression": "B",
		"refId":      "C",
		"type":       "threshold",
	}

	jsonBytes, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	return &apiextensionsv1.JSON{Raw: jsonBytes}, nil
}

func (r *GrafanaPrometheusRuleGroupReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaPrometheusRuleGroup, mGroup *models.AlertRuleGroup, disableProvenance *string) error {
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
				WithXDisableProvenance(disableProvenance)

			_, err = gClient.Provisioning.PostAlertRule(params) //nolint:errcheck
			if err != nil {
				// If rule creation fails due to conflict (already exists with different UID),
				// try to update the group
				if strings.Contains(err.Error(), "conflict") {
					continue
				}

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
		WithXDisableProvenance(disableProvenance)

	_, err = gClient.Provisioning.PutAlertRuleGroup(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("updating group: %s", err.Error())
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaPrometheusRuleGroupReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaPrometheusRuleGroup) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaPrometheusRuleGroup")

	isCleanupInGrafanaRequired := true

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		log.Info("Skipping Grafana finalize logic as folder no longer exists")

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
func (r *GrafanaPrometheusRuleGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaPrometheusRuleGroup{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
