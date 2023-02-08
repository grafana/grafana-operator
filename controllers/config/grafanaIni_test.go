package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/stretchr/testify/require"
)

const (
	Bar = "bar"
)

var (
	// Server
	enableGzip       = false
	enforceDomain    = false
	ServeFromSubPath = false
	RouterLogging    = false

	// Auth
	loginMaximumInactiveLifetimeDays = 1
	loginMaximumLifetimeDays         = 2
	tokenRotationIntervalMinutes     = 10
	disableLoginForm                 = true
	disableSignoutMenu               = true
	sigV4AuthEnabled                 = true
	oauthAutoLogin                   = true

	// AuthAzureAd
	azureAdEnabled = true
	allowSignUp    = false

	// AuthJwt
	jwtEnabled                 = true
	jwtAutoSignUp              = true
	jwtEnableLoginToken        = true
	jwtRoleAttributeStrict     = true
	jwtUrlLogin                = true
	jwtAllowAssignGrafanaAdmin = true
	jwtSkipOrgRoleSync         = true

	// AuthGenericOauth
	genericOauthEnabled               = true
	genericOauthAllowSignUp           = true
	genericOauthRoleAttributeStrict   = true
	genericOauthTLSSkipVerifyInsecure = true
	genericOauthUsePkce               = true

	// AuthGitlab
	gitlabEnabled                 = true
	gitlabAllowSignUp             = true
	gitlabRoleAttributeStrict     = true
	gitlabAllowAssignGrafanaAdmin = true

	//Dataproxy
	dataProxyDialTimeout                  = 10
	dataProxyExpectContinueTimeoutSeconds = 1
	dataProxyIdleConnTimeoutSeconds       = 90
	dataProxyKeepAliveSeconds             = 30
	dataProxyLogging                      = false
	dataProxyMaxConnsPerHost              = 0
	dataProxyMaxIdleConnections           = 100
	dataProxyResponseLimit                = 0
	dataProxyRowLimit                     = 1000000
	dataProxySendUserHeader               = false
	dataProxyTimeout                      = 30
	dataProxyTlsHandshakeTimeoutSeconds   = 10

	// GrafanaConfigUnifiedAlerting
	enableGrafanaConfigUnifiedAlerting = true
	executeAlerts                      = true
	maxAttempts                        = 2

	// Rendering
	concurrentRenderRequestLimit = 10

	// Live
	maxConnections = 10
)

var testGrafanaConfig = v1alpha1.GrafanaConfig{
	Server: &v1alpha1.GrafanaConfigServer{
		HttpAddr:         "http://grafana",
		HttpPort:         "3000",
		Protocol:         "http",
		Socket:           "socket",
		Domain:           "example.com",
		EnforceDomain:    &enforceDomain,
		RootUrl:          "root_url",
		ServeFromSubPath: &ServeFromSubPath,
		StaticRootPath:   "/",
		EnableGzip:       &enableGzip,
		CertFile:         "/mnt/cert.crt",
		CertKey:          "/mnt/cert.key",
		RouterLogging:    &RouterLogging,
	},
	DataProxy: &v1alpha1.GrafanaConfigDataProxy{
		DialTimeout:                  &dataProxyDialTimeout,
		ExpectContinueTimeoutSeconds: &dataProxyExpectContinueTimeoutSeconds,
		IdleConnTimeoutSeconds:       &dataProxyIdleConnTimeoutSeconds,
		KeepAliveSeconds:             &dataProxyKeepAliveSeconds,
		Logging:                      &dataProxyLogging,
		MaxIdleConnections:           &dataProxyMaxIdleConnections,
		ResponseLimit:                &dataProxyResponseLimit,
		RowLimit:                     &dataProxyRowLimit,
		MaxConnsPerHost:              &dataProxyMaxConnsPerHost,
		SendUserHeader:               &dataProxySendUserHeader,
		Timeout:                      &dataProxyTimeout,
		TlsHandshakeTimeoutSeconds:   &dataProxyTlsHandshakeTimeoutSeconds,
	},
	Database: &v1alpha1.GrafanaConfigDatabase{
		Url:      "Url",
		Type:     "type",
		Path:     "path",
		Host:     "host",
		Name:     "name",
		User:     "user",
		Password: "password",
		SslMode:  "sslMode",
	},
	Auth: &v1alpha1.GrafanaConfigAuth{
		LoginCookieName:                      "grafana_session",
		LoginMaximumInactiveLifetimeDays:     &loginMaximumInactiveLifetimeDays,
		LoginMaximumInactiveLifetimeDuration: "4h",
		LoginMaximumLifetimeDays:             &loginMaximumLifetimeDays,
		LoginMaximumLifetimeDuration:         "8h",
		TokenRotationIntervalMinutes:         &tokenRotationIntervalMinutes,
		DisableLoginForm:                     &disableLoginForm,
		DisableSignoutMenu:                   &disableSignoutMenu,
		SigV4AuthEnabled:                     &sigV4AuthEnabled,
		SignoutRedirectUrl:                   "https://RedirectURL.com",
		OauthAutoLogin:                       &oauthAutoLogin,
	},
	AuthAzureAD: &v1alpha1.GrafanaConfigAuthAzureAD{
		Enabled:        &azureAdEnabled,
		ClientId:       "Client",
		ClientSecret:   "ClientSecret",
		Scopes:         "Scopes",
		AuthUrl:        "https://AuthURL.com",
		TokenUrl:       "https://TokenURL.com",
		AllowedDomains: "azure.com",
		AllowSignUp:    &allowSignUp,
	},
	AuthJwt: &v1alpha1.GrafanaConfigAuthJwt{
		Enabled:                 &jwtEnabled,
		EnableLoginToken:        &jwtEnableLoginToken,
		HeaderName:              "X-JWT-Assertion",
		EmailClaim:              "sub",
		ExpectClaims:            "{\"iss\": \"https://your-token-issuer\", \"your-custom-claim\": \"foo\"}",
		UsernameClaim:           "sub",
		JwkSetUrl:               "https://your-auth-provider.example.com/.well-known/jwks.json",
		JwkSetFile:              "/path/to/jwks.json",
		KeyFile:                 "/path/to/key.pem",
		RoleAttributePath:       "Viewer",
		RoleAttributeStrict:     &jwtRoleAttributeStrict,
		AutoSignUp:              &jwtAutoSignUp,
		CacheTtl:                "60m",
		UrlLogin:                &jwtUrlLogin,
		AllowAssignGrafanaAdmin: &jwtAllowAssignGrafanaAdmin,
		SkipOrgRoleSync:         &jwtSkipOrgRoleSync,
	},
	AuthGenericOauth: &v1alpha1.GrafanaConfigAuthGenericOauth{
		Enabled:               &genericOauthEnabled,
		Name:                  "Name",
		AllowSignUp:           &genericOauthAllowSignUp,
		ClientId:              "ClientOauth",
		ClientSecret:          "ClientSecretOauth",
		Scopes:                "ScopesOauth",
		AuthUrl:               "https://AuthURLOauth.com",
		TokenUrl:              "https://TokenURLOauth.com",
		UsePkce:               &genericOauthUsePkce,
		ApiUrl:                "https://ApiURLOauth.com",
		TeamsURL:              "https://TeamsURLOauth.com",
		TeamIds:               "1,2",
		TeamIdsAttributePath:  "team_ids[*]",
		AllowedDomains:        "mycompanyOauth.com",
		RoleAttributePath:     "roles[*]",
		RoleAttributeStrict:   &genericOauthRoleAttributeStrict,
		EmailAttributePath:    "email",
		TLSSkipVerifyInsecure: &genericOauthTLSSkipVerifyInsecure,
		TLSClientCert:         "/genericOauth/clientCert",
		TLSClientKey:          "/genericOauth/clientKey",
		TLSClientCa:           "/genericOauth/clientCa",
	},
	AuthGitlab: &v1alpha1.GrafanaConfigAuthGitlab{
		Enabled:                 &gitlabEnabled,
		AllowSignUp:             &gitlabAllowSignUp,
		ClientId:                "GITLAB_APPLICATION_ID",
		ClientSecret:            "GITLAB_SECRET",
		Scopes:                  "readAPI",
		AuthUrl:                 "https://gitlab.com/oauth/authorize",
		TokenUrl:                "https://gitlab.com/oauth/token",
		ApiUrl:                  "https://gitlab.com/api/v4",
		AllowedGroups:           "example, foo/bar",
		RoleAttributePath:       "is_admin && 'Admin' || 'Viewer'",
		RoleAttributeStrict:     &gitlabRoleAttributeStrict,
		AllowAssignGrafanaAdmin: &gitlabAllowAssignGrafanaAdmin,
	},
	Live: &v1alpha1.GrafanaConfigLive{
		MaxConnections: &maxConnections,
		AllowedOrigins: "https://origin.com",
	},
	UnifiedAlerting: &v1alpha1.GrafanaConfigUnifiedAlerting{
		Enabled:           &enableGrafanaConfigUnifiedAlerting,
		ExecuteAlerts:     &executeAlerts,
		EvaluationTimeout: "3s",
		MaxAttempts:       &maxAttempts,
		MinInterval:       "1m",
	},
	Rendering: &v1alpha1.GrafanaConfigRendering{
		ServerURL:                    "server_url",
		CallbackURL:                  "callback_url",
		ConcurrentRenderRequestLimit: &concurrentRenderRequestLimit,
	},
	FeatureToggles: &v1alpha1.GrafanaConfigFeatureToggles{
		Enable: "ngalert",
	},
}

var testIni = `[auth]
disable_login_form = true
disable_signout_menu = true
login_cookie_name = grafana_session
login_maximum_inactive_lifetime_days = 1
login_maximum_inactive_lifetime_duration = 4h
login_maximum_lifetime_days = 2
login_maximum_lifetime_duration = 8h
oauth_auto_login = true
signout_redirect_url = https://RedirectURL.com
sigv4_auth_enabled = true
token_rotation_interval_minutes = 10

[auth.azuread]
allow_sign_up = false
allowed_domains = azure.com
auth_url = https://AuthURL.com
client_id = Client
client_secret = ClientSecret
enabled = true
scopes = Scopes
token_url = https://TokenURL.com

[auth.generic_oauth]
allow_sign_up = true
allowed_domains = mycompanyOauth.com
api_url = https://ApiURLOauth.com
auth_url = https://AuthURLOauth.com
client_id = ClientOauth
client_secret = ClientSecretOauth
email_attribute_path = email
enabled = true
name = Name
role_attribute_path = roles[*]
role_attribute_strict = true
scopes = ScopesOauth
team_ids = 1,2
team_ids_attribute_path = team_ids[*]
teams_url = https://TeamsURLOauth.com
tls_client_ca = /genericOauth/clientCa
tls_client_cert = /genericOauth/clientCert
tls_client_key = /genericOauth/clientKey
tls_skip_verify_insecure = true
token_url = https://TokenURLOauth.com
use_pkce = true

[auth.gitlab]
allow_assign_grafana_admin = true
allow_sign_up = true
allowed_groups = example, foo/bar
api_url = https://gitlab.com/api/v4
auth_url = https://gitlab.com/oauth/authorize
client_id = GITLAB_APPLICATION_ID
client_secret = GITLAB_SECRET
enabled = true
role_attribute_path = is_admin && 'Admin' || 'Viewer'
role_attribute_strict = true
scopes = readAPI
token_url = https://gitlab.com/oauth/token

[auth.jwt]
allow_assign_grafana_admin = true
auto_sign_up = true
cache_ttl = 60m
email_claim = sub
enable_login_token = true
enabled = true
expect_claims = {"iss": "https://your-token-issuer", "your-custom-claim": "foo"}
header_name = X-JWT-Assertion
jwk_set_file = /path/to/jwks.json
jwk_set_url = https://your-auth-provider.example.com/.well-known/jwks.json
key_file = /path/to/key.pem
role_attribute_path = Viewer
role_attribute_strict = true
skip_org_role_sync = true
url_login = true
username_claim = sub

[database]
host = host
name = name
password = password
path = path
ssl_mode = sslMode
type = type
url = Url
user = user

[dataproxy]
dialTimeout = 10
expect_continue_timeout_seconds = 1
idle_conn_timeout_seconds = 90
keep_alive_seconds = 30
logging = false
max_conns_per_host = 0
max_idle_connections = 100
response_limit = 0
row_limit = 1000000
send_user_header = false
timeout = 30
tls_handshake_timeout_seconds = 10

[feature_toggles]
enable = ngalert

[live]
allowed_origins = https://origin.com
max_connections = 10

[paths]
data = /var/lib/grafana
logs = /var/log/grafana
plugins = /var/lib/grafana/plugins
provisioning = /etc/grafana/provisioning/

[rendering]
callback_url = callback_url
concurrent_render_request_limit = 10
server_url = server_url

[server]
cert_file = /mnt/cert.crt
cert_key = /mnt/cert.key
domain = example.com
enable_gzip = false
enforce_domain = false
http_addr = http://grafana
http_port = 3000
protocol = http
root_url = root_url
router_logging = false
serve_from_sub_path = false
socket = socket
static_root_path = /

[unified_alerting]
enabled = true
evaluation_timeout = 3s
execute_alerts = true
max_attempts = 2
min_interval = 1m

`

func TestWrite(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	sb, sha := i.Write()

	hash := sha256.New()
	_, err := io.WriteString(hash, testIni)
	require.NoError(t, err)
	require.Equal(t, sb, testIni)
	require.Equal(t, sha, fmt.Sprintf("%x", hash.Sum(nil)))
}

func TestCfgUnifiedAlerting(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgUnifiedAlerting(config)
	testConfig := map[string][]string{
		"unified_alerting": {
			"enabled = true",
			"execute_alerts = true",
			"evaluation_timeout = 3s",
			"max_attempts = 2",
			"min_interval = 1m",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgServer(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgServer(config)
	testConfig := map[string][]string{
		"server": {
			"http_addr = http://grafana",
			"http_port = 3000",
			"protocol = http",
			"socket = socket",
			"domain = example.com",
			"enforce_domain = false",
			"root_url = root_url",
			"serve_from_sub_path = false",
			"static_root_path = /",
			"enable_gzip = false",
			"cert_file = /mnt/cert.crt",
			"cert_key = /mnt/cert.key",
			"router_logging = false",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestDataProxy(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgDataProxy(config)
	testConfig := map[string][]string{
		"dataproxy": {
			"dialTimeout = 10",
			"expect_continue_timeout_seconds = 1",
			"idle_conn_timeout_seconds = 90",
			"keep_alive_seconds = 30",
			"logging = false",
			"max_conns_per_host = 0",
			"max_idle_connections = 100",
			"response_limit = 0",
			"row_limit = 1000000",
			"send_user_header = false",
			"timeout = 30",
			"tls_handshake_timeout_seconds = 10",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgAuth(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgAuth(config)
	testConfig := map[string][]string{
		"auth": {
			"login_cookie_name = grafana_session",
			"login_maximum_inactive_lifetime_days = 1",
			"login_maximum_inactive_lifetime_duration = 4h",
			"login_maximum_lifetime_days = 2",
			"login_maximum_lifetime_duration = 8h",
			"token_rotation_interval_minutes = 10",
			"disable_login_form = true",
			"disable_signout_menu = true",
			"sigv4_auth_enabled = true",
			"signout_redirect_url = https://RedirectURL.com",
			"oauth_auto_login = true",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgLive(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgLive(config)
	testConfig := map[string][]string{
		"live": {
			"max_connections = 10",
			"allowed_origins = https://origin.com",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgAuthAzureAD(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgAuthAzureAD(config)
	testConfig := map[string][]string{
		"auth.azuread": {
			"enabled = true",
			"client_id = Client",
			"client_secret = ClientSecret",
			"scopes = Scopes",
			"auth_url = https://AuthURL.com",
			"token_url = https://TokenURL.com",
			"allowed_domains = azure.com",
			"allow_sign_up = false",
		},
	}
	require.Equal(t, config, testConfig)
}
func TestCfgAuthJwt(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgAuthJwt(config)
	testConfig := map[string][]string{
		"auth.jwt": {
			"enabled = true",
			"enable_login_token = true",
			"header_name = X-JWT-Assertion",
			"email_claim = sub",
			"expect_claims = {\"iss\": \"https://your-token-issuer\", \"your-custom-claim\": \"foo\"}",
			"username_claim = sub",
			"jwk_set_url = https://your-auth-provider.example.com/.well-known/jwks.json",
			"jwk_set_file = /path/to/jwks.json",
			"key_file = /path/to/key.pem",
			"role_attribute_path = Viewer",
			"cache_ttl = 60m",
			"role_attribute_strict = true",
			"auto_sign_up = true",
			"url_login = true",
			"allow_assign_grafana_admin = true",
			"skip_org_role_sync = true",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgDatabase(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgDatabase(config)
	testConfig := map[string][]string{
		"database": {
			"url = Url",
			"type = type",
			"path = path",
			"host = host",
			"name = name",
			"user = user",
			"password = password",
			"ssl_mode = sslMode",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestCfgRendering(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	config := map[string][]string{}
	config = i.cfgRendering(config)
	testConfig := map[string][]string{
		"rendering": {
			"server_url = server_url",
			"callback_url = callback_url",
			"concurrent_render_request_limit = 10",
		},
	}
	require.Equal(t, config, testConfig)
}

func TestAppendBool(t *testing.T) {
	testList := []string{"foo"}
	key := Bar
	value := false
	compareList := []string{"foo", "bar = false"}
	newList := appendBool(testList, key, &value)
	require.NotEqual(t, len(newList), 0)
	require.ElementsMatch(t, newList, compareList)
}

func TestAppendStr(t *testing.T) {
	testList := []string{"foo"}
	key := Bar
	value := "baz"
	compareList := []string{"foo", "bar = baz"}
	newList := appendStr(testList, key, value)
	require.NotEqual(t, len(newList), 0)
	require.ElementsMatch(t, newList, compareList)
}

func TestAppendInt(t *testing.T) {
	testList := []string{"foo"}
	key := Bar
	value := 10
	compareList := []string{"foo", "bar = 10"}
	newList := appendInt(testList, key, &value)
	require.NotEqual(t, len(newList), 0)
	require.ElementsMatch(t, newList, compareList)
}
