package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NotificationTemplate Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaNotificationTemplate
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaNotificationTemplate{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaNotificationTemplateSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					Name:              "Suspended",
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaNotificationTemplate{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaNotificationTemplateSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					Name:              "NoMatch",
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
			r := GrafanaNotificationTemplateReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaNotificationTemplate{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
