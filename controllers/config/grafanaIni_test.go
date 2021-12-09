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

	// AuthAzureAd
	azureAdEnabled = true
	allowSignUp    = false

	// GrafanaConfigUnifiedAlerting
	enableGrafanaConfigUnifiedAlerting = true
	executeAlerts                      = true
	maxAttempts                        = 2

	// Rendering
	concurrentRenderRequestLimit = 10
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

var testIni = `[auth.azuread]
allow_sign_up = false
allowed_domains = azure.com
auth_url = https://AuthURL.com
client_id = Client
client_secret = ClientSecret
enabled = true
scopes = Scopes
token_url = https://TokenURL.com

[database]
host = host
name = name
password = password
path = path
ssl_mode = sslMode
type = type
url = Url
user = user

[feature_toggles]
enable = ngalert

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
