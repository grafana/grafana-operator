package grafana

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/stretchr/testify/assert"
)

func Test_getGrafanaServerProtocol(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]map[string]string
		want   string
	}{
		{
			name: "Server protocol empty",
			config: map[string]map[string]string{
				"server": {
					"protocol": "",
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server protocol http",
			config: map[string]map[string]string{
				"server": {
					"protocol": "http",
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server protocol https",
			config: map[string]map[string]string{
				"server": {
					"protocol": "https",
				},
			},
			want: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: tt.config,
				},
			}

			got := getGrafanaServerProtocol(cr)

			assert.Equal(t, tt.want, got)
		})
	}
}
