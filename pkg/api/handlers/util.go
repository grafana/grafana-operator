package handlers

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"gopkg.in/yaml.v2"
)

func grafanaFromCRD(g *v1alpha1.Grafana) *models.Grafana {

	return &models.Grafana{
		Name: &g.Name,
		Config: models.GrafanaConfig{
			AdminUser:          g.Spec.Config.Security.AdminUser,
			Hostname:           g.Spec.Ingress.Hostname,
			DisableSignoutMenu: *g.Spec.Config.Auth.DisableSignoutMenu,
			DisableLoginForm:   *g.Spec.Config.Auth.DisableLoginForm,
		},
	}
}

func getProxyHost(g *models.Grafana) (p string, err error) {
	if g.Config.Hostname == "" {
		return p, fmt.Errorf("No Grafana hostname provided")
	}
	fmt.Println("------------", *g.Name)
	n := fmt.Sprintf(defaultProxyNameFormat, *g.Name)
	p = fmt.Sprintf(g.Config.Hostname, n)
	fmt.Println(p)
	return
}

func yamlUnmarshalHandler(in interface{}, out interface{}) error {
	var tpl bytes.Buffer
	h, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	t, err := template.New("config").Parse(string(h))
	err = t.Execute(&tpl, nil)

	return yaml.Unmarshal(h, out)
}

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}
