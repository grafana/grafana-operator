package controllers

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Folder: Reconciler", func() {
	It("Results in NoMatchingInstances Condition", func() {
		// Create object
		cr := &v1beta1.GrafanaFolder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-match",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: instanceSelectorNoMatchingInstances,
			},
		}
		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		Expect(err).ToNot(HaveOccurred())

		// Reconciliation Request
		req := requestFromMeta(cr.ObjectMeta)

		// Reconcile
		r := GrafanaFolderReconciler{Client: k8sClient}
		_, err = r.Reconcile(ctx, req)
		Expect(err).ShouldNot(HaveOccurred()) // NoMatchingInstances is a valid reconciliation result

		resultCr := &v1beta1.GrafanaFolder{}
		Expect(r.Get(ctx, req.NamespacedName, resultCr)).Should(Succeed()) // NoMatchingInstances is a valid status

		// Verify NoMatchingInstances condition
		Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", conditionNoMatchingInstance)))
	})
})
