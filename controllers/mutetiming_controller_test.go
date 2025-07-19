package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MuteTiming Reconciler: Provoke Conditions", func() {
	timeInterval := []*v1beta1.TimeInterval{
		{
			DaysOfMonth: []string{"1"},
			Location:    "Europe/Copenhagen",
			Times: []*v1beta1.TimeRange{
				{
					StartTime: "00:00",
					EndTime:   "02:00",
				},
			},
			Weekdays: []string{"sunday"},
		},
	}

	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaMuteTiming
		wantCondition string
		wantReason    string
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					TimeIntervals:     timeInterval,
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					TimeIntervals:     timeInterval,
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecApplyFailed,
					TimeIntervals:     timeInterval,
				},
			},
			wantCondition: conditionMuteTimingSynchronized,
			wantReason:    conditionReasonApplyFailed,
			wantErr:       "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			cr: &v1beta1.GrafanaMuteTiming{
				ObjectMeta: objectMetaSynchronized,
				Spec: v1beta1.GrafanaMuteTimingSpec{
					GrafanaCommonSpec: commonSpecSynchronized,
					Name:              "Synchronized",
					TimeIntervals:     timeInterval,
				},
			},
			wantCondition: conditionMuteTimingSynchronized,
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
			r := GrafanaMuteTimingReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaMuteTiming{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
