package v1alpha1

type PluginList []GrafanaPlugin

// Returns true if the list contains the same plugin in the exact or a different version
func (l PluginList) HasSomeVersionOf(plugin *GrafanaPlugin) bool {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			return true
		}
	}
	return false
}

// Returns true if the list contains the same plugin in the same version
func (l PluginList) HasExactVersionOf(plugin *GrafanaPlugin) bool {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name && listedPlugin.Version == plugin.Version {
			return true
		}
	}
	return false
}

// Returns the number of different versions of a given plugin in the list
func (l PluginList) VersionsOf(plugin *GrafanaPlugin) int {
	i := 0
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			i = i + 1
		}
	}
	return i
}

// Returns the number of different versions of a given plugin in the list
func (l PluginList) SetOrigin(dashboard *GrafanaDashboard) {
	for i := range l {
		l[i].Origin = dashboard
	}
}
