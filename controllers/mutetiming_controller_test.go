package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MuteTiming Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaMuteTiming
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					TimeIntervals: []*v1beta1.TimeInterval{
						{
							DaysOfMonth: []string{"1"},
							Location:    "Europe/Copenhagen",
							Months:      []string{"1"},
							Times: []*v1beta1.TimeRange{
								{
									StartTime: "00:00",
									EndTime:   "02:00",
								},
							},
							Weekdays: []string{"1"},
							Years:    []string{"2025"},
						},
					},
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					TimeIntervals: []*v1beta1.TimeInterval{
						{
							DaysOfMonth: []string{"1"},
							Location:    "Europe/Copenhagen",
							Months:      []string{"1"},
							Times: []*v1beta1.TimeRange{
								{
									StartTime: "00:00",
									EndTime:   "02:00",
								},
							},
							Weekdays: []string{"1"},
							Years:    []string{"2025"},
						},
					},
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
			r := GrafanaMuteTimingReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaMuteTiming{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
