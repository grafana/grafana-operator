package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Folder Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaFolder
		wantCondition string
		wantReason    string
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.GrafanaFolder{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaFolderSpec{
					GrafanaCommonSpec: commonSpecSuspended,
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaFolder{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaFolderSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaFolder{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaFolderSpec{
					GrafanaCommonSpec: commonSpecApplyFailed,
				},
			},
			wantCondition: conditionFolderSynchronized,
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
			r := GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaFolder{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
