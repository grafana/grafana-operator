package controllers

import (
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlertRulegroup Reconciler: Provoke Conditions", func() {
	noDataState := "NoData"
	durationString := "60s"
	dayDuration := "1d"
	weekDuration := "1w"
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
						For:          &durationString,
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
		{
			name: "Duration conversion with day duration",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				FolderRef:         "pre-existing",
				Interval:          metav1.Duration{Duration: 60 * time.Second},
				Rules: []v1beta1.AlertRule{
					{
						Title:     "DayDurationRule",
						UID:       "day-duration-rule",
						Condition: "A",
						Data: []*v1beta1.AlertQuery{
							{
								RefID:         "A",
								DatasourceUID: "__expr__",
								Model:         &v1.JSON{Raw: []byte(`{"expression": "1", "refId": "A"}`)},
							},
						},
						NoDataState:  &noDataState,
						ExecErrState: "Error",
						For:          &dayDuration, // 1d
						Annotations:  map[string]string{},
						Labels:       map[string]string{},
						IsPaused:     false,
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionAlertGroupSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
		{
			name: "Duration conversion with week duration",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				FolderRef:         "pre-existing",
				Interval:          metav1.Duration{Duration: 60 * time.Second},
				Rules: []v1beta1.AlertRule{
					{
						Title:     "WeekDurationRule",
						UID:       "week-duration-rule",
						Condition: "A",
						Data: []*v1beta1.AlertQuery{
							{
								RefID:         "A",
								DatasourceUID: "__expr__",
								Model:         &v1.JSON{Raw: []byte(`{"expression": "1", "refId": "A"}`)},
							},
						},
						NoDataState:  &noDataState,
						ExecErrState: "Error",
						For:          &weekDuration, // 1w
						Annotations:  map[string]string{},
						Labels:       map[string]string{},
						IsPaused:     false,
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

			r := &GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("AlertRuleGroup Controller Conversion", func() {
	Context("Duration conversion in crToModel", func() {
		It("Should properly convert duration with day duration", func() {
			dayDuration := "1d"

			alertRule := v1beta1.AlertRule{
				Title:        "TestRule",
				UID:          "test-uid",
				ExecErrState: "KeepLast",
				For:          &dayDuration,
				Data:         []*v1beta1.AlertQuery{},
			}

			arg := &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-group",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaAlertRuleGroupSpec{
					Name:      "test-group",
					FolderUID: "test-folder",
					Interval:  metav1.Duration{Duration: 60 * time.Second},
					Rules:     []v1beta1.AlertRule{alertRule},
				},
			}

			model, err := crToModel(arg, "test-folder")
			Expect(err).ToNot(HaveOccurred())

			Expect(model.Rules).To(HaveLen(1))
			Expect(model.Rules[0].For).ToNot(BeNil())
			Expect(model.Rules[0].For.String()).To(Equal("24h0m0s"))
		})

		It("Should properly convert duration with week duration", func() {
			weekDuration := "1w"

			alertRule := v1beta1.AlertRule{
				Title:        "TestRule",
				UID:          "test-uid",
				ExecErrState: "KeepLast",
				For:          &weekDuration,
				Data:         []*v1beta1.AlertQuery{},
			}

			arg := &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-group",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaAlertRuleGroupSpec{
					Name:      "test-group",
					FolderUID: "test-folder",
					Interval:  metav1.Duration{Duration: 60 * time.Second},
					Rules:     []v1beta1.AlertRule{alertRule},
				},
			}

			model, err := crToModel(arg, "test-folder")
			Expect(err).ToNot(HaveOccurred())

			Expect(model.Rules).To(HaveLen(1))
			Expect(model.Rules[0].For).ToNot(BeNil())
			Expect(model.Rules[0].For.String()).To(Equal("168h0m0s"))
		})
	})
})
