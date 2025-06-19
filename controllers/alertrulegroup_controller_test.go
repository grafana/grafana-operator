package controllers

import (
	"context"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlertRulegroup: Reconciler", func() {
	It("Results in NoMatchingInstances Condition", func() {
		// Create object
		noDataState := "NoData"
		cr := &v1beta1.GrafanaAlertRuleGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-match",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: instanceSelectorNoMatchingInstances,
				FolderUID:         "GroupUID",
				Rules: []v1beta1.AlertRule{
					{
						Title:        "TestRule",
						UID:          "akdj-wonvo",
						ExecErrState: "KeepLast",
						NoDataState:  &noDataState,
						For:          &metav1.Duration{Duration: 60 * time.Second},
						Data:         []*v1beta1.AlertQuery{},
					},
				},
			},
		}
		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		Expect(err).ToNot(HaveOccurred())

		// Reconciliation Request
		req := requestFromMeta(cr.ObjectMeta)

		// Reconcile
		r := GrafanaAlertRuleGroupReconciler{Client: k8sClient}
		_, err = r.Reconcile(ctx, req)
		Expect(err).ShouldNot(HaveOccurred()) // NoMatchingInstances is a valid reconciliation result

		resultCr := &v1beta1.GrafanaAlertRuleGroup{}
		Expect(r.Get(ctx, req.NamespacedName, resultCr)).Should(Succeed()) // NoMatchingInstances is a valid status

		// Verify NoMatchingInstances condition
		Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", conditionNoMatchingInstance)))
	})
})
