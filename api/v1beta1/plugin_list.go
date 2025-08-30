package v1beta1

import (
	"fmt"
	"maps"
	"slices"
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

// Update updates the plugin to the requested version if its newer
func (p *GrafanaPlugin) Update(version string) error {
	if p.Version == version {
		return nil
	}

	if version == "" {
		return nil
	}

	if p.Version == PluginVersionLatest {
		return nil
	}

	if version == PluginVersionLatest {
		p.Version = version

		return nil
	}

	listedVersion, err := semver.Parse(p.Version)
	if err != nil {
		return err
	}

	requestedVersion, err := semver.Parse(version)
	if err != nil {
		return err
	}

	if listedVersion.Compare(requestedVersion) == -1 {
		p.Version = version

		return nil
	}

	return nil
}

type PluginMap map[string]GrafanaPlugin

func (m PluginMap) Merge(plugins PluginList) {
	for _, p := range plugins {
		if p.HasInvalidVersion() {
			continue
		}

		if plugin, ok := m[p.Name]; ok {
			// TODO: it can return errors, but if we add CRD validation, that's not necessary
			plugin.Update(p.Version)
			m[p.Name] = plugin
		} else {
			m[p.Name] = p
		}
	}
}

func (m PluginMap) GetPluginList() PluginList {
	return slices.Collect(maps.Values(m))
}

func NewPluginMap() PluginMap {
	pm := PluginMap{}

	return pm
}

func NewPluginMapFromList(plugins PluginList) PluginMap {
	pm := PluginMap{}

	pm.Merge(plugins)

	return pm
}

type PluginList []GrafanaPlugin

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
