package grafana

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	GrafanaImage                    = "docker.io/grafana/grafana"
	GrafanaVersion                  = "5.4.2"
	LogLevel                        = "error"
	GrafanaConfigMapName            = "grafana-config"
	GrafanaProvidersConfigMapName   = "grafana-providers"
	GrafanaDatasourcesConfigMapName = "grafana-datasources"
	GrafanaDashboardsConfigMapName  = "grafana-dashboards"
	GrafanaServiceAccountName       = "grafana-serviceaccount"
	GrafanaDeploymentName           = "grafana-deployment"
	GrafanaRouteName                = "grafana-route"
	GrafanaServiceName              = "grafana-service"
)

type GrafanaParamaeters struct {
	GrafanaImage                    string
	GrafanaVersion                  string
	Namespace                       string
	GrafanaConfigMapName            string
	GrafanaProvidersConfigMapName   string
	GrafanaDatasourcesConfigMapName string
	PrometheusUrl                   string
	GrafanaDashboardsConfigMapName  string
	GrafanaServiceAccountName       string
	GrafanaDeploymentName           string
	LogLevel                        string
	GrafanaRouteName                string
	GrafanaServiceName              string
}

type GrafanaTemplateHelper struct {
	Parameters   GrafanaParamaeters
	TemplatePath string
}

// Creates a new templates helper and populates the values for all
// templates properties. Some of them (like the hostname) are set
// by the user in the custom resource
func newTemplateHelper(cr *integreatly.Grafana) *GrafanaTemplateHelper {
	param := GrafanaParamaeters{
		GrafanaImage:                    GrafanaImage,
		GrafanaVersion:                  GrafanaVersion,
		Namespace:                       cr.Namespace,
		GrafanaConfigMapName:            GrafanaConfigMapName,
		GrafanaProvidersConfigMapName:   GrafanaProvidersConfigMapName,
		GrafanaDatasourcesConfigMapName: GrafanaDatasourcesConfigMapName,
		PrometheusUrl:                   cr.Spec.PrometheusUrl,
		GrafanaDashboardsConfigMapName:  GrafanaDashboardsConfigMapName,
		GrafanaServiceAccountName:       GrafanaServiceAccountName,
		GrafanaDeploymentName:           GrafanaDeploymentName,
		LogLevel:                        LogLevel,
		GrafanaRouteName:                GrafanaRouteName,
		GrafanaServiceName:              GrafanaServiceName,
	}

	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "./templates"
	}

	return &GrafanaTemplateHelper{
		Parameters:   param,
		TemplatePath: templatePath,
	}
}

// load a templates from a given resource name. The templates must be located
// under ./templates and the filename must be <resource-name>.yaml
func (h *GrafanaTemplateHelper) loadTemplate(name string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.yaml", h.TemplatePath, name)
	tpl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := template.New("grafana").Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = parsed.Execute(&buffer, h.Parameters)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}


