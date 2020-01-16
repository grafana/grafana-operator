package config

import (
	"crypto/md5"
	"fmt"
	"io"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"gopkg.in/yaml.v2"
)

type GrafanaProxyConfig struct {
	cfg *v1alpha1.GrafanaProxyConfig
}

func NewGrafanaProxyConfig(cfg *v1alpha1.GrafanaProxyConfig) *GrafanaProxyConfig {
	return &GrafanaProxyConfig{
		cfg: cfg,
	}
}

func (i *GrafanaProxyConfig) Write() (cfg string, h string) {
	outb, err := yaml.Marshal(i.cfg)
	if err != nil {
		return
	}
	cfg = string(outb)
	hash := md5.New()
	io.WriteString(hash, cfg)

	return cfg, fmt.Sprintf("%x", hash.Sum(nil))
}
