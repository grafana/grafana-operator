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
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
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
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaLibraryPanelSpec{
					GrafanaCommonSpec:  commonSpecNoMatchingInstances,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
			wantErr:       ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaLibraryPanelSpec{
					GrafanaCommonSpec:  commonSpecApplyFailed,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionLibraryPanelSynchronized,
			wantReason:    conditionReasonApplyFailed,
			wantErr:       "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			cr: &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: objectMetaSynchronized,
				Spec: v1beta1.GrafanaLibraryPanelSpec{
					GrafanaCommonSpec: commonSpecSynchronized,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						JSON: `{
							"uid": "do-adhv-ank",
							"name": "API docs Example",
							"type": "text",
							"model": {},
							"version": 1
						}`,
					},
				},
			},
			wantCondition: conditionLibraryPanelSynchronized,
			wantReason:    conditionReasonApplySuccessful,
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaLibraryPanelReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaLibraryPanel{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
