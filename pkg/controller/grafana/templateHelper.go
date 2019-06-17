package grafana

import (
	"bytes"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	DefaultLogLevel = "info"
)

type GrafanaParamaeters struct {
	GrafanaImage                    string
	GrafanaVersion                  string
	Namespace                       string
	GrafanaConfigMapName            string
	GrafanaProvidersConfigMapName   string
	GrafanaDatasourcesConfigMapName string
	GrafanaDashboardsConfigMapName  string
	GrafanaServiceAccountName       string
	GrafanaDeploymentName           string
	LogLevel                        string
	GrafanaRouteName                string
	GrafanaServiceName              string
	PluginsInitContainerImage       string
	PluginsInitContainerTag         string
	GrafanaIngressName              string
	Hostname                        string
	AdminUser                       string
	AdminPassword                   string
	BasicAuth                       bool
	DisableLoginForm                bool
	DisableSignoutMenu              bool
	Anonymous                       bool
}

type TemplateHelper struct {
	Parameters   GrafanaParamaeters
	TemplatePath string
}

func option(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func getLogLevel(userLogLevel string) string {
	level := strings.TrimSpace(userLogLevel)
	level = strings.ToLower(level)

	switch level {
	case "debug":
		return level
	case "info":
		return level
	case "warn":
		return level
	case "error":
		return level
	case "critical":
		return level
	default:
		return DefaultLogLevel
	}
}

// Creates a new templates helper and populates the values for all
// templates properties. Some of them (like the hostname) are set
// by the user in the custom resource
func newTemplateHelper(cr *integreatly.Grafana) *TemplateHelper {
	controllerConfig := common.GetControllerConfig()

	param := GrafanaParamaeters{
		GrafanaImage:                    controllerConfig.GetConfigString(common.ConfigGrafanaImage, common.GrafanaImage),
		GrafanaVersion:                  controllerConfig.GetConfigString(common.ConfigGrafanaImageTag, common.GrafanaVersion),
		Namespace:                       cr.Namespace,
		GrafanaConfigMapName:            common.GrafanaConfigMapName,
		GrafanaProvidersConfigMapName:   common.GrafanaProvidersConfigMapName,
		GrafanaDatasourcesConfigMapName: common.GrafanaDatasourcesConfigMapName,
		GrafanaDashboardsConfigMapName:  common.GrafanaDashboardsConfigMapName,
		GrafanaServiceAccountName:       common.GrafanaServiceAccountName,
		GrafanaDeploymentName:           common.GrafanaDeploymentName,
		LogLevel:                        getLogLevel(cr.Spec.LogLevel),
		GrafanaRouteName:                common.GrafanaRouteName,
		GrafanaServiceName:              common.GrafanaServiceName,
		PluginsInitContainerImage:       controllerConfig.GetConfigString(common.ConfigPluginsInitContainerImage, common.PluginsInitContainerImage),
		PluginsInitContainerTag:         controllerConfig.GetConfigString(common.ConfigPluginsInitContainerTag, common.PluginsInitContainerTag),
		GrafanaIngressName:              common.GrafanaIngressName,
		Hostname:                        cr.Spec.Hostname,
		AdminUser:                       option(cr.Spec.AdminUser, "root"),
		AdminPassword:                   option(cr.Spec.AdminPassword, "secret"),
		BasicAuth:                       cr.Spec.BasicAuth,
		DisableLoginForm:                cr.Spec.DisableLoginForm,
		DisableSignoutMenu:              cr.Spec.DisableSignoutMenu,
		Anonymous:                       cr.Spec.Anonymous,
	}

	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "./templates"
	}

	return &TemplateHelper{
		Parameters:   param,
		TemplatePath: templatePath,
	}
}

// load a templates from a given resource name. The templates must be located
// under ./templates and the filename must be <resource-name>.yaml
func (h *TemplateHelper) loadTemplate(name string) ([]byte, error) {
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
