package config

import (
	"github.com/go-ini/ini"
	testing2 "github.com/integr8ly/grafana-operator/pkg/controller/testing"
	"testing"
)

func TestIniConfig_Build(t *testing.T) {
	cr := testing2.MockCR.DeepCopy()
	cr.Spec.Config.Auth.LoginCookieName = "dummy"
	cr.Spec.Config.Auth.DisableLoginForm = true

	config := NewIniConfig(cr)
	err := config.Build()
	if err != nil {
		t.Error(err)
	}

	parsed, err := ini.Load([]byte(config.Contents))
	if err != nil {
		t.Error(err)
	}

	sect, err := parsed.GetSection("paths")
	if err != nil {
		t.Error(err)
	}

	if key, err := sect.GetKey("data"); err != nil || key.String() != GrafanaDataPath {
		t.Errorf("invalid value for grafana data path")
	}

	if key, err := sect.GetKey("logs"); err != nil || key.String() != GrafanaLogsPath {
		t.Errorf("invalid value for grafana logs path")
	}

	if key, err := sect.GetKey("plugins"); err != nil || key.String() != GrafanaPluginsPath {
		t.Errorf("invalid value for grafana plugins path")
	}

	sect, err = parsed.GetSection("auth")
	if err != nil {
		t.Error(err)
	}

	if key, err := sect.GetKey("disable_login_form"); err != nil || key.String() != "true" {
		t.Errorf("invalid value for disable_login_form")
	}

	if key, err := sect.GetKey("login_cookie_name"); err != nil || key.String() != "dummy" {
		t.Errorf("invalid value for login_cookie_name")
	}

	// We didn't set that key so it should not be present
	if sect.HasKey("signout_redirect_url") {
		t.Errorf("got value for signout_redirect_url but was not expected")
	}
}
