package config

import (
	"crypto/md5"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
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
		items = appendStr(items, "type", i.cfg.RemoteCache.ConnStr)
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

	sb := strings.Builder{}

	var keys []string
	for key, _ := range config {
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

	hash := md5.New()
	io.WriteString(hash, sb.String())

	return sb.String(), fmt.Sprintf("%x", hash.Sum(nil))
}
