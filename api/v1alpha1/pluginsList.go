package v1alpha1

import (
	"github.com/blang/semver"
)

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

// Get the plugin from the list regardless of the version
func (l PluginList) GetInstalledVersionOf(plugin *GrafanaPlugin) *GrafanaPlugin {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			return &listedPlugin
		}
	}
	return nil
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

// Returns true if the list contains the same plugin but in a newer version
func (l PluginList) HasNewerVersionOf(plugin *GrafanaPlugin) (bool, error) {
	for _, listedPlugin := range l {
		if listedPlugin.Name != plugin.Name {
			continue
		}

		listedVersion, err := semver.Make(listedPlugin.Version)
		if err != nil {
			return false, err
		}

		requestedVersion, err := semver.Make(plugin.Version)
		if err != nil {
			return false, err
		}

		if listedVersion.Compare(requestedVersion) == 1 {
			return true, nil
		}
	}
	return false, nil
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
