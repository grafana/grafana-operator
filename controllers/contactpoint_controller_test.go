package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContactPoint Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaContactPoint
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaContactPointSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					Name:              "ContactPointName",
					Settings:          &v1.JSON{Raw: []byte("{}")},
					Type:              "webhook",
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaContactPointSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					Name:              "ContactPointName",
					Settings:          &v1.JSON{Raw: []byte("{}")},
					Type:              "webhook",
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
			r := GrafanaContactPointReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaContactPoint{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
