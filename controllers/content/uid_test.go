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
		//
		// Returns false as the value in status already reflects the highest precedence
		// (spec.CustomUID -> contentUID -> metadata.uid)
		//
		{
			name:       "metadata.uid in status (no contentUID/customUID overrides)",
			crUID:      crUID,
			statusUID:  crUID,
			contentUID: "",
			customUID:  "",
			want:       false,
		},
		{
			name:       "contentUID in status (no customUID)",
			crUID:      crUID,
			statusUID:  contentUID,
			contentUID: contentUID,
			customUID:  "",
			want:       false,
		},
		{
			name:       "customUID in status (contentUID is set)",
			crUID:      crUID,
			statusUID:  customUID,
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "customUID in status (same customUID is set)",
			crUID:      crUID,
			statusUID:  customUID,
			contentUID: "",
			customUID:  customUID,
			want:       false,
		},
		//
		// Returns true as the higher precedence value is now set
		//
		{
			name:       ".metadata.uid in status, contentUID got added (no customUID)",
			crUID:      crUID,
			statusUID:  crUID,
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "old contentUID in status, contentUID has changed (no customUID)",
			crUID:      crUID,
			statusUID:  "oldContentUID",
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "old contentUID in status, contentUID got removed (no customUID)",
			crUID:      crUID,
			statusUID:  "oldContentUID",
			contentUID: "",
			customUID:  "",
			want:       true,
		},
		//
		// Returns true for changes in customUID (the field is marked as immutable through CEL,
		// so the scenario is unlikely to happen)
		//
		{
			name:       "old customUID value in status, customUID has changed (no contentUID)",
			crUID:      crUID,
			statusUID:  "oldCustomUID",
			contentUID: "",
			customUID:  customUID,
			want:       true,
		},
		{
			name:       "old customUID value in status, customUID has changed (contentUID is set)",
			crUID:      crUID,
			statusUID:  "oldCustomUID",
			contentUID: contentUID,
			customUID:  customUID,
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
