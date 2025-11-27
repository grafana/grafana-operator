package v1beta1

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestGrafanaPluginHasValidVersion(t *testing.T) {
	tests := []struct {
		name   string
		plugin GrafanaPlugin
		want   bool
	}{
		{
			name: "latest version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "latest",
			},
			want: true,
		},
		{
			name: "valid semver version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: true,
		},
		{
			name: "semver version with v prefix", // Not supported yet
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "v1.0.0",
			},
			want: false,
		},
		{
			name: "invalid version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "a.b.c",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugin.HasValidVersion()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrafanaPluginString(t *testing.T) {
	tests := []struct {
		name   string
		plugin GrafanaPlugin
		want   string
	}{
		{
			name: "latest version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "latest",
			},
			want: "a",
		},
		{
			name: "semver",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: "a 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugin.String()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrafanaPluginUpdate(t *testing.T) {
	tests := []struct {
		name    string
		plugin  GrafanaPlugin
		version string
		want    GrafanaPlugin
	}{
		{
			name: "same version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "1.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
		{
			name: "empty target version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
		{
			name: "latest version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: PluginVersionLatest,
			want: GrafanaPlugin{
				Name:    "a",
				Version: PluginVersionLatest,
			},
		},
		{
			name: "already latest version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: PluginVersionLatest,
			},
			version: "1.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: PluginVersionLatest,
			},
		},
		{
			name: "both have latest version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: PluginVersionLatest,
			},
			version: PluginVersionLatest,
			want: GrafanaPlugin{
				Name:    "a",
				Version: PluginVersionLatest,
			},
		},
		{
			name: "older version passed",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "0.1.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
		{
			name: "newer version passed",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "2.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
		},
		// Error cases (as we have validation at CRD level, the cases below were added mostly to document function behavior)
		{
			name: "incorrect source version, but correct target version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "a.b.c",
			},
			version: "1.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
		{
			name: "incorrect target version",
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "a.b.c",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
		{
			name: "source semver version with v prefix", // Not supported yet
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "v1.0.0",
			},
			version: "2.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
		},
		{
			name: "target semver version with v prefix", // Not supported yet
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			version: "v2.0.0",
			want: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Update(tt.version)

			got := tt.plugin

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPluginListString(t *testing.T) {
	err := quick.Check(func(a string, b string, c string) bool {
		if strings.Contains(a, ",") || strings.Contains(b, ",") || strings.Contains(c, ",") {
			return true // skip plugins with ,
		}

		pl := PluginList{
			{
				Name:    a,
				Version: "7.2",
			},
			{
				Name:    b,
				Version: "2.2",
			},
			{
				Name:    c,
				Version: "6.7",
			},
		}
		out := pl.String()

		split := strings.Split(out, ",")
		if len(split) != 3 {
			return false
		}

		if split[0] > split[1] {
			return false
		}

		if split[1] > split[2] {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Errorf("plugin list was not sorted: %s", err.Error())
	}

	t.Run("Correct string", func(t *testing.T) {
		pl := PluginList{
			{
				Name:    "a",
				Version: "1.0.0",
			},
			{
				Name:    "b",
				Version: "latest",
			},
			{
				Name:    "c",
				Version: "2.0.0",
			},
		}

		got := pl.String()
		want := "a 1.0.0,b,c 2.0.0"

		assert.Equal(t, want, got)
	})
}

func TestPluginListSanitize(t *testing.T) {
	tests := []struct {
		name    string
		plugins PluginList
		want    PluginList
	}{
		{
			name: "duplicates removal",
			plugins: PluginList{
				{
					Name:    "a",
					Version: "3.0.0",
				},
				{
					Name:    "a",
					Version: "1.0.0",
				},
				{
					Name:    "b",
					Version: "2.0.0",
				},
			},
			want: PluginList{
				{
					Name:    "a",
					Version: "3.0.0",
				},
				{
					Name:    "b",
					Version: "2.0.0",
				},
			},
		},
		{
			name: "skip incorrect versions",
			plugins: PluginList{
				{
					Name:    "a",
					Version: "a.b.c",
				},
				{
					Name:    "b",
					Version: "2.0.0",
				},
				{
					Name:    "c",
					Version: "latest",
				},
			},
			want: PluginList{
				{
					Name:    "b",
					Version: "2.0.0",
				},
				{
					Name:    "c",
					Version: "latest",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugins.Sanitize()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPluginMapMerge(t *testing.T) {
	plugins := []GrafanaPlugin{
		{
			Name:    "a",
			Version: "latest",
		},
		{
			Name:    "a",
			Version: "1.0.0",
		},
		{
			Name:    "a",
			Version: "1.0.1",
		},
		{
			Name:    "b",
			Version: "2.0.1",
		},
		{
			Name:    "b",
			Version: "2.0.0",
		},
	}

	want := PluginMap{
		"a": {
			Name:    "a",
			Version: "latest",
		},
		"b": {
			Name:    "b",
			Version: "2.0.1",
		},
	}

	got := PluginMap{}
	got.Merge(plugins)

	assert.Equal(t, want, got)
}
