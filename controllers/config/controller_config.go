package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
)

const (
	ConfigGrafanaImage                      = "grafana.image.url"
	ConfigGrafanaImageTag                   = "grafana.image.tag"
	ConfigPluginsInitContainerImage         = "grafana.plugins.init.container.image.url"
	ConfigPluginsInitContainerTag           = "grafana.plugins.init.container.image.tag"
	ConfigOperatorNamespace                 = "grafana.operator.namespace"
	ConfigDashboardLabelSelector            = "grafana.dashboard.selector"
	ConfigOpenshift                         = "mode.openshift"
	ConfigJsonnetBasePath                   = "grafonnet.location"
	GrafanaDataPath                         = "/var/lib/grafana"
	GrafanaLogsPath                         = "/var/log/grafana"
	GrafanaPluginsPath                      = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath                 = "/etc/grafana/provisioning/"
	GrafanaProvisioningPluginsPath          = "/etc/grafana/provisioning/plugins"
	GrafanaProvisioningDashboardsPath       = "/etc/grafana/provisioning/dashboards"
	GrafanaProvisioningNotifiersPath        = "/etc/grafana/provisioning/notifiers"
	PluginsInitContainerImage               = "quay.io/grafana-operator/grafana_plugins_init"
	PluginsInitContainerTag                 = "0.0.5"
	PluginsUrl                              = "https://grafana.com/api/plugins/%s/versions/%s"
	RequeueDelay                            = time.Second * 10
	SecretsMountDir                         = "/etc/grafana-secrets/" // #nosec G101
	ConfigMapsMountDir                      = "/etc/grafana-configmaps/"
	ConfigRouteWatch                        = "watch.routes"
	ConfigGrafanaDashboardsSynced           = "grafana.dashboards.synced"
	ConfigGrafanaNotificationChannelsSynced = "grafana.notificationchannels.synced"
	JsonnetBasePath                         = "/opt/jsonnet"
)

type ControllerConfig struct {
	*sync.Mutex
	Values     map[string]interface{}
	Plugins    map[string]v1alpha1.PluginList
	Dashboards []*v1alpha1.GrafanaDashboardRef
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Mutex:      &sync.Mutex{},
			Values:     map[string]interface{}{},
			Plugins:    map[string]v1alpha1.PluginList{},
			Dashboards: []*v1alpha1.GrafanaDashboardRef{},
		}
	})
	return instance
}

func (c *ControllerConfig) GetDashboardId(namespace, name string) string {
	return fmt.Sprintf("%v/%v", namespace, name)
}

func (c *ControllerConfig) GetAllPlugins() v1alpha1.PluginList {
	c.Lock()
	defer c.Unlock()

	var plugins v1alpha1.PluginList
	for _, v := range GetControllerConfig().Plugins {
		plugins = append(plugins, v...)
	}
	return plugins
}

func (c *ControllerConfig) GetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) v1alpha1.PluginList {
	c.Lock()
	defer c.Unlock()
	return c.Plugins[c.GetDashboardId(dashboard.Namespace, dashboard.Name)]
}

func (c *ControllerConfig) SetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	id := c.GetDashboardId(dashboard.Namespace, dashboard.Name)
	c.Lock()
	defer c.Unlock()
	c.Plugins[id] = dashboard.Spec.Plugins
}

func (c *ControllerConfig) RemovePluginsFor(namespace, name string) {
	id := c.GetDashboardId(namespace, name)
	delete(c.Plugins, id)
}

func (c *ControllerConfig) AddDashboard(dashboard *v1alpha1.GrafanaDashboard, grafana *v1alpha1.Grafana, folderId *int64, folderName string) []*v1alpha1.GrafanaDashboardRef {
	var dashboardRefs []*v1alpha1.GrafanaDashboardRef
	ns := dashboard.Namespace

	if grafana.Status.InstalledDashboards != nil {
		copy(dashboardRefs, grafana.Status.InstalledDashboards)
	}

	if _, exists := c.HasDashboard(grafana, dashboard.UID()); !exists {
		dashboardRefs = append(dashboardRefs, &v1alpha1.GrafanaDashboardRef{
			Name:       dashboard.Name,
			Namespace:  ns,
			UID:        dashboard.UID(),
			Hash:       dashboard.Hash(),
			FolderId:   folderId,
			FolderName: dashboard.Spec.CustomFolderName,
		})
	} else {
		dashboardRefs = append(dashboardRefs, &v1alpha1.GrafanaDashboardRef{
			Name:       dashboard.Name,
			Namespace:  ns,
			UID:        dashboard.UID(),
			Hash:       dashboard.Hash(),
			FolderId:   folderId,
			FolderName: folderName,
		})
	}

	return dashboardRefs
}

func (c *ControllerConfig) HasDashboard(grafana *v1alpha1.Grafana, uid string) (int, bool) {
	for i, d := range grafana.Status.InstalledDashboards {
		if d.UID == uid {
			return i, true
		}
	}

	return -1, false
}

func (c *ControllerConfig) InvalidateDashboards() {
	c.Lock()
	defer c.Unlock()
	for _, v := range c.Dashboards {
		v.Hash = ""
	}
}

func (c *ControllerConfig) SetDashboards(dashboards []*v1alpha1.GrafanaDashboardRef) {
	c.Lock()
	defer c.Unlock()
	c.Dashboards = dashboards
}

func (c *ControllerConfig) RemoveDashboard(grafana *v1alpha1.Grafana, uid string) []*v1alpha1.GrafanaDashboardRef {
	var dashboardRefs []*v1alpha1.GrafanaDashboardRef

	if grafana.Status.InstalledDashboards != nil {
		copy(dashboardRefs, grafana.Status.InstalledDashboards)
	}

	if i, exists := c.HasDashboard(grafana, uid); exists {
		dashboardRefs = append(dashboardRefs[:i], dashboardRefs[i+1:]...)
	}

	return dashboardRefs
}

func (c *ControllerConfig) GetDashboards(namespace string) []*v1alpha1.GrafanaDashboardRef {
	c.Lock()
	defer c.Unlock()
	// Checking for dashboards at the cluster level? across namespaces?
	if namespace == "" {
		var dashboards []*v1alpha1.GrafanaDashboardRef
		dashboards = append(dashboards, c.Dashboards...)

		return dashboards
	}

	if c.Dashboards != nil {
		return c.Dashboards
	}
	return []*v1alpha1.GrafanaDashboardRef{}
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
	delete(c.Values, key)
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

func (c *ControllerConfig) Cleanup(plugins bool) {
	c.Lock()
	defer c.Unlock()
	c.Dashboards = []*v1alpha1.GrafanaDashboardRef{}

	if plugins {
		c.Plugins = map[string]v1alpha1.PluginList{}
	}
}
