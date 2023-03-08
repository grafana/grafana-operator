package v1beta1

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/blang/semver"
)

type GrafanaPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PluginList []GrafanaPlugin

type PluginMap map[string]PluginList

func (l PluginList) Hash() string {
	sb := strings.Builder{}
	for _, plugin := range l {
		sb.WriteString(plugin.Name)
		sb.WriteString(plugin.Version)
	}
	hash := sha256.New()
	io.WriteString(hash, sb.String()) //nolint
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (l PluginList) String() string {
	plugins := make([]string, 0, len(l))
	for _, plugin := range l {
		plugins = append(plugins, fmt.Sprintf("%s %s", plugin.Name, plugin.Version))
	}
	return strings.Join(plugins, ",")
}

// Update update plugin version
func (l PluginList) Update(plugin *GrafanaPlugin) {
	for _, installedPlugin := range l {
		if installedPlugin.Name == plugin.Name {
			installedPlugin.Version = plugin.Version
			break
		}
	}
}

// Sanitize remove duplicates and enforce semver
func (l PluginList) Sanitize() PluginList {
	var sanitized PluginList
	for _, plugin := range l {
		plugin := plugin
		_, err := semver.Parse(plugin.Version)
		if err != nil {
			continue
		}
		if !sanitized.HasSomeVersionOf(&plugin) {
			sanitized = append(sanitized, plugin)
		}
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

// GetInstalledVersionOf gets the plugin from the list regardless of the version
func (l PluginList) GetInstalledVersionOf(plugin *GrafanaPlugin) *GrafanaPlugin {
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			return &listedPlugin
		}
	}
	return nil
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

// VersionsOf returns the number of different versions of a given plugin in the list
func (l PluginList) VersionsOf(plugin *GrafanaPlugin) int {
	i := 0
	for _, listedPlugin := range l {
		if listedPlugin.Name == plugin.Name {
			i++
		}
	}
	return i
}
