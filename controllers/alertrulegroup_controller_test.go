package controllers

import (
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("AlertRulegroup Reconciler: Provoke Conditions", func() {
	t := GinkgoT()

	noDataState := "NoData"
	rules := []v1beta1.AlertRule{
		{
			Title:        "TestRule",
			UID:          "akdj-wonvo",
			ExecErrState: "KeepLast",
			NoDataState:  &noDataState,
			Data:         []*v1beta1.AlertQuery{},
		},
	}

	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaAlertRuleGroupSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				FolderUID:         "GroupUID",
				Rules:             rules,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				FolderUID:         "GroupUID",
				Rules:             rules,
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
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				FolderRef:         "pre-existing",
				Rules:             rules,
			},
			want: metav1.Condition{
				Type:   conditionAlertGroupSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				FolderRef:         "pre-existing",
				Interval:          metav1.Duration{Duration: 60 * time.Second},
				Rules: []v1beta1.AlertRule{
					{
						Title:     "MathRule",
						UID:       "oefiodwa-dam-dwa",
						Condition: "A",
						Data: []*v1beta1.AlertQuery{
							{
								RefID:             "A",
								RelativeTimeRange: nil,
								DatasourceUID:     "__expr__",
								Model: &v1.JSON{Raw: []byte(`{
		                                "conditions": [
		                                    {
		                                        "evaluator": {
		                                            "params": [
		                                                0,
		                                                0
		                                            ],
		                                            "type": "gt"
		                                        },
		                                        "operator": {
		                                            "type": "and"
		                                        },
		                                        "query": {
		                                            "params": []
		                                        },
		                                        "reducer": {
		                                            "params": [],
		                                            "type": "avg"
		                                        },
		                                        "type": "query"
		                                    }
		                                ],
		                                "datasource": {
		                                    "name": "Expression",
		                                    "type": "__expr__",
		                                    "uid": "__expr__"
		                                },
		                                "expression": "1 > 0",
		                                "hide": false,
		                                "intervalMs": 1000,
		                                "maxDataPoints": 100,
		                                "refId": "B",
		                                "type": "math"
		                            }`)},
							},
						},
						NoDataState:  &noDataState,
						ExecErrState: "Error",
						For:          &metav1.Duration{Duration: 60 * time.Second},
						Annotations:  map[string]string{},
						Labels:       map[string]string{},
						IsPaused:     true,
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionAlertGroupSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			err := k8sClient.Create(testCtx, cr)
			require.NoError(t, err)

			r := GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			req := requestFromMeta(tt.meta)

			// Reconcile

			_, err = r.Reconcile(testCtx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}

			cr = &v1beta1.GrafanaAlertRuleGroup{}

			err = r.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			containsEqualCondition(cr.Status.Conditions, tt.want)
		})
	}
})
