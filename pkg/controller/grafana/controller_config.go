package grafana

import "sync"

const (
	ConfigGrafanaImage    = "grafana.image.url"
	ConfigGrafanaImageTag = "grafana.image.tag"
)

type ControllerConfig struct {
	Values map[string]string
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Values: map[string]string{},
		}
	})
	return instance
}

func (c *ControllerConfig) AddConfigItem(key, value string) {
	if key != "" && value != "" {
		c.Values[key] = value
	}
}

func (c *ControllerConfig) GetConfigItem(key, defaultValue string) string {
	if c.HasConfigItem(key) {
		return c.Values[key]
	}
	return defaultValue
}

func (c *ControllerConfig) HasConfigItem(key string) bool {
	_, ok := c.Values[key]
	return ok
}
