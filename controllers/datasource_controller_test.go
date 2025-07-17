package controllers

import (
	"encoding/json"
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetDatasourceContent(t *testing.T) {
	reconciler := &GrafanaDatasourceReconciler{
		Client: k8sClient,
	}

	t.Run("secureJsonData is preserved", func(t *testing.T) {
		secureJSONData := []byte(`{"httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"}`)
		secureJSONDataRaw := (*json.RawMessage)(&secureJSONData)
		// Here, we only check if contents of secureJsonData is preserved, not secrets replacement
		want := "Bearer ${PROMETHEUS_TOKEN}"

		cr := &v1beta1.GrafanaDatasource{
			Spec: v1beta1.GrafanaDatasourceSpec{
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					SecureJSONData: *secureJSONDataRaw,
				},
			},
		}

		content, hash, err := reconciler.buildDatasourceModel(testCtx, cr)
		got := content.SecureJSONData["httpHeaderValue1"]

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Equal(t, want, got)
	})
}

var _ = Describe("Datasource: substitute reference values", func() {
	It("Correctly substitutes valuesFrom", func() {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "ds-valuesfrom-configmap",
			},
			Data: map[string]string{
				"customTraceId": "substituted",
			},
		}
		sc := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "ds-values-from-secret",
			},
			StringData: map[string]string{
				"PROMETHEUS_TOKEN": "secret_token",
				"URL":              "https://demo.promlabs.com",
			},
		}
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "substitute-reference-values",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"dashboards": "grafana",
						},
					},
				},
				CustomUID: "substitute",
				ValuesFrom: []v1beta1.ValueFrom{
					{
						TargetPath: "secureJsonData.httpHeaderValue1",
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: sc.Name,
								},
								Key: "PROMETHEUS_TOKEN",
							},
						},
					},
					{
						TargetPath: "url",
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: sc.Name,
								},
								Key: "URL",
							},
						},
					},
					{
						TargetPath: "jsonData.exemplarTraceIdDestinations[1].name",
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Key: "customTraceId",
							},
						},
					},
				},
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					Name:   "substitute-prometheus",
					Type:   "prometheus",
					Access: "proxy",
					URL:    "${URL}",
					JSONData: json.RawMessage([]byte(`{
						"tlsSkipVerify": true,
						"timeInterval": "10s",
						"httpHeaderName1": "Authorization",
						"exemplarTraceIdDestinations": [
							{"name": "traceID"},
							{"name": "${customTraceId}"}
						]
					}`)),
					SecureJSONData: json.RawMessage([]byte(`{
						"httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"
					}`)),
				},
			},
		}
		Expect(k8sClient.Create(testCtx, cm)).Should(Succeed())
		Expect(k8sClient.Create(testCtx, sc)).Should(Succeed())
		Expect(k8sClient.Create(testCtx, ds)).Should(Succeed())

		req := requestFromMeta(ds.ObjectMeta)
		r := GrafanaDatasourceReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := r.Reconcile(testCtx, req)
		Expect(err).ToNot(HaveOccurred())

		Expect(r.Get(testCtx, req.NamespacedName, ds)).Should(Succeed())
		Expect(ds.Status.Conditions).Should(ContainElement(HaveField("Type", conditionDatasourceSynchronized)))
		Expect(ds.Status.Conditions).Should(ContainElement(HaveField("Reason", conditionReasonApplySuccessful)))

		cl, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		Expect(err).ToNot(HaveOccurred())

		model, err := cl.Datasources.GetDataSourceByUID(ds.Spec.CustomUID)
		Expect(err).ToNot(HaveOccurred())

		Expect(model.Payload.URL).To(Equal("https://demo.promlabs.com"))
		Expect(model.Payload.SecureJSONFields["httpHeaderValue1"]).To(BeTrue())

		// Serialize and Derserialize jsonData
		b, err := json.Marshal(model.Payload.JSONData)
		Expect(err).ToNot(HaveOccurred())

		type ExemplarTraceIDDestination struct {
			Name string `json:"name"`
		}
		type SubstitutedJSONData struct {
			ExemplarTraceIDDestinations []ExemplarTraceIDDestination `json:"exemplarTraceIdDestinations"`
		}
		var jsonData SubstitutedJSONData // map with array of
		err = json.Unmarshal(b, &jsonData)
		Expect(err).ToNot(HaveOccurred())
		Expect(jsonData.ExemplarTraceIDDestinations[1].Name).To(Equal("substituted"))
	})
})

var _ = Describe("Datasource Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaDatasource
		wantCondition string
		wantReason    string
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.GrafanaDatasource{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaDatasourceSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					Datasource:        &v1beta1.GrafanaDatasourceInternal{},
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaDatasource{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaDatasourceSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					Datasource:        &v1beta1.GrafanaDatasourceInternal{},
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaDatasource{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaDatasourceSpec{
					GrafanaCommonSpec: commonSpecApplyFailed,
					Datasource:        &v1beta1.GrafanaDatasourceInternal{},
				},
			},
			wantCondition: conditionDatasourceSynchronized,
			wantReason:    conditionReasonApplyFailed,
			wantErr:       "failed to apply to all instances",
		},
		{
			name: "Referenced secret does not exist",
			cr: &v1beta1.GrafanaDatasource{
				ObjectMeta: objectMetaInvalidSpec,
				Spec: v1beta1.GrafanaDatasourceSpec{
					GrafanaCommonSpec: commonSpecInvalidSpec,
					Datasource:        &v1beta1.GrafanaDatasourceInternal{},
					ValuesFrom: []v1beta1.ValueFrom{{
						TargetPath: "secureJsonData.httpHeaderValue1",
						ValueFrom: v1beta1.ValueFromSource{SecretKeyRef: &corev1.SecretKeySelector{
							Key: "credentials",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "PROMETHEUS_TOKEN",
							},
						}},
					}},
				},
			},
			wantCondition: conditionInvalidSpec,
			wantReason:    conditionReasonInvalidModel,
			wantErr:       "building datasource model",
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaDatasourceReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaDatasource{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
