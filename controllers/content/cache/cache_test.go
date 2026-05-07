package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func TestSetContentCache(t *testing.T) {
	url1 := "http://localhost:8080/1.json"
	url2 := "http://localhost:8080/2.json"

	raw1 := map[string]any{"title": "Test1"}
	raw2 := map[string]any{"title": "Test2"}

	j1, err := json.Marshal(raw1)
	require.NoError(t, err)

	j2, err := json.Marshal(raw2)
	require.NoError(t, err)

	gz1, err := Gzip(j1)
	require.NoError(t, err)

	gz2, err := Gzip(j2)
	require.NoError(t, err)

	now := metav1.NewTime(time.Now())
	hourAgo := metav1.NewTime(time.Now().Add(-time.Hour))

	tests := []struct {
		name            string
		url             string
		raw             map[string]any // TODO: rename
		contentDuration time.Duration
		status          v1beta1.GrafanaContentStatus
		want            v1beta1.GrafanaContentStatus
	}{
		{
			name:            "no cache: cache is populated",
			url:             url1,
			raw:             raw1,
			contentDuration: time.Hour,
			status:          v1beta1.GrafanaContentStatus{},
			want: v1beta1.GrafanaContentStatus{
				ContentCache:     gz1,
				ContentTimestamp: now,
				ContentURL:       url1,
			},
		},
		{
			name:            "existing valid cache: cache is not updated",
			url:             url1,
			raw:             raw1,
			contentDuration: 24 * time.Hour,
			status: v1beta1.GrafanaContentStatus{
				ContentURL:       url1,
				ContentCache:     gz1,
				ContentTimestamp: hourAgo,
			},
			want: v1beta1.GrafanaContentStatus{
				ContentCache:     gz1,
				ContentTimestamp: hourAgo,
				ContentURL:       url1,
			},
		},
		{
			name:            "existing valid cache with wrong url: cache is updated",
			url:             url2,
			raw:             raw2,
			contentDuration: 24 * time.Hour,
			status: v1beta1.GrafanaContentStatus{
				ContentURL:       url1,
				ContentCache:     gz1,
				ContentTimestamp: hourAgo,
			},
			want: v1beta1.GrafanaContentStatus{
				ContentURL:       url2,
				ContentCache:     gz2,
				ContentTimestamp: now,
			},
		},
		{
			name:            "partially populated cache (valid URL, not expired): missing content is added",
			url:             url1,
			raw:             raw1,
			contentDuration: 24 * time.Hour,
			status: v1beta1.GrafanaContentStatus{
				ContentURL:       url1,
				ContentCache:     []byte{},
				ContentTimestamp: hourAgo,
			},
			want: v1beta1.GrafanaContentStatus{
				ContentURL:       url1,
				ContentCache:     gz1,
				ContentTimestamp: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.status.DeepCopy()

			err := setContentCache(status, tt.url, tt.raw, tt.contentDuration)
			require.NoError(t, err)

			assert.Equal(t, tt.want.ContentCache, status.ContentCache)
			assert.WithinDuration(t, tt.want.ContentTimestamp.Time, status.ContentTimestamp.Time, time.Second)
			assert.Equal(t, tt.want.ContentURL, status.ContentURL)
		})
	}
}
