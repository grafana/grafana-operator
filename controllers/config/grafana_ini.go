package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
	"strings"
)

// NOTE: even though there is no need to return map, it's added here to make sure
// we can test the case where the passed value is nil
func setDefaults(cfg map[string]map[string]string) map[string]map[string]string {
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

	return cfg
}

func WriteIni(cfg map[string]map[string]string) string {
	cfg = setDefaults(cfg)

	sections := make([]string, 0, len(cfg))
	hasGlobal := false

	for key := range cfg {
		if key == "global" {
			hasGlobal = true
			continue
		}

		sections = append(sections, key)
	}

	sort.Strings(sections)

	if hasGlobal {
		sections = append([]string{"global"}, sections...)
	}

	sb := &strings.Builder{}

	for _, section := range sections {
		if cfg[section] == nil || len(cfg[section]) == 0 {
			continue
		}

		writeSection(section, cfg[section], sb)
	}

	return sb.String()
}

func writeSection(name string, settings map[string]string, sb *strings.Builder) {
	if name != "global" {
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
