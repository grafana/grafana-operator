package connectors

type GrafanaAuthProxyConnector struct {
	Type   string      `yaml:"type"`
	ID     string      `yaml:"id"`
	Name   string      `yaml:"name"`
	Config interface{} `yaml:"config"`
}
