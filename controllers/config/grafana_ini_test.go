package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var globalSectionRaw = map[string]string{
	"app_mode":      "production",
	"instance_name": "my_instance",
}

var globalSectionRendered = `app_mode = production
instance_name = my_instance

`

var standardSectionRaw = map[string]string{
	"type": "sqlite3",
	"host": "127.0.0.1:3306",
}

var standardSectionRendered = `[database]
host = 127.0.0.1:3306
type = sqlite3

`

func getDefaultConfig(t *testing.T) map[string]map[string]string {
	t.Helper()

	defaultCfg := map[string]map[string]string{
		"paths": {
			"data":         GrafanaDataPath,
			"logs":         GrafanaLogsPath,
			"plugins":      GrafanaPluginsPath,
			"provisioning": GrafanaProvisioningPath,
		},
		"dashboards": {
			"versions_to_keep": GrafanaDashboardVersionsToKeep,
		},
		"unified_alerting": {
			"rule_version_record_limit": GrafanaRuleVersionRecordLimit,
		},
	}

	return defaultCfg
}

func TestSetDefaults(t *testing.T) {
	t.Run("Nil config is properly handled", func(t *testing.T) {
		cfg := setDefaults(nil)

		got := cfg
		want := getDefaultConfig(t)

		assert.Equal(t, want, got)
	})

	t.Run("All defaults are set", func(t *testing.T) {
		cfg := map[string]map[string]string{}
		cfg = setDefaults(cfg)

		got := cfg
		want := getDefaultConfig(t)

		assert.Equal(t, want, got)
	})

	t.Run("Paths are overwritten", func(t *testing.T) {
		cfg := map[string]map[string]string{
			"paths": {
				"data":         "a",
				"logs":         "b",
				"plugins":      "c",
				"provisioning": "d",
			},
		}
		cfg = setDefaults(cfg)

		got := cfg
		want := getDefaultConfig(t)

		assert.Equal(t, want, got)
	})

	t.Run("Limits overrides are respected", func(t *testing.T) {
		dashboardsOverrides := map[string]string{
			"versions_to_keep": "99",
		}
		unifiedAlertingOverrides := map[string]string{
			"rule_version_record_limit": "999",
		}

		cfg := map[string]map[string]string{
			"dashboards":       dashboardsOverrides,
			"unified_alerting": unifiedAlertingOverrides,
		}
		cfg = setDefaults(cfg)

		got := cfg
		want := getDefaultConfig(t)
		want["dashboards"] = dashboardsOverrides
		want["unified_alerting"] = unifiedAlertingOverrides

		assert.Equal(t, want, got)
	})

	t.Run("Custom sections are preserved", func(t *testing.T) {
		customSectionOverrides := map[string]string{
			"custom_setting": "custom_value",
		}

		cfg := map[string]map[string]string{
			"custom_section": customSectionOverrides,
		}
		cfg = setDefaults(cfg)

		got := cfg
		want := getDefaultConfig(t)
		want["custom_section"] = customSectionOverrides

		assert.Equal(t, want, got)
	})
}

func TestWriteIni(t *testing.T) {
	t.Run("Global section comes first", func(t *testing.T) {
		cfg := getDefaultConfig(t)
		cfg["global"] = globalSectionRaw
		cfg["abc"] = map[string]string{
			"setting": "value",
		}

		got := WriteIni(cfg)

		assert.True(t, strings.HasPrefix(got, globalSectionRendered))
	})

	t.Run("Standard section is present", func(t *testing.T) {
		cfg := getDefaultConfig(t)
		cfg["database"] = standardSectionRaw

		got := WriteIni(cfg)

		assert.Contains(t, got, standardSectionRendered)
	})

	t.Run("Empty and nil sections are skipped", func(t *testing.T) {
		cfg := getDefaultConfig(t)
		cfg["empty_section"] = map[string]string{}
		cfg["nil_section"] = nil

		got := WriteIni(cfg)

		assert.NotContains(t, got, "[empty_section]")
		assert.NotContains(t, got, "[nil_section]")
	})
}

func TestWriteSection(t *testing.T) {
	tests := []struct {
		name        string
		sectionName string
		settings    map[string]string
		want        string
	}{
		{
			name:        "Global section should NOT have title",
			sectionName: "global",
			settings:    globalSectionRaw,
			want:        globalSectionRendered,
		},
		{
			name:        "Standard section should have title",
			sectionName: "database",
			settings:    standardSectionRaw,
			want:        standardSectionRendered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			writeSection(tt.sectionName, tt.settings, sb)
			got := sb.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetHash(t *testing.T) {
	t.Run("Hash is properly calculated", func(t *testing.T) {
		cfg := `[abc]`

		got := GetHash(cfg)
		want := "c93852370e1b1afda3feeb94ffda3df904cea135ea2107abc216358853df375a"

		assert.Equal(t, want, got)
	})
}
