package config

import (
	"crypto/sha256"
	"fmt"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"io"
	"sort"
	"strings"
)

type GrafanaIni struct {
	cfg *v1beta1.GrafanaConfig
}

func NewGrafanaIni(cfg *v1beta1.GrafanaConfig) *GrafanaIni {
	return &GrafanaIni{
		cfg: cfg,
	}
}

func (i *GrafanaIni) Write() (string, string) {
	if i.cfg.Paths == nil {
		i.cfg.Paths = make(map[string]string)
	}

	// default paths that can't be overridden
	i.cfg.Paths["data"] = GrafanaDataPath
	i.cfg.Paths["logs"] = GrafanaLogsPath
	i.cfg.Paths["plugins"] = GrafanaPluginsPath
	i.cfg.Paths["provisioning"] = GrafanaProvisioningPath

	sb := strings.Builder{}
	sb.WriteString(i.writeSection("log", i.cfg.Log))
	sb.WriteString(i.writeSection("paths", i.cfg.Paths))
	sb.WriteString(i.writeSection("server", i.cfg.Server))

	hash := sha256.New()
	io.WriteString(hash, sb.String()) // nolint

	return sb.String(), fmt.Sprintf("%x", hash.Sum(nil))
}

func (i *GrafanaIni) writeSection(name string, settings map[string]string) string {
	if settings == nil || len(settings) == 0 {
		return ""
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("[%s]", name))
	sb.WriteByte('\n')

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
	return sb.String()
}
