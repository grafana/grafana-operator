package v1beta1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func TestGrafanaDashboardIsUpdatedUID(t *testing.T) {
	crUID := "crUID"
	tests := []struct {
		name         string
		crUID        string
		statusUID    string
		dashboardUID string
		want         bool
	}{
		{
			name:         "Status.UID and dashboard UID are empty",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "newUID",
			want:         false,
		},
		{
			name:         "Status.UID is empty, dashboard UID is not",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "newUID",
			want:         false,
		},
		{
			name:         "Status.UID is not empty (same as CR uid), new UID is empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: "",
			want:         false,
		},
		{
			name:         "Status.UID is not empty (different from CR uid), new UID is empty",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			want:         true,
		},
		{
			name:         "Status.UID is not empty, new UID is different",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "newUID",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := getDashboardCR(t, tt.crUID, tt.statusUID)
			got := cr.IsUpdatedUID(tt.dashboardUID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getDashboardCR(t *testing.T, crUID string, statusUID string) GrafanaDashboard {
	t.Helper()

	cr := GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: "grafana-operator-system",
			UID:       types.UID(crUID),
		},
		Spec: GrafanaDashboardSpec{
			InstanceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"dashboard": "grafana",
				},
			},
		},
		Status: GrafanaDashboardStatus{
			UID: statusUID,
		},
	}

	return cr
}
