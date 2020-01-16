package config

// Options passed via cmd line
type Options struct {
	DebugLevel     string
	MetricPort     int
	APIPort        int
	ListenPort     int
	ConfigFilePath string
}
