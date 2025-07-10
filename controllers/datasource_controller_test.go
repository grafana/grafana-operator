package controllers

import (
	"context"
	"encoding/json"
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
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

		content, hash, err := reconciler.buildDatasourceModel(context.TODO(), cr)
		got := content.SecureJSONData["httpHeaderValue1"]

		assert.Nil(t, err)
		assert.NotEmpty(t, hash)
		assert.Equal(t, want, got)
	})
}

var _ = Describe("Datasource: Reconciler", func() {
	It("Results in NoMatchingInstances Condition", func() {
		// Create object
		cr := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-match",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: instanceSelectorNoMatchingInstances,
				Datasource:        &v1beta1.GrafanaDatasourceInternal{},
			},
		}
		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		Expect(err).ToNot(HaveOccurred())

		// Reconciliation Request
		req := requestFromMeta(cr.ObjectMeta)

		// Reconcile
		r := GrafanaDatasourceReconciler{Client: k8sClient}
		_, err = r.Reconcile(ctx, req)
		Expect(err).ShouldNot(HaveOccurred()) // NoMatchingInstances is a valid reconciliation result

		resultCr := &v1beta1.GrafanaDatasource{}
		Expect(r.Get(ctx, req.NamespacedName, resultCr)).Should(Succeed()) // NoMatchingInstances is a valid status

		// Verify NoMatchingInstances condition
		Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", conditionNoMatchingInstance)))
	})
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
		err := k8sClient.Create(context.Background(), &cm)
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
		content, hash, err := r.buildDatasourceModel(context.TODO(), cr)
		Expect(err).ToNot(HaveOccurred())
		Expect(hash).ToNot(BeEmpty())
		Expect(content.URL).To(Equal(cm.Data["CUSTOM_URL"]))
		marshaled, err := json.Marshal(content.JSONData)
		Expect(err).ToNot(HaveOccurred())
		Expect(marshaled).To(ContainSubstring(cm.Data["CUSTOM_TRACEID"]))
	})
})
