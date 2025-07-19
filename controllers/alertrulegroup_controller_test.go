package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlertRulegroup Reconciler: Provoke Conditions", func() {
	noDataState := "NoData"
	rules := []v1beta1.AlertRule{
		{
			Title:        "TestRule",
			UID:          "akdj-wonvo",
			ExecErrState: "KeepLast",
			NoDataState:  &noDataState,
			Data:         []*v1beta1.AlertQuery{},
		},
	}
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaAlertRuleGroup
		wantCondition string
		wantReason    string
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaAlertRuleGroupSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					FolderUID:         "GroupUID",
					Rules:             rules,
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaAlertRuleGroupSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					FolderUID:         "GroupUID",
					Rules:             rules,
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaAlertRuleGroup{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaAlertRuleGroupSpec{
					GrafanaCommonSpec: commonSpecApplyFailed,
					FolderRef:         "pre-existing",
					Rules:             rules,
				},
			},
			wantCondition: conditionAlertGroupSynchronized,
			wantReason:    conditionReasonApplyFailed,
			wantErr:       "failed to apply to all instances",
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaAlertRuleGroup{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
