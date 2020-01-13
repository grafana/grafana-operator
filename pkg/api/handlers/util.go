package handlers

import (
	"github.com/integr8ly/grafana-operator/pkg/api/models"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

func grafanaFromCRD(g *v1alpha1.Grafana) *models.Grafana {

	return &models.Grafana{
		Name: &g.Name,
		Config: models.GrafanaConfig{
			AdminUser:          g.Spec.AdminUser,
			Hostname:           g.Spec.Ingress.Hostname,
			DisableSignoutMenu: g.Spec.DisableSignoutMenu,
			DisableLoginForm:   g.Spec.DisableLoginForm,
		},
	}
}
