package config

const (
	// Grafana
	GrafanaImage   = "docker.io/grafana/grafana"
	GrafanaVersion = "9.5.17"

	// Paths
	GrafanaDataPath               = "/var/lib/grafana"
	GrafanaLogsPath               = "/var/log/grafana"
	GrafanaPluginsPath            = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath       = "/etc/grafana/provisioning/"
	GrafanaDashboardsRuntimeBuild = "/tmp/dashboards"

	// Grafana env vars and admin user
	DefaultAdminUser           = "admin"
	GrafanaAdminUserEnvVar     = "GF_SECURITY_ADMIN_USER"
	GrafanaAdminPasswordEnvVar = "GF_SECURITY_ADMIN_PASSWORD" // #nosec G101
	GrafanaPluginsEnvVar       = "GF_INSTALL_PLUGINS"

	// Networking
	GrafanaHttpPort       int = 3000
	GrafanaHttpPortName       = "grafana"
	GrafanaServerProtocol     = "http"

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
