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
// Update updates the plugin to the requested version if it's valid and newer
func (p *GrafanaPlugin) Update(version string) {
	if p.Version == version {
		return
	}

	if version == "" {
		return
	}

	if p.Version == PluginVersionLatest {
		return
	}

	if version == PluginVersionLatest {
		p.Version = version

		return
	}

	// Version is not valid, so don't do anything
	requestedVersion, err := semver.Parse(version)
	if err != nil {
		return
	}

	// Helps to recover in case we have previously stored invalid version
	listedVersion, err := semver.Parse(p.Version)
	if err != nil {
		p.Version = version
		return
	}

	if listedVersion.Compare(requestedVersion) == -1 {
		p.Version = version

		return
	}
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
	plugins := NewPluginMapFromList(l)

	return plugins.GetPluginList()
}
