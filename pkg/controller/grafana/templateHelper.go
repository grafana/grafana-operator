package grafana

import (
	"bytes"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"text/template"

	"github.com/integr8ly/grafana-operator/pkg/controller/common"

	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"
)

// GrafanaParameters provides the context for the template
type GrafanaParameters struct {
	AdminPassword                   string
	AdminUser                       string
	Anonymous                       bool
	BasicAuth                       bool
	DisableLoginForm                bool
	DisableSignoutMenu              bool
	GrafanaConfigHash               string
	GrafanaConfigMapName            string
	GrafanaDashboardsConfigMapName  string
	GrafanaDatasourcesConfigMapName string
	GrafanaDeploymentName           string
	GrafanaImage                    string
	GrafanaIngressAnnotations       map[string]string
	GrafanaIngressLabels            map[string]string
	GrafanaIngressName              string
	GrafanaIngressPath              string
	GrafanaProvidersConfigMapName   string
	GrafanaRouteName                string
	GrafanaServiceAccountName       string
	GrafanaServiceAnnotations       map[string]string
	GrafanaServiceLabels            map[string]string
	GrafanaServiceName              string
	GrafanaServiceType              string
	GrafanaVersion                  string
	Hostname                        string
	LogLevel                        string
	Namespace                       string
	NumberOfDashboardCMs            int
	PluginsInitContainerImage       string
	PluginsInitContainerTag         string
	PodLabelValue                   string
	Replicas                        int
}

// TemplateHelper is the deployment helper object
type TemplateHelper struct {
	Parameters   GrafanaParameters
	TemplatePath string
}

func option(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func getServiceType(serviceType string) string {
	switch v1.ServiceType(strings.TrimSpace(serviceType)) {
	case v1.ServiceTypeClusterIP:
		return serviceType
	case v1.ServiceTypeNodePort:
		return serviceType
	case v1.ServiceTypeLoadBalancer:
		return serviceType
	default:
		return common.DefaultServiceType
	}
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
		return common.DefaultLogLevel
	}
}

// Creates a new templates helper and populates the values for all
// templates properties. Some of them (like the hostname) are set
// by the user in the custom resource
func newTemplateHelper(cr *integreatly.Grafana) *TemplateHelper {
	controllerConfig := common.GetControllerConfig()

	param := GrafanaParameters{
		AdminPassword:                   option(cr.Spec.AdminPassword, "secret"),
		AdminUser:                       option(cr.Spec.AdminUser, "root"),
		Anonymous:                       cr.Spec.Anonymous,
		BasicAuth:                       cr.Spec.BasicAuth,
		DisableLoginForm:                cr.Spec.DisableLoginForm,
		DisableSignoutMenu:              cr.Spec.DisableSignoutMenu,
		GrafanaConfigHash:               cr.Status.LastConfig,
		GrafanaConfigMapName:            common.GrafanaConfigMapName,
		GrafanaDashboardsConfigMapName:  common.GrafanaDashboardsConfigMapName,
		GrafanaDatasourcesConfigMapName: common.GrafanaDatasourcesConfigMapName,
		GrafanaDeploymentName:           common.GrafanaDeploymentName,
		GrafanaImage:                    controllerConfig.GetConfigString(common.ConfigGrafanaImage, common.GrafanaImage),
		GrafanaIngressAnnotations:       cr.Spec.Ingress.Annotations,
		GrafanaIngressLabels:            cr.Spec.Ingress.Labels,
		GrafanaIngressName:              common.GrafanaIngressName,
		GrafanaIngressPath:              cr.Spec.Ingress.Path,
		GrafanaProvidersConfigMapName:   common.GrafanaProvidersConfigMapName,
		GrafanaRouteName:                common.GrafanaRouteName,
		GrafanaServiceAccountName:       common.GrafanaServiceAccountName,
		GrafanaServiceAnnotations:       cr.Spec.Service.Annotations,
		GrafanaServiceLabels:            cr.Spec.Service.Labels,
		GrafanaServiceName:              common.GrafanaServiceName,
		GrafanaServiceType:              cr.Spec.Service.Type,
		GrafanaVersion:                  controllerConfig.GetConfigString(common.ConfigGrafanaImageTag, common.GrafanaVersion),
		Hostname:                        cr.Spec.Ingress.Hostname,
		LogLevel:                        getLogLevel(cr.Spec.LogLevel),
		Namespace:                       cr.Namespace,
		PluginsInitContainerImage:       controllerConfig.GetConfigString(common.ConfigPluginsInitContainerImage, common.PluginsInitContainerImage),
		PluginsInitContainerTag:         controllerConfig.GetConfigString(common.ConfigPluginsInitContainerTag, common.PluginsInitContainerTag),
		PodLabelValue:                   controllerConfig.GetConfigString(common.ConfigPodLabelValue, common.PodLabelDefaultValue),
		Replicas:                        cr.Spec.InitialReplicas,
		NumberOfDashboardCMs:            controllerConfig.GetConfigInt(common.ConfigNumberOfDashboardsCMs, common.NumOfDashboardCMsDefaultValue),
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

func N(number int) (stream chan int) {
	stream = make(chan int)
	go func() {
		for i := 0; i < number; i++ {
			stream <- i
		}
		close(stream)
	}()
	return
}

// load a templates from a given resource name. The templates must be located
// under ./templates and the filename must be <resource-name>.yaml
func (h *TemplateHelper) loadTemplate(name string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.yaml", h.TemplatePath, name)
	tpl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := template.New("grafana").Funcs(template.FuncMap{"N": N}).Parse(string(tpl))
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
