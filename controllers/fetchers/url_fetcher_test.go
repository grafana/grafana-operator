package fetchers

import (
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

	fetchedDashboard, err := FetchDashboardFromUrl(dashboard)
	assert.Nil(t, err)
	assert.Equal(t, dashboardJSON, fetchedDashboard, "Fetched dashboard doesn't match the original")

	assert.False(t, dashboard.Status.ContentTimestamp.Time.IsZero(), "ContentTimestamp should have been updated")
	assert.Equal(t, compressedJSON, dashboard.Status.ContentCache, "ContentCache should have been updated")
	assert.Equal(t, ts.URL, dashboard.Status.ContentUrl, "ContentUrl should have been updated")
}
