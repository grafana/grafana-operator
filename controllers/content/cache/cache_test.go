package cache

import (
	"testing"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompressDecompress(t *testing.T) {
	contentJSON := []byte(`{"dummyField": "dummyData"}`)

	compressed, err := Gzip(contentJSON)
	require.NoError(t, err)

	decompressed, err := Gunzip(compressed)
	require.NoError(t, err)

	require.Equal(t, contentJSON, decompressed)
}

func TestGrafanaDashboardStatus_getContentCache(t *testing.T) {
	timestamp := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	infinite := 0 * time.Second
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)

	cachedDashboard, err := Gzip(dashboardJSON)
	require.NoError(t, err)

	url := "http://127.0.0.1:8080/1.json"

	// Correctly populated cache
	status := v1beta1.GrafanaContentStatus{
		ContentCache:     cachedDashboard,
		ContentTimestamp: timestamp,
		ContentURL:       url,
	}

	// Corrupted cache
	statusCorrupted := v1beta1.GrafanaContentStatus{
		ContentCache:     []byte("abc"),
		ContentTimestamp: timestamp,
		ContentURL:       url,
	}

	tests := []struct {
		name     string
		status   v1beta1.GrafanaContentStatus
		url      string
		duration time.Duration
		want     []byte
	}{
		{
			name:     "no cache: fields are not populated",
			url:      status.ContentURL,
			duration: infinite,
			status:   v1beta1.GrafanaContentStatus{},
			want:     []byte{},
		},
		{
			name:     "no cache: url is different",
			url:      "http://another-url/2.json",
			duration: infinite,
			status:   status,
			want:     []byte{},
		},
		{
			name:     "no cache: expired",
			url:      status.ContentURL,
			duration: 1 * time.Minute,
			status:   status,
			want:     []byte{},
		},
		{
			name:     "no cache: corrupted gzip",
			url:      statusCorrupted.ContentURL,
			duration: infinite,
			status:   statusCorrupted,
			want:     []byte{},
		},
		{
			name:     "valid cache: not expired yet",
			url:      status.ContentURL,
			duration: 24 * time.Hour,
			status:   status,
			want:     dashboardJSON,
		},
		{
			name:     "valid cache: not expired yet (infinite)",
			url:      status.ContentURL,
			duration: infinite,
			status:   status,
			want:     dashboardJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getContentCache(&tt.status, tt.url, tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}
