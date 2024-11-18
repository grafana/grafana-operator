package v1beta1

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	dashUID := "dashUID"
	specUID := "specUID"
	tests := []struct {
		name         string
		crUID        string
		statusUID    string
		dashboardUID string
		specUID      string
		want         bool
	}{
		// Validate always false when statusUID is empty
		// Since dashboardUID is ignoredk the only variable is customUID
		{
			name:         "Empty StatusUID always results in false",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "",
			specUID:      "",
			want:         false,
		},
		{
			name:         "Always false when statusUID is empty regardless of dashUID being set",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: dashUID,
			specUID:      "",
			want:         false,
		},
		{
			name:         "Always false when statusUID is empty regardless of customUID being set",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "",
			specUID:      specUID,
			want:         false,
		},
		{
			name:         "Always false when statusUID is empty regardless of customUID or dashUID being set",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: dashUID,
			specUID:      specUID,
			want:         false,
		},
		// Validate that crUID is always overwritten by dashUID or customUID
		// dashboardUID is always overwritten by customUID which falls back to crUID
		{
			name:         "DashboardUID and customUID empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: "",
			specUID:      "",
			want:         false,
		},
		{
			name:         "DashboardUID set and customUID empty",
			crUID:        crUID,
			statusUID:    dashUID,
			dashboardUID: dashUID,
			specUID:      "",
			want:         false,
		},
		{
			name:         "DashboardUID set and customUID set",
			crUID:        crUID,
			statusUID:    specUID,
			dashboardUID: dashUID,
			specUID:      specUID,
			want:         false,
		},
		{
			name:         "DashboardUID empty and customUID set",
			crUID:        crUID,
			statusUID:    specUID,
			dashboardUID: "",
			specUID:      specUID,
			want:         false,
		},
		// Validate updates are detected correctly
		{
			name:         "DashboardUID updated and customUID empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: dashUID,
			specUID:      "",
			want:         true,
		},
		{
			name:         "DashboardUID updated and customUID set",
			crUID:        crUID,
			statusUID:    specUID,
			dashboardUID: dashUID,
			specUID:      specUID,
			want:         false,
		},
		{
			name:         "new dashUID and no customUID",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: dashUID,
			specUID:      "",
			want:         true,
		},
		{
			name:         "dashUID removed and no customUID",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			specUID:      "",
			want:         true,
		},
		// Validate that statusUID detection works even in impossible cases expecting cr or customUID to change
		{
			name:         "IMPOSSIBLE: Old status with new customUID",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			specUID:      specUID,
			want:         true,
		},
		{
			name:         "IMPOSSIBLE: Old Status with all UIDs being equal",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: crUID,
			specUID:      crUID,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := getDashboardCR(t, tt.crUID, tt.statusUID, tt.specUID, tt.dashboardUID)
			uid := cr.CustomUIDOrUID(tt.dashboardUID)

			got := cr.IsUpdatedUID(uid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getDashboardCR(t *testing.T, crUID string, statusUID string, specUID string, dashUID string) GrafanaDashboard {
	t.Helper()
	var dashboardModel map[string]interface{} = make(map[string]interface{})
	dashboardModel["uid"] = dashUID
	dashboard, _ := json.Marshal(dashboardModel) //nolint:errcheck

	cr := GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: "grafana-operator-system",
			UID:       types.UID(crUID),
		},
		Spec: GrafanaDashboardSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"dashboard": "grafana",
					},
				},
			},
			CustomUID: specUID,
			Json:      string(dashboard),
		},
		Status: GrafanaDashboardStatus{
			UID: statusUID,
		},
	}

	return cr
}

func newDashboard(name string, uid string) *GrafanaDashboard {
	return &GrafanaDashboard{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDashboard",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDashboardSpec{
			CustomUID: uid,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			Json: "",
		},
	}
}

var _ = Describe("Dashboard type", func() {
	Context("Ensure Dashboard spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newDashboard("missing-uid", "")
			By("Create new Dashboard without uid")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Adding a uid")
			dash.Spec.CustomUID = "new-dash-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			dash := newDashboard("existing-uid", "existing-uid")
			By("Creating Dashboard with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("And setting UID to ''")
			dash.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			dash := newDashboard("removing-uid", "existing-uid")
			By("Create new Dashboard with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Changing the existing UID")
			dash.Spec.CustomUID = "new-dash-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})
	})
})
