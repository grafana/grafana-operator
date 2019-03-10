package grafana

import (
	"fmt"
	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"net/http"
	"strings"
)

const (
	PluginsEnvVar = "GRAFANA_PLUGINS"
	PluginsUrl    = "https://grafana.com/api/plugins/%s/versions/%s"
)

type PluginsHelperImpl struct {
	BaseUrl string
}

func newPluginsHelper() *PluginsHelperImpl {
	helper := new(PluginsHelperImpl)
	helper.BaseUrl = PluginsUrl
	return helper
}

// Query the Grafana plugin database for the given plugin and version
// A 200 OK response indicates that the plugin exists and can be downloaded
func (h *PluginsHelperImpl) pluginExists(plugin integreatly.GrafanaPlugin) bool {
	url := fmt.Sprintf(h.BaseUrl, plugin.Name, plugin.Version)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}

// Turns an array of plugins into a string representation of the form
// `<name>:<version>,...` that is used as the value for the GRAFANA_PLUGINS
// environment variable
func (h *PluginsHelperImpl) buildEnv(cr *integreatly.Grafana) string {
	var env []string
	for _, plugin := range cr.Status.InstalledPlugins {
		env = append(env, fmt.Sprintf("%s:%s", plugin.Name, plugin.Version))
	}
	return strings.Join(env, ",")
}

// Checks if a given plugin is already installed. We do not allow to install
// multiple versions of the same plugin
func (h *PluginsHelperImpl) pluginInstalled(plugin integreatly.GrafanaPlugin, cr *integreatly.Grafana) bool {
	for _, installedPlugin := range cr.Status.InstalledPlugins {
		if installedPlugin.Name == plugin.Name {
			return true
		}
	}
	return false
}
