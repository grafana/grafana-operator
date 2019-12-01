package config

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/pkg/errors"
)

const (
	PathsSectionName = "paths"
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

	section.NewKey("data", GrafanaDataPath)
	section.NewKey("logs", GrafanaLogsPath)
	section.NewKey("plugins", GrafanaPluginsPath)
	section.NewKey("provisioning", GrafanaProvisioningPath)
	return nil
}

// We need to set default values for all boolean properties because otherwise
// they would automatically get assigned a default value of false
func (i *IniConfig) setDefaults(config *ini.File) error {
	// Default values for all boolean settings, taken from:
	// https://grafana.com/docs/installation/configuration/
	_, err := config.Section("auth.basic").NewKey("enabled", "true")
	if err != nil {
		return err
	}



	_, err = config.Section("auth.proxy").NewKey("auto_sign_up", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("analytics").NewKey("reporting_enabled", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("analytics").NewKey("check_for_updates", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("metrics").NewKey("enabled", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("snapshots").NewKey("external_enabled", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("alerting").NewKey("enabled", "true")
	if err != nil {
		return err
	}

	_, err = config.Section("alerting").NewKey("execute_alerts", "true")
	if err != nil {
		return err
	}
	return err
}

func (i *IniConfig) buildBaseConfig(config *ini.File) error {
	// Import all properties from the CR
	if err := config.ReflectFrom(&i.Cr.Spec.Config); err != nil {
		return err
	}

	// Always append the paths section last because we do not
	// allow to override it
	return i.appendPathsSection(config)
}

// Creates the ini config from the CR
func (i *IniConfig) _Build() error {
	config := ini.Empty()
	config.ValueMapper = func (string) string {
		return ""
	}

	// Prepopulate default values
	err := i.setDefaults(config)
	if err != nil {
		return errors.Wrap(err, "error setting default values")
	}

	err = i.buildBaseConfig(config)
	if err != nil {
		return errors.Wrap(err, "error reading configuration")
	}

	s := bytes.NewBufferString("")
	_, err = config.WriteTo(s)
	if err != nil {
		return errors.Wrap(err, "error writing configuration")
	}

	i.Contents = s.String()
	i.Hash = fmt.Sprintf("%x", md5.Sum([]byte(i.Contents)))
	return nil
}

func (i *IniConfig) _DiffersFrom(lastConfigHash string) bool {
	return i.Hash != lastConfigHash
}

func _NewIniConfig(cr *v1alpha1.Grafana) *IniConfig {
	return &IniConfig{
		Cr:       cr,
		Contents: "",
		Hash:     "",
	}
}
