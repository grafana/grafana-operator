package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContactPoint Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaContactPoint
		wantCondition string
		wantReason    string
		wantErr       string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaContactPointSpec{
					GrafanaCommonSpec: commonSpecSuspended,
					Name:              "ContactPointName",
					Settings:          &v1.JSON{Raw: []byte("{}")},
					Type:              "webhook",
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaContactPointSpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					Name:              "ContactPointName",
					Settings:          &v1.JSON{Raw: []byte("{}")},
					Type:              "webhook",
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
		{
			name: "Failed to apply to instance",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaApplyFailed,
				Spec: v1beta1.GrafanaContactPointSpec{
					GrafanaCommonSpec: commonSpecApplyFailed,
					Name:              "ContactPointName",
					Settings:          &v1.JSON{Raw: []byte("{}")},
					Type:              "webhook",
				},
			},
			wantCondition: conditionContactPointSynchronized,
			wantReason:    conditionReasonApplyFailed,
			wantErr:       "failed to apply to all instances",
		},
		{
			name: "Referenced secret does not exist",
			cr: &v1beta1.GrafanaContactPoint{
				ObjectMeta: objectMetaInvalidSpec,
				Spec: v1beta1.GrafanaContactPointSpec{
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
			},
			wantCondition: conditionInvalidSpec,
			wantReason:    conditionReasonInvalidSettings,
			wantErr:       "building contactpoint settings",
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaContactPointReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			if test.wantErr == "" {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix(test.wantErr))
			}

			resultCr := &v1beta1.GrafanaContactPoint{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
