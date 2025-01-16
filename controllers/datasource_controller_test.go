package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestGetDatasourceContent(t *testing.T) {
	reconciler := &GrafanaDatasourceReconciler{
		Client: k8sClient,
		Log:    ctrl.Log.WithName("TestDatasourceReconciler"),
	}

	t.Run("secureJsonData is preserved", func(t *testing.T) {
		secureJsonData := []byte(`{"httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"}`)
		secureJsonDataRaw := (*json.RawMessage)(&secureJsonData)
		// Here, we only check if contents of secureJsonData is preserved, not secrets replacement
		want := "Bearer ${PROMETHEUS_TOKEN}"

		cr := &v1beta1.GrafanaDatasource{
			Spec: v1beta1.GrafanaDatasourceSpec{
				Datasource: &v1beta1.GrafanaDatasourceInternal{
					SecureJSONData: *secureJsonDataRaw,
				},
			},
		}

		content, hash, err := reconciler.buildDatasourceModel(context.TODO(), cr)
		got := content.SecureJSONData["httpHeaderValue1"]

		assert.Nil(t, err)
		assert.NotEmpty(t, hash)
		assert.Equal(t, want, got)
	})
}
