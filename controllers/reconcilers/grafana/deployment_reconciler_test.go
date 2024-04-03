package grafana

import (
	"fmt"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	config2 "github.com/grafana/grafana-operator/v5/controllers/config"

	"github.com/stretchr/testify/assert"
)

func Test_getGrafanaImage(t *testing.T) {
	cr := &v1beta1.Grafana{
		Spec: v1beta1.GrafanaSpec{
			Version: "",
		},
	}

	expectedDeploymentImage := fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)

	assert.Equal(t, expectedDeploymentImage, getGrafanaImage(cr))
}

func Test_getGrafanaImage_specificVersion(t *testing.T) {
	cr := &v1beta1.Grafana{
		Spec: v1beta1.GrafanaSpec{
			Version: "10.4.0",
		},
	}

	expectedDeploymentImage := fmt.Sprintf("%s:10.4.0", config2.GrafanaImage)

	assert.Equal(t, expectedDeploymentImage, getGrafanaImage(cr))
}

func Test_getGrafanaImage_withEnvironmentOverride(t *testing.T) {
	cr := &v1beta1.Grafana{
		Spec: v1beta1.GrafanaSpec{
			Version: "",
		},
	}

	expectedDeploymentImage := "I want this grafana image"
	t.Setenv("RELATED_IMAGE_GRAFANA", expectedDeploymentImage)

	assert.Equal(t, expectedDeploymentImage, getGrafanaImage(cr))
}
