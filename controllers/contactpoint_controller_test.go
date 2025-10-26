package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
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
			name: "Top level receiver missing settings with type defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "missing-settings",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Name:              "ContactPointName",
				Settings:          nil,
				Type:              "email",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonTopLevelReceiver,
			},
			wantErr: ErrInvalidTopLevelReceiver.Error(),
		},
		{
			name: "Top level receiver missing type with settings defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "missing-type",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Name:              "ContactPointName",
				Settings:          &v1.JSON{Raw: []byte("{}")},
				Type:              "",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonTopLevelReceiver,
			},
			wantErr: ErrInvalidTopLevelReceiver.Error(),
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
		{
			name: "Successfully applied multiple receiver contactpoint to instance",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "synchronized-multiple-receivers",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Name:              "ContactPointName",
				Receivers: []v1beta1.ContactPointReceiver{
					{
						Settings: &v1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
						Type:     "webhook",
					},
					{
						Settings: &v1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
						Type:     "webhook",
					},
				},
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

func TestContactPointIndexing(t *testing.T) {
	reconciler := &GrafanaContactPointReconciler{
		Client: k8sClient,
	}

	t.Run("indexSecretSource returns correct secret references", func(t *testing.T) {
		cp := &v1beta1.GrafanaContactPoint{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-contactpoint",
			},
			Spec: v1beta1.GrafanaContactPointSpec{
				ValuesFrom: []v1beta1.ValueFrom{
					{
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret1",
								},
								Key: "key1",
							},
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret2",
								},
								Key: "key2",
							},
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "configmap1",
								},
								Key: "key1",
							},
						},
					},
				},
			},
		}

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cp)

		expected := []string{"test-namespace/secret1", "test-namespace/secret2"}
		require.Equal(t, expected, result)
	})

	t.Run("indexConfigMapSource returns correct configmap references", func(t *testing.T) {
		cp := &v1beta1.GrafanaContactPoint{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-contactpoint",
			},
			Spec: v1beta1.GrafanaContactPointSpec{
				ValuesFrom: []v1beta1.ValueFrom{
					{
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret1",
								},
								Key: "key1",
							},
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "configmap1",
								},
								Key: "key1",
							},
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "configmap2",
								},
								Key: "key2",
							},
						},
					},
				},
			},
		}

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cp)

		expected := []string{"test-namespace/configmap1", "test-namespace/configmap2"}
		require.Equal(t, expected, result)
	})

	t.Run("indexSecretSource returns empty slice when no secret references", func(t *testing.T) {
		cp := &v1beta1.GrafanaContactPoint{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-contactpoint",
			},
			Spec: v1beta1.GrafanaContactPointSpec{
				ValuesFrom: []v1beta1.ValueFrom{
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "configmap1",
								},
								Key: "key1",
							},
						},
					},
				},
			},
		}

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cp)

		require.Empty(t, result)
	})

	t.Run("indexConfigMapSource returns empty slice when no configmap references", func(t *testing.T) {
		cp := &v1beta1.GrafanaContactPoint{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-contactpoint",
			},
			Spec: v1beta1.GrafanaContactPointSpec{
				ValuesFrom: []v1beta1.ValueFrom{
					{
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secret1",
								},
								Key: "key1",
							},
						},
					},
				},
			},
		}

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cp)

		require.Empty(t, result)
	})
}
