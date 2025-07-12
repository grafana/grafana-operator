package controllers

import (
	"encoding/json"
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
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

var _ = Describe("Datasource: Reconciler", func() {
	It("Correctly substitutes valuesFrom", func() {
		cm := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-valuesfrom-plain",
				Namespace: "default",
			},
			Data: map[string]string{
				"CUSTOM_URL":     "https://demo.promlabs.com",
				"CUSTOM_TRACEID": "substituted",
			},
		}
		err := k8sClient.Create(testCtx, &cm)
		Expect(err).ToNot(HaveOccurred())
		cr := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				ValuesFrom: []v1beta1.ValueFrom{
					{
						TargetPath: "url",
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Key: "CUSTOM_URL",
							},
						},
					},
					{
						TargetPath: "jsonData.list[0].value",
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Key: "CUSTOM_TRACEID",
							},
						},
					},
				},
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					URL:      "${CUSTOM_URL}",
					JSONData: json.RawMessage([]byte(`{"list":[{"value":"${CUSTOM_TRACEID}"}]}`)),
				},
			},
		}

		r := GrafanaDatasourceReconciler{Client: k8sClient}
		content, hash, err := r.buildDatasourceModel(testCtx, cr)
		Expect(err).ToNot(HaveOccurred())
		Expect(hash).ToNot(BeEmpty())
		Expect(content.URL).To(Equal(cm.Data["CUSTOM_URL"]))
		marshaled, err := json.Marshal(content.JSONData)
		Expect(err).ToNot(HaveOccurred())
		Expect(marshaled).To(ContainSubstring(cm.Data["CUSTOM_TRACEID"]))
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
