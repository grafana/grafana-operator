package controllers

import (
	"encoding/json"
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

func TestGetDatasourceContent(t *testing.T) {
	reconciler := &GrafanaDatasourceReconciler{
		Client: k8sClient,
	}

	t.Run("secureJsonData is preserved", func(t *testing.T) {
		secureJSONData := []byte(`{"httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"}`)
		secureJSONDataRaw := (*json.RawMessage)(&secureJSONData)
		// Here, we only check if contents of secureJsonData is preserved, not secrets replacement
		want := "Bearer ${PROMETHEUS_TOKEN}"

		cr := &v1beta1.GrafanaDatasource{
			Spec: v1beta1.GrafanaDatasourceSpec{
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					SecureJSONData: *secureJSONDataRaw,
				},
			},
		}

		content, hash, err := reconciler.buildDatasourceModel(testCtx, cr)
		got := content.SecureJSONData["httpHeaderValue1"]

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Equal(t, want, got)
	})
}

func TestDatasourceIndexing(t *testing.T) {
	reconciler := &GrafanaDatasourceReconciler{
		Client: k8sClient,
	}

	t.Run("indexSecretSource returns correct secret references", func(t *testing.T) {
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-datasource",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
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
		result := indexFunc(ds)

		expected := []string{"test-namespace/secret1", "test-namespace/secret2"}
		require.Equal(t, expected, result)
	})

	t.Run("indexConfigMapSource returns correct configmap references", func(t *testing.T) {
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-datasource",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
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
		result := indexFunc(ds)

		expected := []string{"test-namespace/configmap1", "test-namespace/configmap2"}
		require.Equal(t, expected, result)
	})

	t.Run("indexSecretSource returns empty slice when no secret references", func(t *testing.T) {
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-datasource",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
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
		result := indexFunc(ds)

		require.Empty(t, result)
	})

	t.Run("indexConfigMapSource returns empty slice when no configmap references", func(t *testing.T) {
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-datasource",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
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
		result := indexFunc(ds)

		require.Empty(t, result)
	})
}

var _ = Describe("Datasource: substitute reference values", func() {
	t := GinkgoT()

	It("Correctly substitutes valuesFrom", func() {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "ds-valuesfrom-configmap",
			},
			Data: map[string]string{
				"customTraceId": "substituted",
			},
		}
		sc := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "ds-values-from-secret",
			},
			StringData: map[string]string{
				"PROMETHEUS_TOKEN": "secret_token",
				"URL":              "https://demo.promlabs.com",
			},
		}
		ds := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "substitute-reference-values",
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"dashboards": "grafana",
						},
					},
				},
				CustomUID: "substitute",
				ValuesFrom: []v1beta1.ValueFrom{
					{
						TargetPath: "secureJsonData.httpHeaderValue1",
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: sc.Name,
								},
								Key: "PROMETHEUS_TOKEN",
							},
						},
					},
					{
						TargetPath: "url",
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: sc.Name,
								},
								Key: "URL",
							},
						},
					},
					{
						TargetPath: "jsonData.exemplarTraceIdDestinations[1].name",
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Key: "customTraceId",
							},
						},
					},
				},
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					Name:   "substitute-prometheus",
					Type:   "prometheus",
					Access: "proxy",
					URL:    "${URL}",
					JSONData: json.RawMessage([]byte(`{
						"tlsSkipVerify": true,
						"timeInterval": "10s",
						"httpHeaderName1": "Authorization",
						"exemplarTraceIdDestinations": [
							{"name": "traceID"},
							{"name": "${customTraceId}"}
						]
					}`)),
					SecureJSONData: json.RawMessage([]byte(`{
						"httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"
					}`)),
				},
			},
		}

		err := k8sClient.Create(testCtx, cm)
		require.NoError(t, err)

		err = k8sClient.Create(testCtx, sc)
		require.NoError(t, err)

		err = k8sClient.Create(testCtx, ds)
		require.NoError(t, err)

		r := GrafanaDatasourceReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		req := requestFromMeta(ds.ObjectMeta)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		cr := &v1beta1.GrafanaDatasource{
			ObjectMeta: *ds.ObjectMeta.DeepCopy(),
		}

		condition := metav1.Condition{
			Type:   conditionDatasourceSynchronized,
			Reason: conditionReasonApplySuccessful,
		}

		err = r.Get(testCtx, req.NamespacedName, cr)
		require.NoError(t, err)

		containsEqualCondition(cr.Status.Conditions, condition)

		cl, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		model, err := cl.Datasources.GetDataSourceByUID(ds.Spec.CustomUID)
		require.NoError(t, err)

		assert.Equal(t, "https://demo.promlabs.com", model.Payload.URL)
		assert.True(t, model.Payload.SecureJSONFields["httpHeaderValue1"])

		// Serialize and Derserialize jsonData
		b, err := json.Marshal(model.Payload.JSONData)
		require.NoError(t, err)

		type ExemplarTraceIDDestination struct {
			Name string `json:"name"`
		}

		type SubstitutedJSONData struct {
			ExemplarTraceIDDestinations []ExemplarTraceIDDestination `json:"exemplarTraceIdDestinations"`
		}

		var jsonData SubstitutedJSONData // map with array of

		err = json.Unmarshal(b, &jsonData)
		require.NoError(t, err)

		assert.Equal(t, "substituted", jsonData.ExemplarTraceIDDestinations[1].Name)
	})
})

var _ = Describe("Datasource Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaDatasourceSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecSuspended,
				Datasource:        &v1beta1.GrafanaDatasourceInternal{},
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				Datasource:        &v1beta1.GrafanaDatasourceInternal{},
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
			spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				Datasource:        &v1beta1.GrafanaDatasourceInternal{},
			},
			want: metav1.Condition{
				Type:   conditionDatasourceSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "Referenced secret does not exist",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Datasource:        &v1beta1.GrafanaDatasourceInternal{},
				ValuesFrom: []v1beta1.ValueFrom{{
					TargetPath: "secureJsonData.httpHeaderValue1",
					ValueFrom: v1beta1.ValueFromSource{SecretKeyRef: &corev1.SecretKeySelector{
						Key: "credentials",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "PROMETHEUS_TOKEN",
						},
					}},
				}},
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidModel,
			},
			wantErr: "building datasource model",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					Name:   "synced-prometheus",
					Type:   "prometheus",
					Access: "proxy",
					URL:    "https://demo.promlabs.com",
				},
			},
			want: metav1.Condition{
				Type:   conditionDatasourceSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaDatasource{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaDatasourceReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})
