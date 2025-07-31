package grafana

import (
	"fmt"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"

	"github.com/stretchr/testify/assert"
)

func TestGetGrafanaImage(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "not specified(default version)",
			version: "",
			want:    fmt.Sprintf("%s:%s", config.GrafanaImage, config.GrafanaVersion),
		},
		{
			name:    "custom tag",
			version: "10.4.0",
			want:    fmt.Sprintf("%s:10.4.0", config.GrafanaImage),
		},
		{
			name:    "fully-qualified image",
			version: "docker.io/grafana/grafana@sha256:b7fcb534f7b3512801bb3f4e658238846435804deb479d105b5cdc680847c272",
			want:    "docker.io/grafana/grafana@sha256:b7fcb534f7b3512801bb3f4e658238846435804deb479d105b5cdc680847c272",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Version: tt.version,
				},
			}

			got := getGrafanaImage(cr)

			assert.Equal(t, tt.want, got)
		})
	}
}
