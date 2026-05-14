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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
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

// NewGenericPrometheusRuleReconciler builds the generic reconciler instance for
// GrafanaPrometheusRule. The reconcile loop, instance matching, dynamic-client
// dispatch, finalization and condition wiring all live in GenericReconciler;
// this only supplies the resource-specific conversion and validation hooks.
func NewGenericPrometheusRuleReconciler(cl client.Client, cfg *Config) *GenericReconciler[v1beta1.GrafanaPrometheusRule, *v1beta1.GrafanaPrometheusRule] {
	return &GenericReconciler[v1beta1.GrafanaPrometheusRule, *v1beta1.GrafanaPrometheusRule]{
		Client:       cl,
		Cfg:          cfg,
		ResourceName: "PrometheusRule",
		GVR:          prometheusRuleGVR,
		Convert: func(ctx context.Context, cl client.Client, cr *v1beta1.GrafanaPrometheusRule) (runtime.Object, error) {
			folderUID, err := getFolderUID(ctx, cl, cr)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", LogMsgResolvingFolderUID, err)
			}

			return crToUnstructuredPrometheusRule(cr, folderUID), nil
		},
		Validate: func(cr *v1beta1.GrafanaPrometheusRule) *ValidationError {
			// CRD-level XValidation covers folder/folderRef immutability and
			// mutual exclusion; the upstream PrometheusRule kind validates the
			// rest of the body server-side.
			return nil
		},
		SynchronizedCondition: conditionPrometheusRuleSynchronized,
	}
}

// crToUnstructuredPrometheusRule builds the wire-format object the upstream
// app-platform kind expects: multi-group Prometheus-shape body with folder
// and datasource UID carried on annotations. The k8s metadata.namespace is
// deliberately left unset so the DynamicClient falls back to the per-instance
// namespace (Spec.External.TenantNamespace, default "default").
func crToUnstructuredPrometheusRule(cr *v1beta1.GrafanaPrometheusRule, folderUID string) *unstructured.Unstructured {
	groups := make([]any, 0, len(cr.Spec.Groups))
	for _, g := range cr.Spec.Groups {
		entry := map[string]any{
			"name":  g.Name,
			"rules": rulesToAny(g.Rules),
		}
		if g.Interval != nil {
			entry["interval"] = g.Interval.Duration.String()
		}

		if g.QueryOffset != nil {
			entry["queryOffset"] = g.QueryOffset.Duration.String()
		}

		if g.Limit != nil {
			entry["limit"] = *g.Limit
		}

		if len(g.Labels) > 0 {
			entry["labels"] = anyMap(g.Labels)
		}

		groups = append(groups, entry)
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
	// Use the CR's k8s UID as the upstream resource name so the generic
	// reconciler's finalize, which deletes by cr.GetUID(), targets the same
	// object we created here.
	obj.SetName(string(cr.GetUID()))

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

func anyMap(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}

	return out
}
