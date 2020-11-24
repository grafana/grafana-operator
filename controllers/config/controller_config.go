package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	ConfigGrafanaImage              = "grafana.image.url"
	ConfigGrafanaImageTag           = "grafana.image.tag"
	ConfigPluginsInitContainerImage = "grafana.plugins.init.container.image.url"
	ConfigPluginsInitContainerTag   = "grafana.plugins.init.container.image.tag"
	ConfigOperatorNamespace         = "grafana.operator.namespace"
	ConfigDashboardLabelSelector    = "grafana.dashboard.selector"
	ConfigOpenshift                 = "mode.openshift"
	ConfigJsonnetBasePath           = "grafonnet.location"
	GrafanaDataPath                 = "/var/lib/grafana"
	GrafanaLogsPath                 = "/var/log/grafana"
	GrafanaPluginsPath              = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath         = "/etc/grafana/provisioning/"
	PluginsInitContainerImage       = "quay.io/integreatly/grafana_plugins_init"
	PluginsInitContainerTag         = "0.0.3"
	PluginsUrl                      = "https://grafana.com/api/plugins/%s/versions/%s"
	RequeueDelay                    = time.Second * 10
	SecretsMountDir                 = "/etc/grafana-secrets/" // #nosec G101
	ConfigMapsMountDir              = "/etc/grafana-configmaps/"
	ConfigRouteWatch                = "watch.routes"
	ConfigGrafanaDashboardsSynced   = "grafana.dashboards.synced"
	JsonnetBasePath                 = "/opt/jsonnet"
)

type ControllerConfig struct {
	*sync.Mutex
	Values     map[string]interface{}
	Plugins    map[string]v1alpha1.PluginList
	Dashboards []*v1alpha1.GrafanaDashboardRef
}

var instance *ControllerConfig
var once sync.Once

var log = logf.Log.WithName("controller_config")

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

func (c *ControllerConfig) DashboardRefList() []*v1alpha1.GrafanaDashboardRef {
	var l []*v1alpha1.GrafanaDashboardRef
	for k, _ := range c.Dashboards {
		l = append(l, c.Dashboards[k]...)
	}
	return l
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

func (c *ControllerConfig) AddDashboard(dashboard *v1alpha1.GrafanaDashboard, folderId *int64, folderName string) {

	ns := dashboard.Namespace
	if i, exists := c.HasDashboard(dashboard.UID()); !exists {
		c.Lock()
		defer c.Unlock()
		c.Dashboards = append(c.Dashboards, &v1alpha1.GrafanaDashboardRef{
			Name:       dashboard.Name,
			Namespace:  ns,
			UID:        dashboard.UID(),
			Hash:       dashboard.Hash(),
			FolderId:   folderId,
			FolderName: dashboard.Spec.CustomFolderName,
		})
	} else {
		c.Lock()
		defer c.Unlock()
		c.Dashboards[i] = &v1alpha1.GrafanaDashboardRef{
			Name:       dashboard.Name,
			Namespace:  ns,
			UID:        dashboard.UID(),
			Hash:       dashboard.Hash(),
			FolderId:   folderId,
			FolderName: folderName,
		}
	}
}

func (c *ControllerConfig) HasDashboard(str string) (int, bool) {
	for i, v := range c.Dashboards {
		if v.UID == str {
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

func (c *ControllerConfig) ResetDashboards() {
	c.Dashboards = make(map[string][]*v1alpha1.GrafanaDashboardRef)
}

func (c *ControllerConfig) SetDashboards(dashboards []*v1alpha1.GrafanaDashboardRef) {
	c.Lock()
	defer c.Unlock()
	c.Dashboards = dashboards
}

func (c *ControllerConfig) RemoveDashboard(hash string) {
	if i, exists := c.HasDashboard(hash); exists {
		c.Lock()
		defer c.Unlock()
		list := c.Dashboards
		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
		c.Dashboards = list
	}
}

func (c *ControllerConfig) GetDashboards(namespace string) []*v1alpha1.GrafanaDashboardRef {
	c.Lock()
	defer c.Unlock()
	// Cluster level?
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

func (c *ControllerConfig) Cleanup(plugins bool) {
	c.Lock()
	defer c.Unlock()
	c.Dashboards = []*v1alpha1.GrafanaDashboardRef{}

	if plugins {
		c.Plugins = map[string]v1alpha1.PluginList{}
	}
}

func (c *ControllerConfig) GetLokis(namespace string) []*v1alpha1.LokiRef {
	c.Lock()
	defer c.Unlock()
	// Cluster level?
	if namespace == "" {
		var lokis []*v1alpha1.LokiRef
		for _, v := range c.Lokis {
			lokis = append(lokis, v...)
		}
		return lokis
	}

	if lokis, ok := c.Lokis[namespace]; ok {
		return lokis
	}
	return []*v1alpha1.LokiRef{}
}

func (c *ControllerConfig) HasLoki(namespace, name string) (int, bool) {
	if lokis, ok := c.Lokis[namespace]; ok {
		for i, loki := range lokis {
			if loki.Name == name {
				return i, true
			}
		}
	}
	return -1, false
}

func (c *ControllerConfig) AddLoki(loki *v1alpha1.Loki) {

	ns := loki.Namespace
	if i, exists := c.HasLoki(ns, loki.Name); !exists {
		c.Lock()
		defer c.Unlock()
		c.Lokis[ns] = append(c.Lokis[ns], &v1alpha1.LokiRef{
			Name:      loki.Name,
			Namespace: loki.Namespace,
			UID:       string(loki.UID),
			LokiURL:   string(loki.Spec.External.Url),
		})
	} else {
		c.Lock()
		defer c.Unlock()
		c.Lokis[ns][i].Name = loki.Name
		c.Lokis[ns][i].Namespace = ns
		c.Lokis[ns][i].UID = string(loki.UID)
		c.Lokis[ns][i].LokiURL = string(loki.Spec.External.Url)

	}
}

//
//func (c *ControllerConfig) InvalidateLokis() {
//	c.Lock()
//	defer c.Unlock()
//	for _, v := range c.Lokis {
//		for _, d := range v {
//			d.Hash = ""
//		}
//	}
//}

func (c *ControllerConfig) ResetLokis() {
	c.Lokis = make(map[string][]*v1alpha1.LokiRef)
}

func (c *ControllerConfig) SetLokis(lokis []*v1alpha1.LokiRef) {
	c.Lock()
	defer c.Unlock()

	for _, l := range lokis {
		c.Lokis[l.Namespace] = append(c.Lokis[l.Namespace], l)
	}
}

func (c *ControllerConfig) RemoveLoki(namespace, name string) {
	if i, exists := c.HasLoki(namespace, name); exists {
		c.Lock()
		defer c.Unlock()
		list := c.Dashboards[namespace]
		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
		c.Dashboards[namespace] = list
	}
}
