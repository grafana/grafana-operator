package config

const (
	// Grafana
	GrafanaImage   = "docker.io/grafana/grafana"
	GrafanaVersion = "7.5.11"

	// Paths
	GrafanaDataPath         = "/var/lib/grafana"
	GrafanaLogsPath         = "/var/log/grafana"
	GrafanaPluginsPath      = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath = "/etc/grafana/provisioning/"

	// Admin user
	DefaultAdminUser           = "admin"
	GrafanaAdminUserEnvVar     = "GF_SECURITY_ADMIN_USER"
	GrafanaAdminPasswordEnvVar = "GF_SECURITY_ADMIN_PASSWORD" // #nosec G101

	// Networking
	GrafanaHttpPort     int = 3000
	GrafanaHttpPortName     = "grafana"

	// Data storage
	GrafanaProvisionPluginVolumeName    = "grafana-provision-plugins"
	GrafanaPluginsVolumeName            = "grafana-plugins"
	GrafanaProvisionDashboardVolumeName = "grafana-provision-dashboards"
	GrafanaProvisionNotifierVolumeName  = "grafana-provision-notifiers"
	GrafanaLogsVolumeName               = "grafana-logs"
	GrafanaDataVolumeName               = "grafana-data"
	SecretsMountDir                     = "/etc/grafana-secrets/" // #nosec G101
	ConfigMapsMountDir                  = "/etc/grafana-configmaps/"
)
