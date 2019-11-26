package config

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
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
	return nil
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
