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
		customUID  = "customUID"
		metaUID    = "metaUID"
	)

	tests := []struct {
		name       string
		metaUID    string
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
			metaUID:    metaUID,
			statusUID:  "",
			contentUID: "",
			customUID:  "",
			want:       false,
		},
		{
			name:       "No UID in status (contentUID is set)",
			metaUID:    metaUID,
			statusUID:  "",
			contentUID: contentUID,
			customUID:  "",
			want:       false,
		},
		{
			name:       "No UID in status (customUID is set)",
			metaUID:    metaUID,
			statusUID:  "",
			contentUID: "",
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "No UID in status (customUID and contentUID are set)",
			metaUID:    metaUID,
			statusUID:  "",
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		//
		// Returns false as the value in status already reflects the highest precedence
		// (customUID -> contentUID -> metaUID)
		//
		{
			name:       "metaUID in status (no contentUID/customUID overrides)",
			metaUID:    metaUID,
			statusUID:  metaUID,
			contentUID: "",
			customUID:  "",
			want:       false,
		},
		{
			name:       "contentUID in status (no customUID)",
			metaUID:    metaUID,
			statusUID:  contentUID,
			contentUID: contentUID,
			customUID:  "",
			want:       false,
		},
		{
			name:       "customUID in status (contentUID is set)",
			metaUID:    metaUID,
			statusUID:  customUID,
			contentUID: contentUID,
			customUID:  customUID,
			want:       false,
		},
		{
			name:       "customUID in status (same customUID is set)",
			metaUID:    metaUID,
			statusUID:  customUID,
			contentUID: "",
			customUID:  customUID,
			want:       false,
		},
		//
		// Returns true as the higher precedence value is now set
		//
		{
			name:       ".metaUID in status, contentUID got added (no customUID)",
			metaUID:    metaUID,
			statusUID:  metaUID,
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "old contentUID in status, contentUID has changed (no customUID)",
			metaUID:    metaUID,
			statusUID:  "oldContentUID",
			contentUID: contentUID,
			customUID:  "",
			want:       true,
		},
		{
			name:       "old contentUID in status, contentUID got removed (no customUID)",
			metaUID:    metaUID,
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
			metaUID:    metaUID,
			statusUID:  "oldCustomUID",
			contentUID: "",
			customUID:  customUID,
			want:       true,
		},
		{
			name:       "old customUID value in status, customUID has changed (contentUID is set)",
			metaUID:    metaUID,
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
					UID: types.UID(tt.metaUID),
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
