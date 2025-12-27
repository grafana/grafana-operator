package content

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestIsUpdatedUID(t *testing.T) {
	const (
		contentUID = "contentUID"
		crUID      = "crUID"
		customUID  = "customUID"
	)

	tests := []struct {
		name       string
		crUID      string
		statusUID  string
		contentUID string
		customUID  string
		want       bool
	}{
		//
		// Returns false when uid in status (statusUID) is not set
		//
		{
			name:       "No UID in status (no contentUID/customUID overrides)",
			crUID:      crUID,
			statusUID:  "",
			contentUID: "",
			customUID:  "",
			want:       false,
		},
		{
			name:       "No UID in status (contentUID is set)",
			crUID:      crUID,
			statusUID:  "",
			contentUID: contentUID,
			customUID:  "",
			want:       false,
		},
		{
			name:       "No UID in status (customUID is set)",
			crUID:      crUID,
			statusUID:  "",
			contentUID: "",
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "No UID in status (customUID and contentUID are set)",
			crUID:      crUID,
			statusUID:  "",
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		// Validate that crUID is always overwritten by contentUID or customUID
		// contentUID is always overwritten by customUID which falls back to crUID
		{
			name:       "contentUID and customUID empty",
			crUID:      crUID,
			statusUID:  crUID,
			contentUID: "",
			customUID:  "",
			want:       false,
		},
		{
			name:       "contentUID set and customUID empty",
			crUID:      crUID,
			statusUID:  contentUID,
			contentUID: contentUID,
			customUID:  "",
			want:       false,
		},
		{
			name:       "contentUID set and customUID set",
			crUID:      crUID,
			statusUID:  customUID,
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "contentUID empty and customUID set",
			crUID:      crUID,
			statusUID:  customUID,
			contentUID: "",
			customUID:  customUID,
			want:       false,
		},
		// Validate updates are detected correctly
		{
			name:       "contentUID updated and customUID empty",
			crUID:      crUID,
			statusUID:  crUID,
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "contentUID updated and customUID set",
			crUID:      crUID,
			statusUID:  customUID,
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "new contentUID and no customUID",
			crUID:      crUID,
			statusUID:  "oldUID",
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "contentUID removed and no customUID",
			crUID:      crUID,
			statusUID:  "oldUID",
			contentUID: "",
			customUID:  "",
			want:       true,
		},
		// Validate that statusUID detection works even in impossible cases expecting cr or customUID to change
		{
			name:       "IMPOSSIBLE: Old status with new customUID",
			crUID:      crUID,
			statusUID:  "oldUID",
			contentUID: "",
			customUID:  customUID,
			want:       true,
		},
		{
			name:       "IMPOSSIBLE: Old Status with all UIDs being equal",
			crUID:      crUID,
			statusUID:  "oldUID",
			contentUID: crUID,
			customUID:  crUID,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := map[string]any{
				"uid": tt.contentUID,
			}

			dashboard, err := json.Marshal(model)
			require.NoError(t, err)

			cr := &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					UID: types.UID(tt.crUID),
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						CustomUID: tt.customUID,
						JSON:      string(dashboard),
					},
				},
				Status: v1beta1.GrafanaDashboardStatus{
					GrafanaContentStatus: v1beta1.GrafanaContentStatus{
						UID: tt.statusUID,
					},
				},
			}

			got := IsUpdatedUID(cr, tt.contentUID)
			assert.Equal(t, tt.want, got)
		})
	}
}
