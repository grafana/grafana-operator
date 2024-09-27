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

func TestGrafanaDashboardUIDPriority(t *testing.T) {
	crUID := "crUID"
	dashUID := "dashUID"
	customUID := "customUID"
	tests := []struct {
		name         string
		crUID        string
		dashboardUID string
		customUID    string
		want         string
	}{
		{
			name:         "Fallback to crUID when customUID is empty",
			crUID:        crUID,
			dashboardUID: "",
			customUID:    "",
			want:         crUID,
		},
		{
			name:         "crUID overwrites dashboardUID",
			crUID:        crUID,
			dashboardUID: dashUID,
			customUID:    "",
			want:         crUID,
		},
		{
			name:         "customUID has priority over crUID",
			crUID:        crUID,
			dashboardUID: "",
			customUID:    customUID,
			want:         customUID,
		},
		{
			name:         "customUID trumps crUID and overwrites dashboardUID",
			crUID:        crUID,
			dashboardUID: dashUID,
			customUID:    customUID,
			want:         customUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := getDashboardCR(t, tt.crUID, tt.crUID, tt.customUID)
			got := cr.CustomUIDOrUID()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrafanaDashboardIsUpdatedUID(t *testing.T) {
	crUID := "crUID"
	dashUID := "dashUID"
	customUID := "customUID"
	tests := []struct {
		name         string
		crUID        string
		statusUID    string
		dashboardUID string
		customUID    string
		want         bool
	}{
		// Validate that dashboardUID is always ignored despite the presence of customUID
		// dashboardUID is always overwritten by customUID which falls back to crUID
		{
			name:         "DashboardUID set and ignored customUID empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: dashUID,
			customUID:    "",
			want:         false,
		},
		{
			name:         "DashboardUID set and ignored customUID set",
			crUID:        crUID,
			statusUID:    customUID,
			dashboardUID: dashUID,
			customUID:    customUID,
			want:         false,
		},
		{
			name:         "DashboardUID empty and ignored customUID empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: "",
			customUID:    "",
			want:         false,
		},
		{
			name:         "DashboardUID empty and ignored customUID set",
			crUID:        crUID,
			statusUID:    customUID,
			dashboardUID: "",
			customUID:    customUID,
			want:         false,
		},
		// Validate always false when statusUID is empty
		// Since dashbaordUID is ignoredk the only variable is customUID
		{
			name:         "Empty StatusUID always results in false",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "",
			customUID:    "",
			want:         false,
		},
		{
			name:         "Always false when statusUID is empty regardless of customUID or dashUID being set",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "",
			customUID:    customUID,
			want:         false,
		},
		// Validate impossible true cases (Backwards compatible before immutable UIDs)
		// NOTE Could be deleted in v1 when dashboard UID is immutable
		{
			name:         "IMPOSSIBLE: Old status with no customUID",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			customUID:    "",
			want:         true,
		},
		{
			name:         "IMPOSSIBLE: Old status with customUID",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			customUID:    customUID,
			want:         true,
		},
		{
			name:         "IMPOSSIBLE: Old Status with all UIDs being equal",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			customUID:    crUID,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := getDashboardCR(t, tt.crUID, tt.statusUID, tt.customUID)
			got := cr.IsUpdatedUID()
			assert.Equal(t, tt.want, got)
		})
	}
}

func getDashboardCR(t *testing.T, crUID string, statusUID string, specUID string) GrafanaDashboard {
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
			CustomUID: specUID,
		},
		Status: GrafanaDashboardStatus{
			UID: statusUID,
		},
	}

	return cr
}
