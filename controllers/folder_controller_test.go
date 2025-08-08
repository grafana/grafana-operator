package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Folder Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaFolderSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSuspended,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
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
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
			},
			want: metav1.Condition{
				Type:   conditionFolderSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "InvalidSpec Condition",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				CustomUID:         "self-ref",
				ParentFolderUID:   "self-ref",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonCyclicParent,
			},
			wantErr: "cyclic folder reference",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
			},
			want: metav1.Condition{
				Type:   conditionFolderSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaFolder{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
