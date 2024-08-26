package fetchers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestFetchDashboardFromUrl(t *testing.T) {
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)
	compressedJSON, err := v1beta1.Gzip(dashboardJSON)
	assert.Nil(t, err, "Failed to compress a dashboard")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(dashboardJSON)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	dashboard := &v1beta1.GrafanaDashboard{
		Spec: v1beta1.GrafanaDashboardSpec{
			Url: ts.URL,
		},
		Status: v1beta1.GrafanaDashboardStatus{},
	}

	fetchedDashboard, err := FetchDashboardFromUrl(context.Background(), dashboard, k8sClient, nil)
	assert.Nil(t, err)
	assert.Equal(t, dashboardJSON, fetchedDashboard, "Fetched dashboard doesn't match the original")

	assert.False(t, dashboard.Status.ContentTimestamp.Time.IsZero(), "ContentTimestamp should have been updated")
	assert.Equal(t, compressedJSON, dashboard.Status.ContentCache, "ContentCache should have been updated")
	assert.Equal(t, ts.URL, dashboard.Status.ContentUrl, "ContentUrl should have been updated")
}

func TestFetchDashboardFromUrlBasicAuth(t *testing.T) {
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)
	compressedJSON, err := v1beta1.Gzip(dashboardJSON)
	assert.Nil(t, err, "Failed to compress a dashboard")

	basicAuthUsername := "admin"
	basicAuthPassword := "admin"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, username, basicAuthUsername)
		assert.Equal(t, password, basicAuthPassword)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(dashboardJSON)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	dashboard := &v1beta1.GrafanaDashboard{
		Spec: v1beta1.GrafanaDashboardSpec{
			Url: ts.URL,
			UrlAuthorization: &v1beta1.GrafanaDashboardUrlAuthorization{
				BasicAuth: &v1beta1.GrafanaDashboardUrlBasicAuth{
					Username: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "credentials",
						},
						Key:      basicAuthUsername,
						Optional: nil,
					},
					Password: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "credentials",
						},
						Key:      basicAuthPassword,
						Optional: nil,
					},
				},
			},
		},
		Status: v1beta1.GrafanaDashboardStatus{},
	}

	credentialsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "credentials",
		},
		StringData: map[string]string{
			"USERNAME": "admin",
			"PASSWORD": "admin",
		},
	}

	err = k8sClient.Create(context.Background(), credentialsSecret)
	assert.NoError(t, err)

	fetchedDashboard, err := FetchDashboardFromUrl(context.Background(), dashboard, k8sClient, nil)
	assert.Nil(t, err)
	assert.Equal(t, dashboardJSON, fetchedDashboard, "Fetched dashboard doesn't match the original")

	assert.False(t, dashboard.Status.ContentTimestamp.Time.IsZero(), "ContentTimestamp should have been updated")
	assert.Equal(t, compressedJSON, dashboard.Status.ContentCache, "ContentCache should have been updated")
	assert.Equal(t, ts.URL, dashboard.Status.ContentUrl, "ContentUrl should have been updated")
}
