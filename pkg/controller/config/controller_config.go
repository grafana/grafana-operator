package config

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
	ConfigPodLabelValue             = "grafana.pod.label"
	ConfigOperatorNamespace         = "grafana.operator.namespace"
	ConfigDashboardLabelSelector    = "grafana.dashboard.selector"
	ConfigGrafanaPluginsUpdated     = "grafana.plugins.updated"
	ConfigOpenshift                 = "mode.openshift"
	GrafanaImage                    = "grafana/grafana"
	GrafanaVersion                  = "6.4.4"
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
	PodLabelDefaultValue            = "grafana"
	DefaultServiceType              = "ClusterIP"
	DefaultLogLevel                 = "info"
	SecretsMountDir                 = "/etc/grafana-secrets/"
	ConfigMapsMountDir              = "/etc/grafana-configmaps/"
	ConfigRouteWatch                = "watch.routes"
	ConfigGrafanaAdminUsername      = "grafana.admin.username"
	ConfigGrafanaAdminPassword      = "grafana.admin.password"
	ConfigGrafanaAdminRoute         = "grafana.route"
	ConfigGrafanaDashboardsSynced   = "grafana.dashboards.synced"
)

type ControllerConfig struct {
	*sync.Mutex
	Values     map[string]interface{}
	Plugins    map[string]v1alpha1.PluginList
	Dashboards map[string][]v1alpha1.GrafanaDashboardRef
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Mutex:      &sync.Mutex{},
			Values:     map[string]interface{}{},
			Plugins:    map[string]v1alpha1.PluginList{},
			Dashboards: map[string][]v1alpha1.GrafanaDashboardRef{},
		}
	})
	return instance
}

func (c *ControllerConfig) GetDashboardId(namespace, name string) string {
	return fmt.Sprintf("%v/%v", namespace, name)
}

func (c *ControllerConfig) GetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) v1alpha1.PluginList {
	c.Lock()
	defer c.Unlock()
	return c.Plugins[c.GetDashboardId(dashboard.Namespace, dashboard.Name)]
}

func (c *ControllerConfig) SetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	id := c.GetDashboardId(dashboard.Namespace, dashboard.Name)
	c.Plugins[id] = dashboard.Spec.Plugins
	c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
}

func (c *ControllerConfig) RemovePluginsFor(namespace, name string) {
	id := c.GetDashboardId(namespace, name)
	if _, ok := c.Plugins[id]; ok {
		delete(c.Plugins, id)
		c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
	}
}

func (c *ControllerConfig) AddDashboard(dashboard *v1alpha1.GrafanaDashboard) {
	ns := dashboard.Namespace
	if _, exists := c.HasDashboard(ns, dashboard.Name); !exists {
		c.Lock()
		defer c.Unlock()
		c.Dashboards[ns] = append(c.Dashboards[ns], v1alpha1.GrafanaDashboardRef{
			Name: dashboard.Name,
			UID:  dashboard.Status.UID,
		})
	}
}

func (c *ControllerConfig) RemoveDashboard(namespace, name string) {
	if i, exists := c.HasDashboard(namespace, name); exists {
		c.Lock()
		defer c.Unlock()
		list := c.Dashboards[namespace]
		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
		c.Dashboards[namespace] = list
	}
}

func (c *ControllerConfig) GetDashboards(namespace string) []v1alpha1.GrafanaDashboardRef {
	c.Lock()
	defer c.Unlock()
	if dashboards, ok := c.Dashboards[namespace]; ok {
		return dashboards
	}
	return []v1alpha1.GrafanaDashboardRef{}
}

func (c *ControllerConfig) AddConfigItem(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if key != "" && value != nil && value != "" {
		c.Values[key] = value
	}
}

func (c *ControllerConfig) RemoveConfigItem(key string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.Values[key]; ok {
		delete(c.Values, key)
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
	c.Lock()
	defer c.Unlock()
	_, ok := c.Values[key]
	return ok
}

func (c *ControllerConfig) HasDashboard(namespace, name string) (int, bool) {
	if dashboards, ok := c.Dashboards[namespace]; ok {
		for i, dashboard := range dashboards {
			if dashboard.Name == name {
				return i, true
			}
		}
	}
	return -1, false
}
