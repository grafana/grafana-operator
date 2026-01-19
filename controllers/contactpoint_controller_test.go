package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
)

var _ = Describe("ContactPoint Reconciler: Provoke Conditions", func() {
	t := GinkgoT()

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
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
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
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
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
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
				Type:              "webhook",
			},
			want: metav1.Condition{
				Type:   conditionContactPointSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: ErrMsgApplyErrors,
		},
		{
			name: "Referenced secret does not exist",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
				Type:              "email",
				ValuesFrom: []v1beta1.ValueFrom{{
					TargetPath: "addresses",
					ValueFrom: v1beta1.ValueFromSource{
						SecretKeyRef: tk8s.GetSecretKeySelector(t, "alert-mails", "contact-mails"),
					},
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
				Settings:          nil,
				Type:              "email",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidContactPoint,
			},
			wantErr: ErrMissingContactPointReceiver.Error(),
		},
		{
			name: "Top level receiver missing type with settings defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "missing-type",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
				Type:              "",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidContactPoint,
			},
			wantErr: ErrMissingContactPointReceiver.Error(),
		},
		{
			name: "No receivers defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "missing-receiver-configs",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidContactPoint,
			},
			wantErr: ErrMissingContactPointReceiver.Error(),
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Settings:          &apiextensionsv1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
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
				Receivers: []v1beta1.ContactPointReceiver{
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
						Type:     "webhook",
					},
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte(`{"url": "http://test.io"}`)},
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

			r := &GrafanaContactPointReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("ContactPoint valuesFrom configurations", Ordered, func() {
	t := GinkgoT()

	sc := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "contacts",
		},
		StringData: map[string]string{
			"one": "one@example.invalid",
			"two": "two@example.invalid",
		},
	}

	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaContactPointSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: "Referenced secrets missing",
			meta: metav1.ObjectMeta{
				GenerateName: "receiver-secrets-missing",
				Namespace:    "default",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
				Type:              "email",
				ValuesFrom: []v1beta1.ValueFrom{{
					TargetPath: "addresses",
					ValueFrom: v1beta1.ValueFromSource{
						SecretKeyRef: tk8s.GetSecretKeySelector(t, "nil", "contact-mails"),
					},
				}},
				Receivers: []v1beta1.ContactPointReceiver{
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte("{}")},
						Type:     "email",
						ValuesFrom: []v1beta1.ValueFrom{{
							TargetPath: "addresses",
							ValueFrom: v1beta1.ValueFromSource{
								SecretKeyRef: tk8s.GetSecretKeySelector(t, "nil", "contact-mails"),
							},
						}},
					},
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte("{}")},
						Type:     "email",
						ValuesFrom: []v1beta1.ValueFrom{{
							TargetPath: "addresses",
							ValueFrom: v1beta1.ValueFromSource{
								SecretKeyRef: tk8s.GetSecretKeySelector(t, "nil", "contact-mails"),
							},
						}},
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidSettings,
			},
			wantErr: "building contactpoint settings",
		},
		{
			name: "Referenced secret exist",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "successful-apply",
			},
			spec: v1beta1.GrafanaContactPointSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Settings:          &apiextensionsv1.JSON{Raw: []byte("{}")},
				Type:              "email",
				ValuesFrom: []v1beta1.ValueFrom{{
					TargetPath: "addresses",
					ValueFrom: v1beta1.ValueFromSource{
						SecretKeyRef: tk8s.GetSecretKeySelector(t, sc.Name, "one"),
					},
				}},
				Receivers: []v1beta1.ContactPointReceiver{
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte("{}")},
						Type:     "email",
						ValuesFrom: []v1beta1.ValueFrom{{
							TargetPath: "addresses",
							ValueFrom: v1beta1.ValueFromSource{
								SecretKeyRef: tk8s.GetSecretKeySelector(t, sc.Name, "one"),
							},
						}},
					},
					{
						Settings: &apiextensionsv1.JSON{Raw: []byte("{}")},
						Type:     "email",
						ValuesFrom: []v1beta1.ValueFrom{{
							TargetPath: "addresses",
							ValueFrom: v1beta1.ValueFromSource{
								SecretKeyRef: tk8s.GetSecretKeySelector(t, sc.Name, "two"),
							},
						}},
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionContactPointSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	BeforeAll(func() {
		t := GinkgoT()

		err := cl.Create(testCtx, &sc)
		require.NoError(t, err)
	})

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaContactPoint{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaContactPointReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

func TestContactPointIndexing(t *testing.T) {
	reconciler := &GrafanaContactPointReconciler{
		Client: cl,
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
							SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret1", "key1"),
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret2", "key2"),
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "configmap1", "key1"),
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
							SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret1", "key1"),
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "configmap1", "key1"),
						},
					},
					{
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "configmap2", "key2"),
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
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "configmap1", "key1"),
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
							SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret1", "key1"),
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
