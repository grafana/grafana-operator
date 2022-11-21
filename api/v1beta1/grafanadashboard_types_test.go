package v1beta1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaDashboardStatus_getContentCache(t *testing.T) {
	timestamp := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	infinite := 0 * time.Second
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)

	cachedDashboard, err := Gzip(dashboardJSON)
	assert.Nil(t, err)

	url := "http://127.0.0.1:8080/1.json"

	// Correctly populated cache
	status := GrafanaDashboardStatus{
		ContentCache:     cachedDashboard,
		ContentTimestamp: timestamp,
		ContentUrl:       url,
	}

	// Corrupted cache
	statusCorrupted := GrafanaDashboardStatus{
		ContentCache:     []byte("abc"),
		ContentTimestamp: timestamp,
		ContentUrl:       url,
	}

	tests := []struct {
		name     string
		status   GrafanaDashboardStatus
		url      string
		duration time.Duration
		want     []byte
	}{
		{
			name:     "no cache: fields are not populated",
			url:      status.ContentUrl,
			duration: infinite,
			status:   GrafanaDashboardStatus{},
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
			url:      status.ContentUrl,
			duration: 1 * time.Minute,
			status:   status,
			want:     []byte{},
		},
		{
			name:     "no cache: corrupted gzip",
			url:      statusCorrupted.ContentUrl,
			duration: infinite,
			status:   statusCorrupted,
			want:     []byte{},
		},
		{
			name:     "valid cache: not expired yet",
			url:      status.ContentUrl,
			duration: 24 * time.Hour,
			status:   status,
			want:     dashboardJSON,
		},
		{
			name:     "valid cache: not expired yet (infinite)",
			url:      status.ContentUrl,
			duration: infinite,
			status:   status,
			want:     dashboardJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.getContentCache(tt.url, tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Gzip(t *testing.T) {
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)
	compressed, err := Gzip(dashboardJSON)
	assert.Nil(t, err, "Failed to compress a dashboard")

	decompressed, err := Gunzip(compressed)
	assert.Nil(t, err, "Failed to decompress a dashboard")

	assert.Equal(t, dashboardJSON, decompressed, "Decompressed dashboard should match the original")
}
