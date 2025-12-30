package controllers

import (
	"testing"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("PrometheusRuleGroup Reconciler: Provoke Conditions", func() {
	rules := []v1beta1.PrometheusRule{
		{
			Alert:       "TestAlert",
			Expr:        "up == 0",
			For:         "5m",
			Labels:      map[string]string{"severity": "critical"},
			Annotations: map[string]string{"summary": "Instance is down"},
		},
	}

	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaPrometheusRuleGroupSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				FolderUID:         "GroupUID",
				DatasourceUID:     "prometheus-uid",
				Rules:             rules,
				Interval:          metav1.Duration{Duration: 60 * time.Second},
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				FolderUID:         "GroupUID",
				DatasourceUID:     "prometheus-uid",
				Rules:             rules,
				Interval:          metav1.Duration{Duration: 60 * time.Second},
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
			spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				FolderRef:         "pre-existing",
				DatasourceUID:     "prometheus-uid",
				Rules:             rules,
				Interval:          metav1.Duration{Duration: 60 * time.Second},
			},
			want: metav1.Condition{
				Type:   conditionPrometheusRuleGroupSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				FolderRef:         "pre-existing",
				DatasourceUID:     "prometheus-uid",
				Interval:          metav1.Duration{Duration: 60 * time.Second},
				Rules: []v1beta1.PrometheusRule{
					{
						Alert:       "HighLatencyAlert",
						Expr:        "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 1",
						For:         "10m",
						Labels:      map[string]string{"severity": "warning"},
						Annotations: map[string]string{"summary": "High latency detected"},
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionPrometheusRuleGroupSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaPrometheusRuleGroupReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("PrometheusRuleGroup Controller Conversion", func() {
	t := GinkgoT()

	Context("prometheusRuleToModel conversion", func() {
		It("Should properly convert Prometheus rules to Grafana alert rules", func() {
			arg := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-prom-group",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
					Name:          "test-prom-group",
					FolderUID:     "test-folder",
					DatasourceUID: "prometheus-ds",
					Interval:      metav1.Duration{Duration: 60 * time.Second},
					Rules: []v1beta1.PrometheusRule{
						{
							Alert:       "HighCPUUsage",
							Expr:        "100 - (avg by(instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 80",
							For:         "5m",
							Labels:      map[string]string{"severity": "warning"},
							Annotations: map[string]string{"summary": "High CPU usage detected"},
						},
					},
				},
			}

			model, err := prometheusRuleToModel(arg, "test-folder")
			require.NoError(t, err)

			assert.Equal(t, "test-prom-group", model.Title)
			assert.Equal(t, "test-folder", model.FolderUID)
			assert.Equal(t, int64(60), model.Interval)
			assert.Len(t, model.Rules, 1)

			rule := model.Rules[0]
			assert.Equal(t, "HighCPUUsage", *rule.Title)
			assert.Equal(t, "test-folder", *rule.FolderUID)
			assert.NotNil(t, rule.For)
			assert.Equal(t, "5m0s", rule.For.String())
			assert.Equal(t, "warning", rule.Labels["severity"])
			assert.Equal(t, "High CPU usage detected", rule.Annotations["summary"])

			// Verify query data structure
			assert.Len(t, rule.Data, 3) // Query + Reduce + Threshold
			assert.Equal(t, "prometheus-ds", rule.Data[0].DatasourceUID)
			assert.Equal(t, "__expr__", rule.Data[1].DatasourceUID)
			assert.Equal(t, "__expr__", rule.Data[2].DatasourceUID)
		})

		It("Should skip recording rules", func() {
			arg := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-recording-group",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
					Name:          "test-recording-group",
					FolderUID:     "test-folder",
					DatasourceUID: "prometheus-ds",
					Interval:      metav1.Duration{Duration: 60 * time.Second},
					Rules: []v1beta1.PrometheusRule{
						{
							Record: "job:http_requests:rate5m",
							Expr:   "sum(rate(http_requests_total[5m])) by (job)",
						},
						{
							Alert:  "ServiceDown",
							Expr:   "up == 0",
							For:    "1m",
							Labels: map[string]string{"severity": "critical"},
						},
					},
				},
			}

			model, err := prometheusRuleToModel(arg, "test-folder")
			require.NoError(t, err)

			// Only the alerting rule should be converted
			assert.Len(t, model.Rules, 1)
			assert.Equal(t, "ServiceDown", *model.Rules[0].Title)
		})

		It("Should return error when no valid alerting rules found", func() {
			arg := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-only-recording",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
					Name:          "test-only-recording",
					FolderUID:     "test-folder",
					DatasourceUID: "prometheus-ds",
					Interval:      metav1.Duration{Duration: 60 * time.Second},
					Rules: []v1beta1.PrometheusRule{
						{
							Record: "job:http_requests:rate5m",
							Expr:   "sum(rate(http_requests_total[5m])) by (job)",
						},
					},
				},
			}

			_, err := prometheusRuleToModel(arg, "test-folder")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "no valid alerting rules found")
		})

		It("Should handle keep_firing_for duration", func() {
			arg := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-keep-firing",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
					Name:          "test-keep-firing",
					FolderUID:     "test-folder",
					DatasourceUID: "prometheus-ds",
					Interval:      metav1.Duration{Duration: 60 * time.Second},
					Rules: []v1beta1.PrometheusRule{
						{
							Alert:         "TestAlert",
							Expr:          "up == 0",
							For:           "5m",
							KeepFiringFor: "10m",
						},
					},
				},
			}

			model, err := prometheusRuleToModel(arg, "test-folder")
			require.NoError(t, err)

			assert.Len(t, model.Rules, 1)
			assert.Equal(t, "10m0s", model.Rules[0].KeepFiringFor.String())
		})

		It("Should handle day and week durations", func() {
			arg := &v1beta1.GrafanaPrometheusRuleGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-long-duration",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaPrometheusRuleGroupSpec{
					Name:          "test-long-duration",
					FolderUID:     "test-folder",
					DatasourceUID: "prometheus-ds",
					Interval:      metav1.Duration{Duration: 60 * time.Second},
					Rules: []v1beta1.PrometheusRule{
						{
							Alert: "LongDurationAlert",
							Expr:  "up == 0",
							For:   "1d",
						},
					},
				},
			}

			model, err := prometheusRuleToModel(arg, "test-folder")
			require.NoError(t, err)

			assert.Len(t, model.Rules, 1)
			assert.Equal(t, "24h0m0s", model.Rules[0].For.String())
		})
	})

	Context("generateRuleUID", func() {
		It("Should generate consistent UIDs", func() {
			uid1 := generateRuleUID("default", "test-group", "TestAlert", 0)
			uid2 := generateRuleUID("default", "test-group", "TestAlert", 0)

			assert.Equal(t, uid1, uid2)
			assert.Len(t, uid1, 40) // SHA256 first 20 bytes in hex
		})

		It("Should generate different UIDs for different inputs", func() {
			uid1 := generateRuleUID("default", "test-group", "TestAlert", 0)
			uid2 := generateRuleUID("default", "test-group", "TestAlert", 1)
			uid3 := generateRuleUID("default", "other-group", "TestAlert", 0)

			assert.NotEqual(t, uid1, uid2)
			assert.NotEqual(t, uid1, uid3)
		})
	})

	Context("buildPrometheusQueryModel", func() {
		It("Should build valid query model", func() {
			model, err := buildPrometheusQueryModel("up == 0", "prometheus-ds")
			require.NoError(t, err)
			require.NotNil(t, model)
			require.NotNil(t, model.Raw)

			// Verify JSON is valid
			assert.Contains(t, string(model.Raw), "prometheus-ds")
			assert.Contains(t, string(model.Raw), "up == 0")
		})
	})
})

func TestGenerateRuleUID(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		crName    string
		alertName string
		idx       int
	}{
		{
			name:      "basic uid generation",
			namespace: "default",
			crName:    "test-group",
			alertName: "TestAlert",
			idx:       0,
		},
		{
			name:      "unicode alert name",
			namespace: "monitoring",
			crName:    "prometheus-rules",
			alertName: "High CPU 警告",
			idx:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid := generateRuleUID(tt.namespace, tt.crName, tt.alertName, tt.idx)

			// UID should be deterministic
			uid2 := generateRuleUID(tt.namespace, tt.crName, tt.alertName, tt.idx)
			assert.Equal(t, uid, uid2)

			// UID should be 40 characters (SHA256 first 20 bytes in hex)
			assert.Len(t, uid, 40)
		})
	}
}

func TestBuildPrometheusQueryModel(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		datasourceUID string
		wantErr       bool
	}{
		{
			name:          "simple expression",
			expr:          "up == 0",
			datasourceUID: "prometheus-ds",
			wantErr:       false,
		},
		{
			name:          "complex expression",
			expr:          "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))",
			datasourceUID: "prom-1",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := buildPrometheusQueryModel(tt.expr, tt.datasourceUID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, model)
			assert.Contains(t, string(model.Raw), tt.expr)
			assert.Contains(t, string(model.Raw), tt.datasourceUID)
		})
	}
}
