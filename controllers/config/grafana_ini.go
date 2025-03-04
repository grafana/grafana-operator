package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
	"strings"
)

func WriteIni(cfg map[string]map[string]string) (string, string) {
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
		cfg["dashboards"]["versions_to_keep"] = "20"
	}

	if cfg["unified_alerting"] == nil {
		cfg["unified_alerting"] = make(map[string]string)
	}

	if cfg["unified_alerting"]["rule_version_record_limit"] == "" {
		cfg["unified_alerting"]["rule_version_record_limit"] = "5"
	}

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

	hash := sha256.New()
	io.WriteString(hash, sb.String()) //nolint

	return sb.String(), fmt.Sprintf("%x", hash.Sum(nil))
}

func writeSection(name string, settings map[string]string, sb *strings.Builder) {
	if name != "global" {
		sb.WriteString(fmt.Sprintf("[%s]", name))
		sb.WriteByte('\n')
	}

	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%s = %s", key, settings[key]))
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
}
