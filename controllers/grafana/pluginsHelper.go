package grafana

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
)

type PluginsHelperImpl struct {
	BaseUrl    string
	HttpClient *http.Client
}

func NewPluginsHelper() *PluginsHelperImpl {
	/* #nosec G402 */
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}

	helper := new(PluginsHelperImpl)
	helper.BaseUrl = config.PluginsUrl
	helper.HttpClient = &http.Client{Transport: insecureTransport}

	return helper
}

// Query the Grafana plugin database for the given plugin and version
// A 200 OK response indicates that the plugin exists and can be downloaded
func (h *PluginsHelperImpl) PluginExists(plugin grafanav1alpha1.GrafanaPlugin) bool {
	url := fmt.Sprintf(h.BaseUrl, plugin.Name, plugin.Version)
	resp, err := h.HttpClient.Get(url)
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Turns an array of plugins into a string representation of the form
// `<name>:<version>,...` that is used as the value for the GRAFANA_PLUGINS
// environment variable
func (h *PluginsHelperImpl) BuildEnv(cr *grafanav1alpha1.Grafana) string {
	env := make([]string, 0, len(cr.Status.InstalledPlugins))

	for _, plugin := range cr.Status.InstalledPlugins {
		env = append(env, fmt.Sprintf("%s:%s", plugin.Name, plugin.Version))
	}
	return strings.Join(env, ",")
}

// Append a status message to the origin dashboard of a plugin
func (h *PluginsHelperImpl) pickLatestVersions(requested grafanav1alpha1.PluginList) (grafanav1alpha1.PluginList, error) {
	var latestVersions grafanav1alpha1.PluginList
	for i := range requested {
		result, err := requested.HasNewerVersionOf(&requested[i])

		// Errors might happen if plugins don't use semver
		// In that case fall back to whichever comes first
		if err != nil {
			return requested, err
		}

		// Skip this version if there is a more recent one
		if result {
			continue
		}
		latestVersions = append(latestVersions, requested[i])
	}
	return latestVersions, nil
}

// Creates the list of plugins that can be added or updated
// Does not directly deal with removing plugins: if a plugin is not in the list and the env var is updated, it will
// automatically be removed
func (h *PluginsHelperImpl) FilterPlugins(cr *grafanav1alpha1.Grafana, requested grafanav1alpha1.PluginList) (grafanav1alpha1.PluginList, bool) {
	filteredPlugins := grafanav1alpha1.PluginList{}
	pluginsUpdated := false

	// Try to pick the latest versions of all plugins
	requested, err := h.pickLatestVersions(requested)
	if err != nil {
		log.Error(err, "unable to pick latest plugin versions")
	}

	// Remove all plugins
	if len(requested) == 0 && len(cr.Status.InstalledPlugins) > 0 {
		return filteredPlugins, true
	}

	for i := range requested {
		// Don't allow to install multiple versions of the same plugin
		if filteredPlugins.HasSomeVersionOf(&requested[i]) {
			installedVersion := filteredPlugins.GetInstalledVersionOf(&requested[i])
			log.V(1).Info(fmt.Sprintf("not installing version %s of %s because %s is already installed", requested[i].Version, requested[i].Name, installedVersion.Version))
			continue
		}

		if cr.Status.FailedPlugins.HasExactVersionOf(&requested[i]) {
			// Don't attempt to install plugins that failed to install previously
			continue
		}

		// Already installed: append it to the list to keep it
		if cr.Status.InstalledPlugins.HasExactVersionOf(&requested[i]) {
			filteredPlugins = append(filteredPlugins, requested[i])
			continue
		}

		// New plugin
		if !cr.Status.InstalledPlugins.HasSomeVersionOf(&requested[i]) {
			filteredPlugins = append(filteredPlugins, requested[i])
			log.V(1).Info(fmt.Sprintf("installing plugin %s@%s", requested[i].Name, requested[i].Version))
			pluginsUpdated = true
			continue
		}

		// Plugin update: allow to update a plugin if only one dashboard requests it
		// The condition is: some version of the plugin is aleady installed, but it's not the exact same version
		// and there is only one dashboard that requires this plugin
		// If multiple dashboards request different versions of the same plugin, then we can't upgrade because
		// there is no way to decide which version is the correct one
		if cr.Status.InstalledPlugins.HasSomeVersionOf(&requested[i]) &&
			!cr.Status.InstalledPlugins.HasExactVersionOf(&requested[i]) &&
			requested.VersionsOf(&requested[i]) == 1 {
			installedVersion := cr.Status.InstalledPlugins.GetInstalledVersionOf(&requested[i])
			filteredPlugins = append(filteredPlugins, requested[i])
			log.V(1).Info(fmt.Sprintf("changing version of plugin %s form %s to %s", requested[i].Name, installedVersion.Version, requested[i].Version))
			pluginsUpdated = true
			continue
		}
	}

	// Check for removed plugins
	for i := range cr.Status.InstalledPlugins {
		if !requested.HasSomeVersionOf(&cr.Status.InstalledPlugins[i]) {
			pluginsUpdated = true
		}
	}

	return filteredPlugins, pluginsUpdated
}
