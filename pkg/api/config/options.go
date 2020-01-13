package config

// Options passed via cmd line
type Options struct {
	DebugLevel     string
	MetricPort     int
	ListenPort     int
	ConfigFilePath string
}
