package config

import (
	"fmt"
	"io/ioutil"

	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Grafana   models.Grafana `yaml:"grafana"`
		Version   string         `yaml:"version"`
		AuthProxy AuthProxy      `yaml:"auth_proxy"`
		Region    string         `yaml:"region"`
	}
	AuthProxy struct {
		Connectors map[string]map[string]string `yaml:"connectors"`
	}
)

func GetConfig(opts Options) (cfg Config, err error) {
	if opts.ConfigFilePath == "" {
		return cfg, fmt.Errorf("no config file provided")
	}
	jsonBytes, err := ioutil.ReadFile(opts.ConfigFilePath)
	if err != nil {
		return cfg, fmt.Errorf("read config file: %s", err.Error())
	}
	err = yaml.Unmarshal(jsonBytes, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parse config file: %s", err.Error())
	}
	return
}
