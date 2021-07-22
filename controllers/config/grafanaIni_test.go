package config

import (
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/integr8ly/grafana-operator/api/integreatly/v1alpha1"
	"github.com/stretchr/testify/require"
)

const (
	Bar = "bar"
)

var (
	// Server
	enableGzip = false

	// AuthAzureAd
	azureAdEnabled = true
	allowSignUp    = false

	// Rendering
	concurrentRenderRequestLimit = 10
)

var testGrafanaConfig = v1alpha1.GrafanaConfig{
	Server: &v1alpha1.GrafanaConfigServer{
		EnableGzip: &enableGzip,
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
enable_gzip = false

`

func TestWrite(t *testing.T) {
	i := NewGrafanaIni(&testGrafanaConfig)
	sb, sha := i.Write()

	hash := sha256.New()
	_, err := io.WriteString(hash, testIni)
	require.NoError(t, err)
	require.Equal(t, sha, fmt.Sprintf("%x", hash.Sum(nil)))
	require.Equal(t, sb, testIni)
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
