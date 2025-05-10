package controllers

import (
	"context"
	"encoding/json"
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
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

func TestGetDatasourcesToDelete(t *testing.T) {
	f := func(ds []v1beta1.GrafanaDatasource, grafana v1beta1.Grafana, expected []v1beta1.NamespacedResource) {
		t.Helper()
		datasourcesList := v1beta1.GrafanaDatasourceList{
			TypeMeta: metav1.TypeMeta{},
			ListMeta: metav1.ListMeta{},
			Items:    ds,
		}
		datasourcesToDelete := getDatasourcesToDelete(&datasourcesList, []v1beta1.Grafana{grafana})
		for _, out := range datasourcesToDelete {
			assert.Equal(t, out, expected)
		}
	}

	f([]v1beta1.GrafanaDatasource{
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "datasource-a",
				Namespace: "namespace",
			},
			Status: v1beta1.GrafanaDatasourceStatus{
				UID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			},
		},
	},
		v1beta1.Grafana{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grafana-1",
				Namespace: "namespace",
			},
			Status: v1beta1.GrafanaStatus{
				Datasources: v1beta1.NamespacedResourceList{
					"namespace/datasource-a/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"namespace/datasource-c/cccccccc-cccc-cccc-cccc-cccccccccccc",
				},
			},
		},
		[]v1beta1.NamespacedResource{
			"namespace/datasource-c/cccccccc-cccc-cccc-cccc-cccccccccccc",
		},
	)
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
})
