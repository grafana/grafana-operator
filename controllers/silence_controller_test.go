package controllers

import (
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Silence Reconciler: Provoke Conditions", func() {
	matchers := []*v1beta1.SilenceMatcher{
		{
			Name:    "alertname",
			Value:   "HighCPU",
			IsEqual: true,
		},
	}

	startsAt := metav1.Time{Time: time.Now()}
	endsAt := metav1.Time{Time: time.Now().Add(2 * time.Hour)}

	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaSilenceSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaSilenceSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				Matchers:          matchers,
				StartsAt:          startsAt,
				EndsAt:            endsAt,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaSilenceSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				Matchers:          matchers,
				StartsAt:          startsAt,
				EndsAt:            endsAt,
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
			spec: v1beta1.GrafanaSilenceSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				Matchers:          matchers,
				StartsAt:          startsAt,
				EndsAt:            endsAt,
			},
			want: metav1.Condition{
				Type:   conditionSilenceSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: LogMsgApplyErrors,
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaSilenceSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Matchers:          matchers,
				StartsAt:          startsAt,
				EndsAt:            endsAt,
				Comment:           "Synchronized",
				CreatedBy:         "grafana-operator",
			},
			want: metav1.Condition{
				Type:   conditionSilenceSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaSilence{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaSilenceReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
