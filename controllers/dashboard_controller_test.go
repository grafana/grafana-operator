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
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Dashboard Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaDashboardSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecSuspended,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecNoMatchingInstances,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecApplyFailed,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionDashboardSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Invalid JSON",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecInvalidSpec,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{]"}, // Invalid json
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidModelResolution,
			},
			wantErr: "resolving dashboard contents",
		},
		{
			name: "No model can be resolved, no model source is defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "invalid-spec-no-model-source",
			},
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidModelResolution,
			},
			wantErr: "resolving dashboard contents",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JSON: `{
							"title": "Minimal Dashboard",
							"links": []
						}`,
				},
			},
			want: metav1.Condition{
				Type:   conditionDashboardSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaDashboard{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaDashboardReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
