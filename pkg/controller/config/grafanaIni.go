package config

import (
	"crypto/sha256"
	"fmt"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"io"
	"sort"
	"strings"
)

type GrafanaIni struct {
	cfg *v1alpha1.GrafanaConfig
}

func NewGrafanaIni(cfg *v1alpha1.GrafanaConfig) *GrafanaIni {
	return &GrafanaIni{
		cfg: cfg,
	}
}

func (i *GrafanaIni) Write() (string, string) {
	config := map[string][]string{}

	appendStr := func(l []string, key, value string) []string {
		if value != "" {
			return append(l, fmt.Sprintf("%v = %v", key, value))
		}
		return l
	}

	appendInt := func(l []string, key string, value *int) []string {
		if value != nil {
			return append(l, fmt.Sprintf("%v = %v", key, *value))
		}
		return l
	}

	appendBool := func(l []string, key string, value *bool) []string {
		if value != nil {
			return append(l, fmt.Sprintf("%v = %v", key, *value))
		}
		return l
	}

	config["paths"] = []string{
		fmt.Sprintf("data = %v", GrafanaDataPath),
		fmt.Sprintf("logs = %v", GrafanaLogsPath),
		fmt.Sprintf("plugins = %v", GrafanaPluginsPath),
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
		items = appendStr(items, "protocol", i.cfg.Server.Protocol)
		items = appendStr(items, "socket", i.cfg.Server.Socket)
		items = appendStr(items, "domain", i.cfg.Server.Domain)
		items = appendBool(items, "enforce_domain", i.cfg.Server.EnforceDomain)
		items = appendStr(items, "root_url", i.cfg.Server.RootUrl)
		items = appendBool(items, "serve_from_sub_path", i.cfg.Server.ServeFromSubPath)
		items = appendStr(items, "static_root_path", i.cfg.Server.StaticRootPath)
		items = appendBool(items, "enable_gzip", i.cfg.Server.EnableGzip)
		items = appendStr(items, "cert_file", i.cfg.Server.CertFile)
		items = appendStr(items, "cert_key", i.cfg.Server.CertKey)
		items = appendBool(items, "router_logging", i.cfg.Server.RouterLogging)
		config["server"] = items
	}

	if i.cfg.Database != nil {
		var items []string
		items = appendStr(items, "url", i.cfg.Database.Url)
		items = appendStr(items, "type", i.cfg.Database.Type)
		items = appendStr(items, "path", i.cfg.Database.Path)
		items = appendStr(items, "host", i.cfg.Database.Host)
		items = appendStr(items, "name", i.cfg.Database.Name)
		items = appendStr(items, "user", i.cfg.Database.User)
		items = appendStr(items, "password", i.cfg.Database.Password)
		items = appendStr(items, "ssl_mode", i.cfg.Database.SslMode)
		items = appendStr(items, "ca_cert_path", i.cfg.Database.CaCertPath)
		items = appendStr(items, "client_key_path", i.cfg.Database.ClientKeyPath)
		items = appendStr(items, "client_cert_path", i.cfg.Database.ClientCertPath)
		items = appendStr(items, "server_cert_name", i.cfg.Database.ServerCertName)
		items = appendInt(items, "max_idle_conn", i.cfg.Database.MaxIdleConn)
		items = appendInt(items, "max_open_conn", i.cfg.Database.MaxOpenConn)
		items = appendInt(items, "conn_max_lifetime", i.cfg.Database.ConnMaxLifetime)
		items = appendBool(items, "log_queries", i.cfg.Database.LogQueries)
		items = appendStr(items, "cache_mode", i.cfg.Database.CacheMode)
		config["database"] = items
	}

	if i.cfg.RemoteCache != nil {
		var items []string
		items = appendStr(items, "type", i.cfg.RemoteCache.Type)
		items = appendStr(items, "connstr", i.cfg.RemoteCache.ConnStr)
		config["remote_cache"] = items
	}

	if i.cfg.Security != nil {
		var items []string
		items = appendStr(items, "admin_user", i.cfg.Security.AdminUser)
		items = appendStr(items, "admin_password", i.cfg.Security.AdminPassword)
		items = appendInt(items, "login_remember_days", i.cfg.Security.LoginRememberDays)
		items = appendStr(items, "secret_key", i.cfg.Security.SecretKey)
		items = appendBool(items, "disable_gravatar", i.cfg.Security.DisableGravatar)
		items = appendStr(items, "data_source_proxy_whitelist", i.cfg.Security.DataSourceProxyWhitelist)
		items = appendBool(items, "cookie_secure", i.cfg.Security.CookieSecure)
		items = appendStr(items, "cookie_samesite", i.cfg.Security.CookieSamesite)
		items = appendBool(items, "allow_embedding", i.cfg.Security.AllowEmbedding)
		items = appendBool(items, "strict_transport_security", i.cfg.Security.StrictTransportSecurity)
		items = appendInt(items, "strict_transport_security_max_age_seconds", i.cfg.Security.StrictTransportSecurityMaxAgeSeconds)
		items = appendBool(items, "strict_transport_security_preload", i.cfg.Security.StrictTransportSecurityPreload)
		items = appendBool(items, "strict_transport_security_subdomains", i.cfg.Security.StrictTransportSecuritySubdomains)
		items = appendBool(items, "x_content_type_options", i.cfg.Security.XContentTypeOptions)
		items = appendBool(items, "x_xss_protection", i.cfg.Security.XXssProtection)
		config["security"] = items
	}

	if i.cfg.Users != nil {
		var items []string
		items = appendBool(items, "allow_sign_up", i.cfg.Users.AllowSignUp)
		items = appendBool(items, "allow_org_create", i.cfg.Users.AllowOrgCreate)
		items = appendBool(items, "auto_assign_org", i.cfg.Users.AutoAssignOrg)
		items = appendStr(items, "auto_assign_org_id", i.cfg.Users.AutoAssignOrgId)
		items = appendStr(items, "auto_assign_org_role", i.cfg.Users.AutoAssignOrgRole)
		items = appendBool(items, "viewers_can_edit", i.cfg.Users.ViewersCanEdit)
		items = appendBool(items, "editors_can_admin", i.cfg.Users.EditorsCanAdmin)
		items = appendStr(items, "login_hint", i.cfg.Users.LoginHint)
		items = appendStr(items, "password_hint", i.cfg.Users.PasswordHint)
		config["users"] = items
	}

	if i.cfg.AuthBasic != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthBasic.Enabled)
		config["auth.basic"] = items
	}

	if i.cfg.AuthAnonymous != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthAnonymous.Enabled)
		items = appendStr(items, "org_name", i.cfg.AuthAnonymous.OrgName)
		items = appendStr(items, "org_role", i.cfg.AuthAnonymous.OrgRole)
		config["auth.anonymous"] = items
	}

	if i.cfg.Auth != nil {
		var items []string
		items = appendStr(items, "login_cookie_name", i.cfg.Auth.LoginCookieName)
		items = appendInt(items, "login_maximum_inactive_lifetime_days", i.cfg.Auth.LoginMaximumInactiveLifetimeDays)
		items = appendInt(items, "login_maximum_lifetime_days", i.cfg.Auth.LoginMaximumLifetimeDays)
		items = appendInt(items, "token_rotation_interval_minutes", i.cfg.Auth.TokenRotationIntervalMinutes)
		items = appendBool(items, "disable_login_form", i.cfg.Auth.DisableLoginForm)
		items = appendBool(items, "disable_signout_menu", i.cfg.Auth.DisableSignoutMenu)
		items = appendStr(items, "signout_redirect_url", i.cfg.Auth.SignoutRedirectUrl)
		items = appendBool(items, "oauth_auto_login", i.cfg.Auth.OauthAutoLogin)
		config["auth"] = items
	}

	if i.cfg.Log != nil {
		var items []string
		items = appendStr(items, "mode", i.cfg.Log.Mode)
		items = appendStr(items, "level", i.cfg.Log.Level)
		items = appendStr(items, "filters", i.cfg.Log.Filters)
		config["log"] = items
	}

	if i.cfg.AuthGoogle != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthGoogle.Enabled)
		items = appendStr(items, "client_id", i.cfg.AuthGoogle.ClientId)
		items = appendStr(items, "client_secret", i.cfg.AuthGoogle.ClientSecret)
		items = appendStr(items, "scopes", i.cfg.AuthGoogle.Scopes)
		items = appendStr(items, "auth_url", i.cfg.AuthGoogle.AuthUrl)
		items = appendStr(items, "token_url", i.cfg.AuthGoogle.TokenUrl)
		items = appendStr(items, "allowed_domains", i.cfg.AuthGoogle.AllowedDomains)
		items = appendBool(items, "allow_sign_up", i.cfg.AuthGoogle.AllowSignUp)
		config["auth.google"] = items
	}

	if i.cfg.AuthGithub != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthGithub.Enabled)
		items = appendBool(items, "allow_sign_up", i.cfg.AuthGithub.AllowSignUp)
		items = appendStr(items, "client_id", i.cfg.AuthGithub.ClientId)
		items = appendStr(items, "client_secret", i.cfg.AuthGithub.ClientSecret)
		items = appendStr(items, "scopes", i.cfg.AuthGithub.Scopes)
		items = appendStr(items, "auth_url", i.cfg.AuthGithub.AuthUrl)
		items = appendStr(items, "token_url", i.cfg.AuthGithub.TokenUrl)
		items = appendStr(items, "api_url", i.cfg.AuthGithub.ApiUrl)
		items = appendStr(items, "team_ids", i.cfg.AuthGithub.TeamIds)
		items = appendStr(items, "allowed_organizations", i.cfg.AuthGithub.AllowedOrganizations)
		config["auth.github"] = items
	}

	if i.cfg.AuthGitlab != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthGitlab.Enabled)
		items = appendBool(items, "allow_sign_up", i.cfg.AuthGitlab.AllowSignUp)
		items = appendStr(items, "client_id", i.cfg.AuthGitlab.ClientId)
		items = appendStr(items, "client_secret", i.cfg.AuthGitlab.ClientSecret)
		items = appendStr(items, "scopes", i.cfg.AuthGitlab.Scopes)
		items = appendStr(items, "auth_url", i.cfg.AuthGitlab.AuthUrl)
		items = appendStr(items, "token_url", i.cfg.AuthGitlab.TokenUrl)
		items = appendStr(items, "api_url", i.cfg.AuthGitlab.ApiUrl)
		config["auth.gitlab"] = items
	}

	if i.cfg.AuthGenericOauth != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthGenericOauth.Enabled)
		items = appendBool(items, "allow_sign_up", i.cfg.AuthGenericOauth.AllowSignUp)
		items = appendStr(items, "client_id", i.cfg.AuthGenericOauth.ClientId)
		items = appendStr(items, "client_secret", i.cfg.AuthGenericOauth.ClientSecret)
		items = appendStr(items, "scopes", i.cfg.AuthGenericOauth.Scopes)
		items = appendStr(items, "auth_url", i.cfg.AuthGenericOauth.AuthUrl)
		items = appendStr(items, "token_url", i.cfg.AuthGenericOauth.TokenUrl)
		items = appendStr(items, "api_url", i.cfg.AuthGenericOauth.ApiUrl)
		items = appendStr(items, "allowed_domains", i.cfg.AuthGenericOauth.AllowedDomains)
		items = appendStr(items, "role_attribute_path", i.cfg.AuthGenericOauth.RoleAttributePath)
		items = appendStr(items, "email_attribute_path", i.cfg.AuthGenericOauth.EmailAttributePath)
		config["auth.generic_oauth"] = items
	}

	if i.cfg.AuthLdap != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthLdap.Enabled)
		items = appendBool(items, "allow_sign_up", i.cfg.AuthLdap.AllowSignUp)
		items = appendStr(items, "config_file", i.cfg.AuthLdap.ConfigFile)
		config["auth.ldap"] = items
	}

	if i.cfg.AuthProxy != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.AuthProxy.Enabled)
		items = appendStr(items, "header_name", i.cfg.AuthProxy.HeaderName)
		items = appendStr(items, "header_property", i.cfg.AuthProxy.HeaderProperty)
		items = appendBool(items, "auto_sign_up", i.cfg.AuthProxy.AutoSignUp)
		items = appendStr(items, "ldap_sync_ttl", i.cfg.AuthProxy.LdapSyncTtl)
		items = appendStr(items, "whitelist", i.cfg.AuthProxy.Whitelist)
		items = appendStr(items, "headers", i.cfg.AuthProxy.Headers)
		items = appendBool(items, "enable_login_token", i.cfg.AuthProxy.EnableLoginToken)
		config["auth.proxy"] = items
	}

	if i.cfg.DataProxy != nil {
		var items []string
		items = appendBool(items, "logging", i.cfg.DataProxy.Logging)
		items = appendInt(items, "timeout", i.cfg.DataProxy.Timeout)
		items = appendBool(items, "send_user_header", i.cfg.DataProxy.SendUserHeader)
		config["dataproxy"] = items
	}

	if i.cfg.Analytics != nil {
		var items []string
		items = appendBool(items, "reporting_enabled", i.cfg.Analytics.ReportingEnabled)
		items = appendStr(items, "google_analytics_ua_id", i.cfg.Analytics.GoogleAnalyticsUaId)
		items = appendBool(items, "check_for_updates", i.cfg.Analytics.CheckForUpdates)
		config["analytics"] = items
	}

	if i.cfg.Dashboards != nil {
		var items []string
		items = appendInt(items, "versions_to_keep", i.cfg.Dashboards.VersionsToKeep)
		config["dashboards"] = items
	}

	if i.cfg.Smtp != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.Smtp.Enabled)
		items = appendStr(items, "host", i.cfg.Smtp.Host)
		items = appendStr(items, "user", i.cfg.Smtp.User)
		items = appendStr(items, "password", i.cfg.Smtp.Password)
		items = appendStr(items, "cert_file", i.cfg.Smtp.CertFile)
		items = appendStr(items, "key_file", i.cfg.Smtp.KeyFile)
		items = appendBool(items, "skip_verify", i.cfg.Smtp.SkipVerify)
		items = appendStr(items, "from_address", i.cfg.Smtp.FromAddress)
		items = appendStr(items, "from_name", i.cfg.Smtp.FromName)
		items = appendStr(items, "ehlo_identity", i.cfg.Smtp.EhloIdentity)
		config["smtp"] = items
	}

	if i.cfg.Metrics != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.Metrics.Enabled)
		items = appendStr(items, "basic_auth_username", i.cfg.Metrics.BasicAuthUsername)
		items = appendStr(items, "basic_auth_password", i.cfg.Metrics.BasicAuthPassword)
		items = appendInt(items, "interval_seconds", i.cfg.Metrics.IntervalSeconds)
		config["metrics"] = items
	}

	if i.cfg.Snapshots != nil {
		var items []string
		items = appendBool(items, "external_enabled", i.cfg.Snapshots.ExternalEnabled)
		items = appendStr(items, "external_snapshot_url", i.cfg.Snapshots.ExternalSnapshotUrl)
		items = appendStr(items, "external_snapshot_name", i.cfg.Snapshots.ExternalSnapshotName)
		items = appendBool(items, "snapshot_remove_expired", i.cfg.Snapshots.SnapshotRemoveExpired)
		config["snapshots"] = items
	}

	if i.cfg.MetricsGraphite != nil {
		var items []string
		items = appendStr(items, "address", i.cfg.MetricsGraphite.Address)
		items = appendStr(items, "prefix", i.cfg.MetricsGraphite.Prefix)
		config["metrics.graphite"] = items
	}

	if i.cfg.ExternalImageStorage != nil {
		var items []string
		items = appendStr(items, "provider", i.cfg.ExternalImageStorage.Provider)
		config["external_image_storage"] = items
	}

	if i.cfg.ExternalImageStorageS3 != nil {
		var items []string
		items = appendStr(items, "bucket", i.cfg.ExternalImageStorageS3.Bucket)
		items = appendStr(items, "region", i.cfg.ExternalImageStorageS3.Region)
		items = appendStr(items, "path", i.cfg.ExternalImageStorageS3.Path)
		items = appendStr(items, "bucket_url", i.cfg.ExternalImageStorageS3.BucketUrl)
		items = appendStr(items, "access_key", i.cfg.ExternalImageStorageS3.AccessKey)
		items = appendStr(items, "secret_key", i.cfg.ExternalImageStorageS3.SecretKey)
		config["external_image_storage.s3"] = items
	}

	if i.cfg.ExternalImageStorageWebdav != nil {
		var items []string
		items = appendStr(items, "url", i.cfg.ExternalImageStorageWebdav.Url)
		items = appendStr(items, "public_url", i.cfg.ExternalImageStorageWebdav.PublicUrl)
		items = appendStr(items, "username", i.cfg.ExternalImageStorageWebdav.Username)
		items = appendStr(items, "password", i.cfg.ExternalImageStorageWebdav.Password)
		config["external_image_storage.webdav"] = items
	}

	if i.cfg.ExternalImageStorageGcs != nil {
		var items []string
		items = appendStr(items, "key_file", i.cfg.ExternalImageStorageGcs.KeyFile)
		items = appendStr(items, "bucket", i.cfg.ExternalImageStorageGcs.Bucket)
		items = appendStr(items, "path", i.cfg.ExternalImageStorageGcs.Path)
		config["external_image_storage.gcs"] = items
	}

	if i.cfg.ExternalImageStorageAzureBlob != nil {
		var items []string
		items = appendStr(items, "account_name", i.cfg.ExternalImageStorageAzureBlob.AccountName)
		items = appendStr(items, "account_key", i.cfg.ExternalImageStorageAzureBlob.AccountKey)
		items = appendStr(items, "container_name", i.cfg.ExternalImageStorageAzureBlob.ContainerName)
		config["external_image_storage.azure_blob"] = items
	}

	if i.cfg.Alerting != nil {
		var items []string
		items = appendBool(items, "enabled", i.cfg.Alerting.Enabled)
		items = appendBool(items, "execute_alerts", i.cfg.Alerting.ExecuteAlerts)
		items = appendStr(items, "error_or_timeout", i.cfg.Alerting.ErrorOrTimeout)
		items = appendStr(items, "nodata_or_nullvalues", i.cfg.Alerting.NodataOrNullvalues)
		items = appendInt(items, "concurrent_render_limit", i.cfg.Alerting.ConcurrentRenderLimit)
		items = appendInt(items, "evaluation_timeout_seconds", i.cfg.Alerting.EvaluationTimeoutSeconds)
		items = appendInt(items, "notification_timeout_seconds", i.cfg.Alerting.NotificationTimeoutSeconds)
		items = appendInt(items, "max_attempts", i.cfg.Alerting.MaxAttempts)
		config["alerting"] = items
	}

	if i.cfg.Panels != nil {
		var items []string
		items = appendBool(items, "disable_sanitize_html", i.cfg.Panels.DisableSanitizeHtml)
		config["panels"] = items
	}

	if i.cfg.Plugins != nil {
		var items []string
		items = appendBool(items, "enable_alpha", i.cfg.Plugins.EnableAlpha)
		config["plugins"] = items
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
