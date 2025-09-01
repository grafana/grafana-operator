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
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// TODO: kubernetes 1.34+ supports isSemver function, we should migrate to it after 1.33 reaches EOL. For now, using the official pattern https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	// +kubebuilder:validation:Pattern=`^((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?|latest)$`
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

	// NOTE: We should not see that happening due to CRD validations
	// Version is not valid, so don't do anything
	requestedVersion, err := semver.Parse(version)
	if err != nil {
		return
	}

	// NOTE: We should not see that happening due to CRD validations
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

// Helps to simplify version consolidation
type PluginMap map[string]GrafanaPlugin

func (m PluginMap) Merge(plugins PluginList) {
	for _, p := range plugins {
		if p.HasInvalidVersion() {
			continue
		}

		if plugin, ok := m[p.Name]; ok {
			plugin.Update(p.Version)
			m[p.Name] = plugin
		} else {
			m[p.Name] = p
		}
	}
}

func (m PluginMap) GetPluginList() PluginList {
	plugins := slices.Collect(maps.Values(m))
	slices.SortFunc(plugins, func(a, b GrafanaPlugin) int {
		return strings.Compare(a.Name, b.Name)
	})

	return plugins
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

// Sanitize remove duplicates and enforce semver
func (l PluginList) Sanitize() PluginList {
	plugins := NewPluginMapFromList(l)

	return plugins.GetPluginList()
}
