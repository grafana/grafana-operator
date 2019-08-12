package grafana

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
)

const (
	PathsSectionName         = "paths"
	SecuritySectionName      = "security"
	LogSectionName           = "log"
	AuthSectionName          = "auth"
	AuthBasicSectionName     = "auth.basic"
	AuthAnonymousSectionName = "auth.anonymous"
)

type IniConfig struct {
	Cr       *v1alpha1.Grafana
	Contents string
	Hash     string
}

// Those paths are fixed and must not be changed
// The grafana deployment relies on volumes mounted at the exact locations
func (i *IniConfig) appendPathsSection(config *ini.File) error {
	section, err := config.NewSection(PathsSectionName)
	if err != nil {
		return err
	}

	section.NewKey("data", common.GrafanaDataPath)
	section.NewKey("logs", common.GrafanaLogsPath)
	section.NewKey("plugins", common.GrafanaPluginsPath)
	section.NewKey("provisioning", common.GrafanaProvisioningPath)
	return nil
}

// The CR still supports a number of config properties defined outside of the `config` object
// To retain backwards compatibility we still support them, but as a second priority: if the same
// fields are also set in `config` then those will overwrite the legacy fields
func (i *IniConfig) appendLegacyConfig(config *ini.File) error {
	// [security]
	securitySection, err := config.NewSection(SecuritySectionName)
	if err != nil {
		return err
	}

	// Only set the keys if the string is not empty
	if i.Cr.Spec.AdminUser != "" {
		securitySection.NewKey("admin_user", i.Cr.Spec.AdminUser)
	}

	if i.Cr.Spec.AdminPassword != "" {
		securitySection.NewKey("admin_password", i.Cr.Spec.AdminPassword)
	}

	// [auth]
	authSection, err := config.NewSection(AuthSectionName)
	if err != nil {
		return err
	}
	authSection.NewKey("disable_login_form", fmt.Sprintf("%v", i.Cr.Spec.DisableLoginForm))
	authSection.NewKey("disable_signout_menu", fmt.Sprintf("%v", i.Cr.Spec.DisableSignoutMenu))

	// [auth.basic]
	authBasicSection, err := config.NewSection(AuthBasicSectionName)
	if err != nil {
		return err
	}
	authBasicSection.NewKey("enabled", fmt.Sprintf("%v", i.Cr.Spec.BasicAuth))

	// [auth.anonymous]
	authAnonymousSection, err := config.NewSection(AuthAnonymousSectionName)
	if err != nil {
		return err
	}
	authAnonymousSection.NewKey("enabled", fmt.Sprintf("%v", i.Cr.Spec.Anonymous))

	// [log]
	logSection, err := config.NewSection(LogSectionName)
	if err != nil {
		return err
	}

	if i.Cr.Spec.LogLevel != "" {
		logSection.NewKey("level", i.Cr.Spec.LogLevel)
	}

	return nil
}

func (i *IniConfig) buildBaseConfig(config *ini.File) error {
	// Add the legacy config first: it can be overwritten if the same
	// properties are defined in the `config` section
	err := i.appendLegacyConfig(config)
	if err != nil {
		return err
	}

	// Import all properties from the CR
	err = config.ReflectFrom(&i.Cr.Spec.Config)
	if err != nil {
		return err
	}

	// Always append the paths section last because we do not
	// allow to override it
	err = i.appendPathsSection(config)
	if err != nil {
		return err
	}

	return nil
}

// Creates the ini config from the CR
func (i *IniConfig) Build() error {
	config := ini.Empty()
	err := i.buildBaseConfig(config)
	if err != nil {
		return err
	}

	s := bytes.NewBufferString("")
	_, err = config.WriteTo(s)
	if err != nil {
		return err
	}

	i.Contents = s.String()
	i.Hash = fmt.Sprintf("%x", md5.Sum([]byte(i.Contents)))
	return nil
}

func (i *IniConfig) DiffersFrom(lastConfigHash string) bool {
	return i.Hash != lastConfigHash
}

func NewIniConfig(cr *v1alpha1.Grafana) *IniConfig {
	return &IniConfig{
		Cr:       cr,
		Contents: "",
		Hash:     "",
	}
}
