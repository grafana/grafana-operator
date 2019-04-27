package common

import (
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"sync"
	"time"
)

const (
	ConfigGrafanaImage           = "grafana.image.url"
	ConfigGrafanaImageTag        = "grafana.image.tag"
	ConfigOperatorNamespace      = "grafana.operator.namespace"
	ConfigDashboardLabelSelector = "grafana.dashboard.selector"
	ConfigGrafanaPluginsUpdated  = "grafana.plugins.updated"
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

func (c *ControllerConfig) EmptyPluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	id := c.GetDashboardId(dashboard)
	c.Plugins[id] = nil
	c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
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
