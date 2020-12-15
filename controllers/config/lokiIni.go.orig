package config

import (
	"crypto/sha256"
	"fmt"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"io"
	"sort"
	"strings"
)

type LokiIni struct {
	cfg *v1alpha1.LokiConfig
}

func NewLokiIni(cfg *v1alpha1.LokiConfig) *LokiIni {
	return &LokiIni{
		cfg: cfg,
	}
}

func (i *LokiIni) Write() (string, string) {
	config := map[string][]string{}

	appendStr := func(l []string, key, value string) []string {
		if value != "" {
			return append(l, fmt.Sprintf("%v = %v", key, value))
		}
		return l
	}

	//appendInt := func(l []string, key string, value *int) []string {
	//	if value != nil {
	//		return append(l, fmt.Sprintf("%v = %v", key, *value))
	//	}
	//	return l
	//}
	//
	//appendBool := func(l []string, key string, value *bool) []string {
	//	if value != nil {
	//		return append(l, fmt.Sprintf("%v = %v", key, *value))
	//	}
	//	return l
	//}

	config["paths"] = []string{
		fmt.Sprintf("data = %v", LokiDataPath),
		fmt.Sprintf("logs = %v", LokiLogsPath),
		fmt.Sprintf("provisioning = %v", GrafanaProvisioningPath),
	}

	if i.cfg.Paths != nil {
		config["paths"] = append(config["paths"],
			fmt.Sprintf("temp_data_lifetime = %v",
				i.cfg.Paths.TempDataLifetime))
	}

	if i.cfg.Server != nil {
		var items []string
		items = appendStr(items, "http_addr", i.cfg.Server.HttpAddr)
		items = appendStr(items, "http_port", i.cfg.Server.HttpPort)
		config["server"] = items
	}

	sb := strings.Builder{}

	var keys []string
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
	io.WriteString(hash, sb.String())

	return sb.String(), fmt.Sprintf("%x", hash.Sum(nil))
}
