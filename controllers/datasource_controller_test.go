package controllers

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
)

func TestGetDatasourceContent(t *testing.T) {
	reconciler := &GrafanaDatasourceReconciler{
		Client: cl,
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
		Client: cl,
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
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "configmap1", "key1"),
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
							SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret1", "key1"),
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
							SecretKeyRef: tk8s.GetSecretKeySelector(t, sc.Name, "PROMETHEUS_TOKEN"),
						},
					},
					{
						TargetPath: "url",
						ValueFrom: v1beta1.ValueFromSource{
							SecretKeyRef: tk8s.GetSecretKeySelector(t, sc.Name, "URL"),
						},
					},
					{
						TargetPath: "jsonData.exemplarTraceIdDestinations[1].name",
						ValueFrom: v1beta1.ValueFromSource{
							ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, cm.Name, "customTraceId"),
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

		err := cl.Create(testCtx, cm)
		require.NoError(t, err)

		err = cl.Create(testCtx, sc)
		require.NoError(t, err)

		err = cl.Create(testCtx, ds)
		require.NoError(t, err)

		r := GrafanaDatasourceReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, ds)

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

		hasCondition := tk8s.HasCondition(t, cr, condition)
		assert.True(t, hasCondition)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		model, err := gClient.Datasources.GetDataSourceByUID(ds.Spec.CustomUID)
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
	t := GinkgoT()

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
					ValueFrom: v1beta1.ValueFromSource{
						SecretKeyRef: tk8s.GetSecretKeySelector(t, "non-existent-secret", "PROMETHEUS_TOKEN"),
					},
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

			r := &GrafanaDatasourceReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("Datasource correlations", Ordered, func() {
	t := GinkgoT()

	const (
		sourceName = "correlations-source"
		targetName = "correlations-target"
		sourceUID  = "correlations-source-uid"
		targetUID  = "correlations-target-uid"
		label      = "logs-to-traces"
	)

	var (
		r         *GrafanaDatasourceReconciler
		sourceReq ctrl.Request
		targetReq ctrl.Request
		sourceKey types.NamespacedName
		targetKey types.NamespacedName
	)

	findCorrelation := func(correlations []*models.Correlation) *models.Correlation {
		for _, correlation := range correlations {
			if correlation.TargetUID == targetUID && correlation.Label == label {
				return correlation
			}
		}

		return nil
	}

	getCorrelations := func() ([]*models.Correlation, error) {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		if err != nil {
			return nil, err
		}

		correlations, err := gClient.Datasources.GetCorrelationsBySourceUID(sourceUID)
		if err != nil {
			var notFound *datasources.GetCorrelationsBySourceUIDNotFound
			if errors.As(err, &notFound) {
				return nil, nil
			}
			return nil, err
		}

		return correlations.Payload, nil
	}

	BeforeAll(func() {
		r = &GrafanaDatasourceReconciler{Client: cl, Scheme: cl.Scheme()}

		target := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      targetName,
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				CustomUID:         targetUID,
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					Name:   targetName,
					Type:   "prometheus",
					Access: "proxy",
					URL:    "https://demo.promlabs.com",
				},
			},
		}
		targetReq = tk8s.GetRequest(t, target)
		targetKey = tk8s.GetRequestKey(t, target)

		err := cl.Create(testCtx, target)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, targetReq)
		require.NoError(t, err)

		targetPayload, err := json.Marshal(map[string]any{
			"datasourceUid": targetUID,
			"query":         "traceId=$traceId",
			"limit":         25,
			"enabled":       true,
		})
		require.NoError(t, err)

		source := &v1beta1.GrafanaDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      sourceName,
			},
			Spec: v1beta1.GrafanaDatasourceSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				CustomUID:         sourceUID,
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					Name:   sourceName,
					Type:   "prometheus",
					Access: "proxy",
					URL:    "https://demo.promlabs.com",
				},
				Correlations: []v1beta1.GrafanaDatasourceCorrelation{
					{
						TargetUID:   targetUID,
						Label:       label,
						Description: "logs to traces",
						Type:        "query",
						Config: &v1beta1.GrafanaDatasourceCorrelationConfig{
							Field:  "traceId",
							Type:   "query",
							Target: &apiextensionsv1.JSON{Raw: targetPayload},
							Transformations: []v1beta1.GrafanaDatasourceCorrelationTransformation{
								{
									Type:       "regex",
									Field:      "message",
									Expression: "traceId=([a-z0-9]+)",
									MapValue:   "$1",
								},
								{
									Type:     "logfmt",
									MapValue: "traceId",
								},
							},
						},
					},
				},
			},
		}
		sourceReq = tk8s.GetRequest(t, source)
		sourceKey = tk8s.GetRequestKey(t, source)

		err = cl.Create(testCtx, source)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, sourceReq)
		require.NoError(t, err)
	})

	It("creates correlations with config and transformations", func() {
		correlations, err := getCorrelations()
		require.NoError(t, err)

		correlation := findCorrelation(correlations)
		require.NotNil(t, correlation)

		assert.Equal(t, targetUID, correlation.TargetUID)
		assert.Equal(t, label, correlation.Label)
		assert.Equal(t, "logs to traces", correlation.Description)

		require.NotNil(t, correlation.Config)
		require.NotNil(t, correlation.Config.Field)
		assert.Equal(t, "traceId", *correlation.Config.Field)

		target, ok := correlation.Config.Target.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, targetUID, target["datasourceUid"])
		assert.Equal(t, "traceId=$traceId", target["query"])
		assert.Equal(t, float64(25), target["limit"])
		assert.Equal(t, true, target["enabled"])

		require.Len(t, correlation.Config.Transformations, 2)
		assert.Equal(t, "regex", correlation.Config.Transformations[0].Type)
		assert.Equal(t, "message", correlation.Config.Transformations[0].Field)
		assert.Equal(t, "traceId=([a-z0-9]+)", correlation.Config.Transformations[0].Expression)
		assert.Equal(t, "$1", correlation.Config.Transformations[0].MapValue)
		assert.Equal(t, "logfmt", correlation.Config.Transformations[1].Type)
		assert.Equal(t, "traceId", correlation.Config.Transformations[1].MapValue)
	})

	It("updates correlations when config changes", func() {
		cr := &v1beta1.GrafanaDatasource{}
		err := cl.Get(testCtx, sourceKey, cr)
		require.NoError(t, err)

		updatedTarget, err := json.Marshal(map[string]any{
			"datasourceUid": targetUID,
			"query":         "traceId=${traceId}",
			"limit":         50,
			"enabled":       false,
		})
		require.NoError(t, err)

		cr.Spec.Correlations = []v1beta1.GrafanaDatasourceCorrelation{
			{
				TargetUID:   targetUID,
				Label:       label,
				Description: "updated logs to traces",
				Type:        "query",
				Config: &v1beta1.GrafanaDatasourceCorrelationConfig{
					Field:  "traceID",
					Type:   "query",
					Target: &apiextensionsv1.JSON{Raw: updatedTarget},
					Transformations: []v1beta1.GrafanaDatasourceCorrelationTransformation{
						{
							Type:       "regex",
							Field:      "body",
							Expression: "traceID=([A-Z0-9]+)",
							MapValue:   "$1",
						},
					},
				},
			},
		}

		err = cl.Update(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, sourceReq)
		require.NoError(t, err)

		correlations, err := getCorrelations()
		require.NoError(t, err)

		correlation := findCorrelation(correlations)
		require.NotNil(t, correlation)

		assert.Equal(t, "updated logs to traces", correlation.Description)
		require.NotNil(t, correlation.Config)
		require.NotNil(t, correlation.Config.Field)
		assert.Equal(t, "traceID", *correlation.Config.Field)

		target, ok := correlation.Config.Target.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, targetUID, target["datasourceUid"])
		assert.Equal(t, "traceId=${traceId}", target["query"])
		assert.Equal(t, float64(50), target["limit"])
		assert.Equal(t, false, target["enabled"])

		require.Len(t, correlation.Config.Transformations, 1)
		assert.Equal(t, "regex", correlation.Config.Transformations[0].Type)
		assert.Equal(t, "body", correlation.Config.Transformations[0].Field)
		assert.Equal(t, "traceID=([A-Z0-9]+)", correlation.Config.Transformations[0].Expression)
		assert.Equal(t, "$1", correlation.Config.Transformations[0].MapValue)
	})

	It("removes correlations when spec is cleared", func() {
		cr := &v1beta1.GrafanaDatasource{}
		err := cl.Get(testCtx, sourceKey, cr)
		require.NoError(t, err)

		cr.Spec.Correlations = nil

		err = cl.Update(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, sourceReq)
		require.NoError(t, err)

		correlations, err := getCorrelations()
		require.NoError(t, err)

		assert.Empty(t, correlations)

		err = cl.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, sourceReq)
		require.NoError(t, err)

		target := &v1beta1.GrafanaDatasource{}
		err = cl.Get(testCtx, targetKey, target)
		require.NoError(t, err)

		err = cl.Delete(testCtx, target)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, targetReq)
		require.NoError(t, err)
	})
})
