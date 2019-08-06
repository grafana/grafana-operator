package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	ConfigGrafanaImage              = "grafana.image.url"
	ConfigGrafanaImageTag           = "grafana.image.tag"
	ConfigPluginsInitContainerImage = "grafana.plugins.init.container.image.url"
	ConfigPluginsInitContainerTag   = "grafana.plugins.init.container.image.tag"
	ConfigOperatorNamespace         = "grafana.operator.namespace"
	ConfigDashboardLabelSelector    = "grafana.dashboard.selector"
	ConfigGrafanaPluginsUpdated     = "grafana.plugins.updated"
	ConfigOpenshift                 = "mode.openshift"
	GrafanaImage                    = "quay.io/openshift/origin-grafana"
	GrafanaVersion                  = "4.2"
	GrafanaConfigMapName            = "grafana-config"
	GrafanaConfigFileName           = "grafana.ini"
	GrafanaProvidersConfigMapName   = "grafana-providers"
	GrafanaDatasourcesConfigMapName = "grafana-datasources"
	GrafanaDashboardsConfigMapName  = "grafana-dashboards"
	GrafanaServiceAccountName       = "grafana-serviceaccount"
	GrafanaDeploymentName           = "grafana-deployment"
	GrafanaRouteName                = "grafana-route"
	GrafanaIngressName              = "grafana-ingress"
	GrafanaServiceName              = "grafana-service"
	GrafanaDataPath                 = "/var/lib/grafana"
	GrafanaLogsPath                 = "/var/log/grafana"
	GrafanaPluginsPath              = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath         = "/etc/grafana/provisioning"
	PluginsInitContainerImage       = "quay.io/integreatly/grafana_plugins_init"
	PluginsInitContainerTag         = "0.0.2"
	PluginsEnvVar                   = "GRAFANA_PLUGINS"
	PluginsUrl                      = "https://grafana.com/api/plugins/%s/versions/%s"
	PluginsMinAge                   = 5
	InitContainerName               = "grafana-plugins-init"
	ResourceFinalizerName           = "grafana.cleanup"
	RequeueDelay                    = time.Second * 15
)

type ControllerConfig struct {
	Values  map[string]interface{}
	Plugins map[string]v1alpha1.PluginList
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Values:  map[string]interface{}{},
			Plugins: map[string]v1alpha1.PluginList{},
		}
	})
	return instance
}

func (c *ControllerConfig) GetDashboardId(dashboard *v1alpha1.GrafanaDashboard) string {
	return fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Spec.Name)
}

func (c *ControllerConfig) GetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) v1alpha1.PluginList {
	return c.Plugins[c.GetDashboardId(dashboard)]
}

func (c *ControllerConfig) SetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	id := c.GetDashboardId(dashboard)
	c.Plugins[id] = dashboard.Spec.Plugins
	c.Plugins[id].SetOrigin(dashboard)
	c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
}

func (c *ControllerConfig) RemovePluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	id := c.GetDashboardId(dashboard)
	if _, ok := c.Plugins[id]; ok {
		delete(c.Plugins, id)
		c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
	}
}

func (c *ControllerConfig) AddConfigItem(key string, value interface{}) {
	if key != "" && value != nil && value != "" {
		c.Values[key] = value
	}
}

func (c *ControllerConfig) GetConfigItem(key string, defaultValue interface{}) interface{} {
	if c.HasConfigItem(key) {
		return c.Values[key]
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigString(key, defaultValue string) string {
	if c.HasConfigItem(key) {
		return c.Values[key].(string)
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigBool(key string, defaultValue bool) bool {
	if c.HasConfigItem(key) {
		return c.Values[key].(bool)
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigTimestamp(key string, defaultValue time.Time) time.Time {
	if c.HasConfigItem(key) {
		return c.Values[key].(time.Time)
	}
	return defaultValue
}

func (c *ControllerConfig) HasConfigItem(key string) bool {
	_, ok := c.Values[key]
	return ok
}
