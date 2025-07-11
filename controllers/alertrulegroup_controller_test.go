package controllers

import (
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			For:          &metav1.Duration{Duration: 60 * time.Second},
			Data:         []*v1beta1.AlertQuery{},
		},
	}
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaAlertRuleGroup
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspend Condition",
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
			name: "NoMatchingInstances Condition",
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
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaAlertRuleGroupReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaAlertRuleGroup{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
