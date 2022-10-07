package grafana

import (
	"testing"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/common"
	routev1 "github.com/openshift/api/route/v1"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

func TestReconcileGrafana_getGrafanaAdminUrl(t *testing.T) {
	r := &ReconcileGrafana{}

	/*
		Service is NOT preferred
	*/

	t.Run("Route", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{}

		state := &common.ClusterState{
			GrafanaRoute: &routev1.Route{
				Spec: routev1.RouteSpec{
					Host: "route",
				},
			},
		}
		want := "https://route"

		got, err := r.getGrafanaAdminUrl(cr, state)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Ingress", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{
			Spec: grafanav1alpha1.GrafanaSpec{
				Ingress: &grafanav1alpha1.GrafanaIngress{
					Hostname: "ingress",
				},
			},
		}

		state := &common.ClusterState{
			GrafanaIngress: &netv1.Ingress{},
		}

		want := "https://ingress"

		got, err := r.getGrafanaAdminUrl(cr, state)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Ingress with empty spec, LB with Hostname", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{
			Spec: grafanav1alpha1.GrafanaSpec{
				Ingress: &grafanav1alpha1.GrafanaIngress{},
			},
		}

		state := &common.ClusterState{
			GrafanaIngress: &netv1.Ingress{
				Status: netv1.IngressStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: "lbhostname",
							},
						},
					},
				},
			},
		}

		want := "https://lbhostname"

		got, err := r.getGrafanaAdminUrl(cr, state)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Ingress with empty spec, LB with IP", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{
			Spec: grafanav1alpha1.GrafanaSpec{
				Ingress: &grafanav1alpha1.GrafanaIngress{},
			},
		}

		state := &common.ClusterState{
			GrafanaIngress: &netv1.Ingress{
				Status: netv1.IngressStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: "1.2.3.4",
							},
						},
					},
				},
			},
		}

		want := "https://1.2.3.4"

		got, err := r.getGrafanaAdminUrl(cr, state)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("Empty specs", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{}

		state := &common.ClusterState{}

		_, err := r.getGrafanaAdminUrl(cr, state)
		assert.NotNil(t, err)
	})

	/*
		Service IS preferred
	*/

	type Srv = grafanav1alpha1.GrafanaConfigServer

	testsPreferService := []struct {
		name     string
		server   *Srv
		want     string
		wantFail bool
	}{
		{
			name:     "server spec is nil",
			server:   nil,
			want:     "http://grafana.monitoring:3000",
			wantFail: false,
		},
		{
			name:     "server protocol: not specified",
			server:   &Srv{Protocol: ""},
			want:     "http://grafana.monitoring:3000",
			wantFail: false,
		},
		{
			name:     "server protocol: http",
			server:   &Srv{Protocol: "http"},
			want:     "http://grafana.monitoring:3000",
			wantFail: false,
		},
		{
			name:     "server protocol: https",
			server:   &Srv{Protocol: "https"},
			want:     "https://grafana.monitoring:3000",
			wantFail: false,
		},
		{
			name:     "server protocol: h2",
			server:   &Srv{Protocol: "h2"},
			want:     "",
			wantFail: true,
		},
		{
			name:     "server protocol: socket",
			server:   &Srv{Protocol: "socket"},
			want:     "",
			wantFail: true,
		},
	}

	for _, tt := range testsPreferService {
		preferService := true

		t.Run(tt.name, func(t *testing.T) {
			cr := &grafanav1alpha1.Grafana{
				Spec: grafanav1alpha1.GrafanaSpec{
					// Ingress is set only to make sure PreferService is respected
					Ingress: &grafanav1alpha1.GrafanaIngress{
						Hostname: "ingress",
					},
					Client: &grafanav1alpha1.GrafanaClient{
						PreferService: &preferService,
					},
					Config: grafanav1alpha1.GrafanaConfig{
						Server: tt.server,
					},
				},
			}
			cr.Namespace = "monitoring"

			state := &common.ClusterState{
				// GrafanaRoute is set only to make sure PreferService is respected
				GrafanaRoute: &routev1.Route{
					Spec: routev1.RouteSpec{
						Host: "route",
					},
				},
				GrafanaService: &v1.Service{},
			}
			state.GrafanaService.Name = "grafana"

			got, err := r.getGrafanaAdminUrl(cr, state)

			if tt.wantFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
