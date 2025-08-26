package v1beta1

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
}

func TestPluginListSanitize(t *testing.T) {
	pl := PluginList{
		{
			Name:    "plugin-a",
			Version: "1.0.0",
		},
		{
			Name:    "plugin-b",
			Version: "2.0.0",
		},
		{
			Name:    "plugin-a",
			Version: "3.0.0",
		},
	}
	sanitized := pl.Sanitize()
	assert.Len(t, sanitized, 2)
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
