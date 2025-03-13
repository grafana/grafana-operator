package grafana

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/stretchr/testify/assert"
)

func Test_getGrafanaServerProtocol(t *testing.T) {
	tests := []struct {
		name string
		cr   *v1beta1.Grafana
		want string
	}{
		{
			name: "Config nil",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: nil,
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server nil",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: map[string]map[string]string{
						"server": nil,
					},
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server protocol empty",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: map[string]map[string]string{
						"server": {
							"protocol": "",
						},
					},
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server protocol http",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: map[string]map[string]string{
						"server": {
							"protocol": "http",
						},
					},
				},
			},
			want: config.GrafanaServerProtocol,
		},
		{
			name: "Server protocol https",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: map[string]map[string]string{
						"server": {
							"protocol": "https",
						},
					},
				},
			},
			want: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			got := getGrafanaServerProtocol(tt.cr)

			assert.Equal(t, want, got)
		})
	}
}
