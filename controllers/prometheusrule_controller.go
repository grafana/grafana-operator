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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionPrometheusRuleSynchronized = "PrometheusRuleSynchronized"

	prometheusRuleAPIGroup    = "rules.alerting.grafana.app"
	prometheusRuleAPIVersion  = "v0alpha1"
	prometheusRuleAPIResource = "prometheusrules"

	prometheusRuleFolderAnnotation     = "grafana.app/folder"
	prometheusRuleDatasourceAnnotation = "rules.alerting.grafana.app/datasource-uid"
)

var prometheusRuleGVR = schema.GroupVersionResource{
	Group:    prometheusRuleAPIGroup,
	Version:  prometheusRuleAPIVersion,
	Resource: prometheusRuleAPIResource,
}

// GrafanaPrometheusRuleReconciler reconciles a GrafanaPrometheusRule object.
type GrafanaPrometheusRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *GrafanaPrometheusRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaPrometheusRuleReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaPrometheusRule{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, LogMsgGettingCR)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingCR, err)
	}

	if cr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				log.Error(err, LogMsgRunningFinalizer)
				return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgRunningFinalizer, err)
			}

			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				log.Error(err, LogMsgRemoveFinalizer)
				return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgRemoveFinalizer, err)
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
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionPrometheusRuleSynchronized)
		log.Error(err, LogMsgGettingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionPrometheusRuleSynchronized)
		log.Error(ErrNoMatchingInstances, LogMsgNoMatchingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgNoMatchingInstances, ErrNoMatchingInstances)
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.V(1).Info(DbgMsgFoundMatchingInstances, "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgResolvingFolderUID)
		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgResolvingFolderUID, err)
	}

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		if err := r.reconcileWithInstance(ctx, &grafana, cr, folderUID); err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Prometheus Rule", conditionPrometheusRuleSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		err = fmt.Errorf(FmtStrApplyErrors, applyErrors)
		log.Error(err, LogMsgApplyErrors)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgApplyErrors, err)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

// crToUnstructured builds the wire-format object the upstream app-platform kind
// expects: multi-group Prometheus-shape body, folder/datasource via annotations.
// The k8s metadata.namespace is filled in by the caller once the target Grafana
// instance's namespace has been resolved.
func (r *GrafanaPrometheusRuleReconciler) crToUnstructured(cr *v1beta1.GrafanaPrometheusRule, folderUID, namespace string) *unstructured.Unstructured {
	groups := make([]any, 0, len(cr.Spec.Groups))
	for _, g := range cr.Spec.Groups {
		groupEntry := map[string]any{
			"name":  g.Name,
			"rules": rulesToAny(g.Rules),
		}
		if g.Interval != nil {
			groupEntry["interval"] = g.Interval.Duration.String()
		}

		if g.QueryOffset != nil {
			groupEntry["queryOffset"] = g.QueryOffset.Duration.String()
		}

		if g.Limit != nil {
			groupEntry["limit"] = *g.Limit
		}

		if len(g.Labels) > 0 {
			groupEntry["labels"] = anyMap(g.Labels)
		}

		groups = append(groups, groupEntry)
	}

	annotations := map[string]string{}
	if folderUID != "" {
		annotations[prometheusRuleFolderAnnotation] = folderUID
	}

	if cr.Spec.DatasourceUID != "" {
		annotations[prometheusRuleDatasourceAnnotation] = cr.Spec.DatasourceUID
	}

	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(prometheusRuleAPIGroup + "/" + prometheusRuleAPIVersion)
	obj.SetKind("PrometheusRule")
	obj.SetName(cr.Name)
	obj.SetNamespace(namespace)

	if len(annotations) > 0 {
		obj.SetAnnotations(annotations)
	}

	obj.Object["spec"] = map[string]any{"groups": groups}

	return obj
}

func rulesToAny(rules []v1beta1.PrometheusRule) []any {
	out := make([]any, 0, len(rules))
	for _, rule := range rules {
		entry := map[string]any{"expr": rule.Expr}
		if rule.Alert != "" {
			entry["alert"] = rule.Alert
		}

		if rule.Record != "" {
			entry["record"] = rule.Record
		}

		if rule.For != nil {
			entry["for"] = rule.For.Duration.String()
		}

		if rule.KeepFiringFor != nil {
			entry["keepFiringFor"] = rule.KeepFiringFor.Duration.String()
		}

		if len(rule.Labels) > 0 {
			entry["labels"] = anyMap(rule.Labels)
		}

		if len(rule.Annotations) > 0 {
			entry["annotations"] = anyMap(rule.Annotations)
		}

		out = append(out, entry)
	}

	return out
}

func (r *GrafanaPrometheusRuleReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaPrometheusRule, folderUID string) error {
	log := logf.FromContext(ctx)

	dyn, _, err := grafanaclient.NewDynamicClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana dynamic client: %w", err)
	}

	namespace, err := grafanaclient.ResolveNamespace(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("resolving target namespace: %w", err)
	}

	desired := r.crToUnstructured(cr, folderUID, namespace)
	resourceClient := dyn.Resource(prometheusRuleGVR).Namespace(namespace)
	name := cr.Name

	existing, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		log.Info("creating prometheus rule", "name", name)

		if _, err := resourceClient.Create(ctx, desired, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("creating prometheus rule: %w", err)
		}
	case err != nil:
		return fmt.Errorf("fetching existing prometheus rule: %w", err)
	default:
		if specsEqual(existing, desired) {
			log.V(1).Info("prometheus rule already in sync, skipping update")
			return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
		}

		log.Info("updating prometheus rule", "name", name)
		desired.SetResourceVersion(existing.GetResourceVersion())

		if _, err := resourceClient.Update(ctx, desired, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("updating prometheus rule: %w", err)
		}
	}

	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaPrometheusRuleReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaPrometheusRule) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaPrometheusRule")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgGettingInstances)
		return fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	for _, instance := range instances {
		dyn, _, err := grafanaclient.NewDynamicClient(ctx, r.Client, &instance)
		if err != nil {
			return fmt.Errorf("building grafana dynamic client: %w", err)
		}

		namespace, err := grafanaclient.ResolveNamespace(ctx, r.Client, &instance)
		if err != nil {
			return fmt.Errorf("resolving target namespace: %w", err)
		}

		err = dyn.Resource(prometheusRuleGVR).Namespace(namespace).
			Delete(ctx, cr.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("deleting prometheus rule: %w", err)
		}

		if err := instance.RemoveNamespacedResource(ctx, r.Client, cr); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaPrometheusRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaPrometheusRule{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

// specsEqual compares only the spec subtree, ignoring server-set metadata fields.
func specsEqual(a, b *unstructured.Unstructured) bool {
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a.Object["spec"], b.Object["spec"])
}

func anyMap(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}

	return out
}
