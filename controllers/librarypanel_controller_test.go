package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LibraryPanel Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaLibraryPanel
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaLibraryPanelSpec{
					GrafanaCommonSpec:  commonSpecSuspended,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaLibraryPanelSpec{
					GrafanaCommonSpec:  commonSpecNoMatchingInstances,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaLibraryPanelReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaLibraryPanel{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
