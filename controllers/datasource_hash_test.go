// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasourceHashStable(t *testing.T) {
	r := &GrafanaDatasourceReconciler{}
	cr := &v1beta1.GrafanaDatasource{
		Spec: v1beta1.GrafanaDatasourceSpec{
			CustomUID: "postgres-hash-test",
			Datasource: &v1beta1.GrafanaDatasourceInternal{
				Name: "postgres", Type: "grafana-postgresql-datasource", Access: "proxy",
				URL: "localhost:5432", User: "reader", Database: "example_database",
				JSONData:       json.RawMessage(`{"sslmode":"disable","database":"example_database","postgresVersion":1600}`),
				SecureJSONData: json.RawMessage(`{"password":"original"}`),
			},
		},
	}
	_, want, err := r.buildDatasourceModel(context.Background(), cr.DeepCopy())
	require.NoError(t, err)

	for range 100 {
		_, got, err := r.buildDatasourceModel(context.Background(), cr.DeepCopy())
		require.NoError(t, err)
		require.Equal(t, want, got, "unchanged datasource must not trigger a new update")
	}

	reordered := cr.DeepCopy()
	reordered.Spec.Datasource.JSONData = json.RawMessage(`{"postgresVersion":1600, "database":"example_database", "sslmode":"disable"}`)
	_, got, err := r.buildDatasourceModel(context.Background(), reordered)
	require.NoError(t, err)
	require.Equal(t, want, got, "JSON key order must not trigger a new update")
}

func TestDatasourceHashChangesWithPayload(t *testing.T) {
	r := &GrafanaDatasourceReconciler{}
	cr := &v1beta1.GrafanaDatasource{
		Spec: v1beta1.GrafanaDatasourceSpec{
			CustomUID: "postgres-hash-test",
			Datasource: &v1beta1.GrafanaDatasourceInternal{
				Name: "postgres", Type: "grafana-postgresql-datasource", Access: "proxy",
				URL: "localhost:5432", User: "reader", Database: "example_database",
				JSONData:       json.RawMessage(`{"sslmode":"disable","database":"example_database"}`),
				SecureJSONData: json.RawMessage(`{"password":"original"}`),
			},
		},
	}
	_, original, err := r.buildDatasourceModel(context.Background(), cr.DeepCopy())
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		change func(*v1beta1.GrafanaDatasource)
	}{
		{"credentials", func(changed *v1beta1.GrafanaDatasource) {
			changed.Spec.Datasource.SecureJSONData = json.RawMessage(`{"password":"rotated"}`)
		}},
		{"database", func(changed *v1beta1.GrafanaDatasource) { changed.Spec.Datasource.Database = "other" }},
		{"jsonData", func(changed *v1beta1.GrafanaDatasource) {
			changed.Spec.Datasource.JSONData = json.RawMessage(`{"sslmode":"require","database":"example_database"}`)
		}},
		{"uid", func(changed *v1beta1.GrafanaDatasource) { changed.Spec.CustomUID = "other" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := cr.DeepCopy()
			tc.change(changed)
			_, got, err := r.buildDatasourceModel(context.Background(), changed)
			require.NoError(t, err)
			require.NotEqual(t, original, got)
		})
	}
}

func TestUnchangedDatasourceDoesNotUpdateGrafana(t *testing.T) {
	var updates atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if req.Method == http.MethodGet && req.URL.Path == "/api/datasources" {
			_, err := w.Write([]byte(`[{"uid":"postgres-hash-test","name":"postgres"}]`))
			assert.NoError(t, err)

			return
		}

		if req.Method == http.MethodPut {
			updates.Add(1)
		}

		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"message":"unexpected datasource update"}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	r := &GrafanaDatasourceReconciler{}
	grafana := &v1beta1.Grafana{
		Spec:   v1beta1.GrafanaSpec{External: &v1beta1.External{URL: server.URL}, Config: map[string]map[string]string{"security": {"admin_user": "dummy", "admin_password": "dummy"}}},
		Status: v1beta1.GrafanaStatus{AdminURL: server.URL},
	}
	cr := &v1beta1.GrafanaDatasource{
		Spec: v1beta1.GrafanaDatasourceSpec{
			CustomUID: "postgres-hash-test",
			Datasource: &v1beta1.GrafanaDatasourceInternal{
				Name: "postgres", Type: "grafana-postgresql-datasource", Access: "proxy",
				URL: "localhost:5432", User: "reader", Database: "example_database",
				JSONData: json.RawMessage(`{"sslmode":"disable","database":"example_database"}`),
			},
		},
	}
	_, storedHash, err := r.buildDatasourceModel(context.Background(), cr.DeepCopy())
	require.NoError(t, err)

	cr.Status.Hash = storedHash
	for range 20 {
		payload, hash, err := r.buildDatasourceModel(context.Background(), cr.DeepCopy())
		require.NoError(t, err)

		applyErr := r.onDatasourceCreated(context.Background(), grafana, cr, payload, hash)
		// Stop after a recorded mutation; the final assertion reports the regression.
		// Any error without a PUT is a setup/transport failure, not proof of the bug.
		if updates.Load() > 0 {
			break
		}

		require.NoError(t, applyErr)
	}

	require.Zero(t, updates.Load(), "unchanged reconciliations must not overwrite Grafana")
}
