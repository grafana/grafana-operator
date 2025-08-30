package v1beta1

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
)

type GrafanaPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (p GrafanaPlugin) HasValidVersion() bool {
	if p.Version == PluginVersionLatest {
		return true
	}

	_, err := semver.Parse(p.Version)

	return err == nil
}

func (p GrafanaPlugin) HasInvalidVersion() bool {
	return !p.HasValidVersion()
}

func (p GrafanaPlugin) String() string {
	if p.Version == PluginVersionLatest {
		return p.Name
	}

	return fmt.Sprintf("%s %s", p.Name, p.Version)
}

type PluginList []GrafanaPlugin

type PluginMap map[string]PluginList

func (l PluginList) String() string {
	plugins := make(sort.StringSlice, 0, len(l))

	for _, plugin := range l {
		plugins = append(plugins, plugin.String())
	}

	sort.Sort(plugins)

	return strings.Join(plugins, ",")
}

// Update update plugin version
func (l PluginList) Update(plugin *GrafanaPlugin) {
	for i, installedPlugin := range l {
		if installedPlugin.Name == plugin.Name {
			l[i].Version = plugin.Version
			break
		}
	}
}

// Sanitize remove duplicates and enforce semver
func (l PluginList) Sanitize() PluginList {
	var sanitized PluginList

	for _, plugin := range l {
		if plugin.HasInvalidVersion() {
			continue
		}

		if sanitized.HasSomeVersionOf(&plugin) {
			hasNewer, err := sanitized.HasNewerVersionOf(&plugin)
			if err != nil {
				continue
			}

			if hasNewer {
				continue
			}

			sanitized.Update(&plugin)

			continue
		}

		sanitized = append(sanitized, plugin)
	}

	return sanitized
}

// HasSomeVersionOf returns true if the list contains the same plugin in the exact or a different version
func (l PluginList) HasSomeVersionOf(plugin *GrafanaPlugin) bool {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			return true
		}
	}

	return false
}

// HasExactVersionOf returns true if the list contains the same plugin in the same version
func (l PluginList) HasExactVersionOf(plugin *GrafanaPlugin) bool {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name && listedPlugin.Version == plugin.Version {
			return true
		}
	}

	return false
}

// HasNewerVersionOf returns true if the list contains the same plugin but in a newer version
func (l PluginList) HasNewerVersionOf(plugin *GrafanaPlugin) (bool, error) {
	for _, listedPlugin := range l {
		if listedPlugin.Name != plugin.Name {
			continue
		}

		if listedPlugin.Version == PluginVersionLatest && plugin.Version != PluginVersionLatest {
			return true, nil
		}

		if plugin.Version == PluginVersionLatest {
			return false, nil
		}

		listedVersion, err := semver.Parse(listedPlugin.Version)
		if err != nil {
			return false, err
		}

		requestedVersion, err := semver.Parse(plugin.Version)
		if err != nil {
			return false, err
		}

		if listedVersion.Compare(requestedVersion) == 1 {
			return true, nil
		}
	}

	return false, nil
}
