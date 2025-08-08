package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("NotificationTemplate Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaNotificationTemplateSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaNotificationTemplateSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				Name:              "Suspended",
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaNotificationTemplateSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				Name:              "NoMatch",
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
			spec: v1beta1.GrafanaNotificationTemplateSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				Name:              "ApplyFailed",
			},
			want: metav1.Condition{
				Type:   conditionNotificationTemplateSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaNotificationTemplateSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Name:              "Synchronized",
				Template:          `{{ define "StatusAlert" }}{{.Status}}{{ end }}`,
			},
			want: metav1.Condition{
				Type:   conditionNotificationTemplateSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaNotificationTemplate{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaNotificationTemplateReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
