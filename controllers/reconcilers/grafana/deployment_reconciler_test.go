package grafana

import (
	"fmt"
	"testing"

	config2 "github.com/grafana/grafana-operator/v5/controllers/config"

	"github.com/stretchr/testify/assert"
)

func Test_getGrafanaImage(t *testing.T) {
	expectedDeploymentImage := fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)

	assert.Equal(t, expectedDeploymentImage, getGrafanaImage())
}

func Test_getGrafanaImage_withEnvironmentOverride(t *testing.T) {
	expectedDeploymentImage := "I want this grafana image"
	t.Setenv("RELATED_IMAGE_GRAFANA", expectedDeploymentImage)

	assert.Equal(t, expectedDeploymentImage, getGrafanaImage())
}
