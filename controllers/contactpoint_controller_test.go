package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ContactPoint Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaContactPointSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte("{}")},
				Type:              "webhook",
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte("{}")},
				Type:              "webhook",
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
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte("{}")},
				Type:              "webhook",
			},
			want: metav1.Condition{
				Type:   conditionContactPointSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Referenced secret does not exist",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte("{}")},
				Type:              "email",
				ValuesFrom: []v1beta1.ValueFrom{{
					TargetPath: "addresses",
					ValueFrom: v1beta1.ValueFromSource{SecretKeyRef: &corev1.SecretKeySelector{
						Key: "contact-mails",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "alert-mails",
						},
					}},
				}},
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidSettings,
			},
			wantErr: "building contactpoint settings",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
				Type:              "webhook",
			},
			want: metav1.Condition{
				Type:   conditionContactPointSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaContactPoint{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaContactPointReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
