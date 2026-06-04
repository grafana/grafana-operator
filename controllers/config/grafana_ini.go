package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
)

const globalSection = "global"

// NOTE: even though there is no need to return map, it's added here to make sure
// we can test the case where the passed value is nil
func SetDefaults(cfg map[string]map[string]string, version string) map[string]map[string]string {
	if cfg == nil {
		cfg = make(map[string]map[string]string)
	}

	if cfg["paths"] == nil {
		cfg["paths"] = make(map[string]string)
	}

	// default paths that can't be overridden
	cfg["paths"]["data"] = GrafanaDataPath
	cfg["paths"]["logs"] = GrafanaLogsPath
	cfg["paths"]["plugins"] = GrafanaPluginsPath
	cfg["paths"]["provisioning"] = GrafanaProvisioningPath

	if cfg["dashboards"] == nil {
		cfg["dashboards"] = make(map[string]string)
	}

	if cfg["dashboards"]["versions_to_keep"] == "" {
		cfg["dashboards"]["versions_to_keep"] = GrafanaDashboardVersionsToKeep
	}

	if cfg["unified_alerting"] == nil {
		cfg["unified_alerting"] = make(map[string]string)
	}

	if cfg["unified_alerting"]["rule_version_record_limit"] == "" {
		cfg["unified_alerting"]["rule_version_record_limit"] = GrafanaRuleVersionRecordLimit
	}

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		// if we can't infer the version, return early
		return cfg
	}

	if parsedVersion.GE(semver.MustParse("13.0.0")) {
		// OOM due to changes in gzip implementation: https://github.com/grafana/grafana/issues/123017
		if cfg["server"] == nil {
			cfg["server"] = make(map[string]string)
		}

		if cfg["server"]["enable_gzip"] == "" {
			cfg["server"]["enable_gzip"] = "false"
		}
	}

	return cfg
}

func WriteIni(cfg map[string]map[string]string) string {
	sections := make([]string, 0, len(cfg))
	hasGlobal := false

	for key := range cfg {
		if key == globalSection {
			hasGlobal = true
			continue
		}

		sections = append(sections, key)
	}

	sort.Strings(sections)

	if hasGlobal {
		sections = append([]string{globalSection}, sections...)
	}

	sb := &strings.Builder{}

	for _, section := range sections {
		if len(cfg[section]) == 0 {
			continue
		}

		writeSection(section, cfg[section], sb)
	}

	return sb.String()
}

func writeSection(name string, settings map[string]string, sb *strings.Builder) {
	if name != globalSection {
		fmt.Fprintf(sb, "[%s]", name)
		sb.WriteByte('\n')
	}

	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(sb, "%s = %s", key, settings[key])
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')
}

func GetHash(cfg string) string {
	hash := sha256.New()
	io.WriteString(hash, cfg) //nolint

	return fmt.Sprintf("%x", hash.Sum(nil))
}
