package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestGrafana_GetScheme(t *testing.T) {
	tests := []struct {
		name string
		cr   *Grafana
		want v1.URIScheme
	}{
		{
			name: "Nil server spec",
			cr:   &Grafana{},
			want: v1.URISchemeHTTP,
		},
		{
			name: "Empty server spec",
			cr: &Grafana{
				Spec: GrafanaSpec{
					Config: GrafanaConfig{
						Server: &GrafanaConfigServer{},
					},
				},
			},
			want: v1.URISchemeHTTP,
		},
		{
			name: "HTTP in server spec",
			cr: &Grafana{
				Spec: GrafanaSpec{
					Config: GrafanaConfig{
						Server: &GrafanaConfigServer{
							Protocol: "http",
						},
					},
				},
			},
			want: v1.URISchemeHTTP,
		},
		{
			name: "HTTPS in server spec",
			cr: &Grafana{
				Spec: GrafanaSpec{
					Config: GrafanaConfig{
						Server: &GrafanaConfigServer{
							Protocol: "https",
						},
					},
				},
			},
			want: v1.URISchemeHTTPS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cr.GetScheme()
			assert.Equal(t, tt.want, got)
		})
	}
}
