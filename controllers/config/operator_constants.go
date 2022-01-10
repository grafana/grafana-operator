package config

const (
	// Paths
	GrafanaDataPath         = "/var/lib/grafana"
	GrafanaLogsPath         = "/var/log/grafana"
	GrafanaPluginsPath      = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath = "/etc/grafana/provisioning/"

	// Admin user
	DefaultAdminUser           = "admin"
	GrafanaAdminUserEnvVar     = "GF_SECURITY_ADMIN_USER"
	GrafanaAdminPasswordEnvVar = "GF_SECURITY_ADMIN_PASSWORD" // #nosec G101
)
