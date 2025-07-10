package content

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetDashboardEnvs(t *testing.T) {
	dashboard := v1beta1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dashboard",
			Namespace: "grafana-operator-system",
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaContentSpec: v1beta1.GrafanaContentSpec{
				Envs: []v1beta1.GrafanaContentEnv{
					{
						Name:  "TEST_ENV",
						Value: "test-env-value",
					},
				},
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	var contentResource v1beta1.GrafanaContentResource = &dashboard
	assert.NotNil(t, contentResource.GrafanaContentSpec(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")
	assert.NotNil(t, contentResource.GrafanaContentStatus(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")

	resolver := NewContentResolver(&dashboard, k8sClient)

	envs, err := resolver.getContentEnvs(ctx)

	require.NoError(t, err)
	assert.NotNil(t, envs)
	assert.True(t, len(envs) == 1, "Expected 1 env, got %d", len(envs))
}

func TestContentIsUpdatedUID(t *testing.T) {
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
			cr := getCR(t, tt.crUID, tt.statusUID, tt.specUID, tt.dashboardUID)
			uid := CustomUIDOrUID(cr, tt.dashboardUID)

			got := IsUpdatedUID(cr, uid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getCR(t *testing.T, crUID string, statusUID string, specUID string, dashUID string) *NopContentResource {
	t.Helper()
	dashboardModel := make(map[string]any)
	dashboardModel["uid"] = dashUID
	dashboard, _ := json.Marshal(dashboardModel) //nolint:errcheck

	cr := NopContentResource{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: "grafana-operator-system",
			UID:       types.UID(crUID),
		},
		Spec: v1beta1.GrafanaContentSpec{
			CustomUID: specUID,
			JSON:      string(dashboard),
		},
		Status: v1beta1.GrafanaContentStatus{
			UID: statusUID,
		},
	}

	return &cr
}
