package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
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
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaMuteTimingSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaMuteTimingSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				TimeIntervals:     timeInterval,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaMuteTimingSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				TimeIntervals:     timeInterval,
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaMuteTimingSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				TimeIntervals:     timeInterval,
			},
			want: metav1.Condition{
				Type:   conditionMuteTimingSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaMuteTimingSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Name:              "Synchronized",
				TimeIntervals:     timeInterval,
			},
			want: metav1.Condition{
				Type:   conditionMuteTimingSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaMuteTiming{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaMuteTimingReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
