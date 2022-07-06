package config

import (
	"crypto/sha256"
	"fmt"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"io"
	"sort"
	"strconv"
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

func appendStr(list []string, key, value string) []string {
	if value != "" {
		return append(list, fmt.Sprintf("%v = %v", key, value))
	}
	return list
}

func appendInt(list []string, key string, value *int) []string {
	if value != nil {
		return append(list, fmt.Sprintf("%v = %v", key, *value))
	}
	return list
}

func appendFloat(list []string, key string, value string) []string {
	if value != "" {
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return list
		}
		return append(list, fmt.Sprintf("%v = %v", key, f))
	}
	return list
}

func appendBool(list []string, key string, value *bool) []string {
	if value != nil {
		return append(list, fmt.Sprintf("%v = %v", key, *value))
	}
	return list
}

func (i *GrafanaIni) Write() (string, string) {
	config := map[string][]string{}
	config = i.parseConfig(config)

	sb := strings.Builder{}

	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := config[key]
		sort.Strings(values)

		// Section begin
		sb.WriteString(fmt.Sprintf("[%s]", key))
		sb.WriteByte('\n')

		// Section keys
		for _, value := range values {
			sb.WriteString(value)
			sb.WriteByte('\n')
		}

		// Section end
		sb.WriteByte('\n')
	}

	hash := sha256.New()
	io.WriteString(hash, sb.String()) // nolint

	return sb.String(), fmt.Sprintf("%x", hash.Sum(nil))
}

//nolint:gocyclo,funlen,cyclop // Splitting it up will just make it more unreadable
func (i *GrafanaIni) parseConfig(config map[string][]string) map[string][]string {
	config["paths"] = []string{
		fmt.Sprintf("data = %v", GrafanaDataPath),
		fmt.Sprintf("logs = %v", GrafanaLogsPath),
		fmt.Sprintf("plugins = %v", GrafanaPluginsPath),
		fmt.Sprintf("provisioning = %v", GrafanaProvisioningPath),
	}

	return config
}
