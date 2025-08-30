package v1beta1

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPluginListUpdate(t *testing.T) {
	tests := []struct {
		name    string
		plugins PluginList
		plugin  GrafanaPlugin
		want    PluginList
	}{
		{
			name: "version is updated",
			plugins: PluginList{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
			want: PluginList{
				{
					Name:    "a",
					Version: "2.0.0",
				},
			},
		},
		{
			name: "no match",
			plugins: PluginList{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "b",
				Version: "2.0.0",
			},
			want: PluginList{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugins.Update(&tt.plugin)

			got := tt.plugins

			assert.Equal(t, tt.want, got)
		})
	}
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

func TestPluginListSomeVersionOf(t *testing.T) {
	tests := []struct {
		name    string
		plugins PluginList
		plugin  GrafanaPlugin
		want    bool
	}{
		{
			name: "has same version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: true,
		},
		{
			name: "has different version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
			want: true,
		},
		{
			name: "doesn't have any versions of the same plugin",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "b",
				Version: "1.0.0",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugins.HasSomeVersionOf(&tt.plugin)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPluginListHasExactVersionOf(t *testing.T) {
	tests := []struct {
		name    string
		plugins PluginList
		plugin  GrafanaPlugin
		want    bool
	}{
		{
			name: "has same version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: true,
		},
		{
			name: "has different version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
			want: false,
		},
		{
			name: "different plugin has same version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "b",
				Version: "2.0.0",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugins.HasExactVersionOf(&tt.plugin)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPluginListHasNewerVersionOf(t *testing.T) {
	tests := []struct {
		name    string
		plugins PluginList
		plugin  GrafanaPlugin
		want    bool
	}{
		{
			name: "has newer version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.1.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: true,
		},
		{
			name: "has newer version (latest)",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "latest",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.0.0",
			},
			want: true,
		},
		{
			name: "has older version",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "1.1.0",
			},
			want: false,
		},
		{
			name: "has older version (latest)",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "latest",
			},
			want: false,
		},
		{
			name: "doesn't have any versions",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "b",
				Version: "1.0.0",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.plugins.HasNewerVersionOf(&tt.plugin)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}

	// Error cases
	tests2 := []struct {
		name    string
		plugins PluginList
		plugin  GrafanaPlugin
	}{
		{
			name: "broken version in plugin list",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "a.b.c",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "2.0.0",
			},
		},
		{
			name: "broken version of target plugin",
			plugins: []GrafanaPlugin{
				{
					Name:    "a",
					Version: "1.0.0",
				},
			},
			plugin: GrafanaPlugin{
				Name:    "a",
				Version: "a.b.c",
			},
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.plugins.HasNewerVersionOf(&tt.plugin)
			require.Error(t, err)
			assert.False(t, got)
		})
	}
}
