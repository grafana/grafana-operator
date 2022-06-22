package model

import (
	"testing"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestGrafanaDeployment_httpProxy(t *testing.T) {
	t.Run("noProxy is setting", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{
			Spec: grafanav1alpha1.GrafanaSpec{
				Deployment: &grafanav1alpha1.GrafanaDeployment{
					HttpProxy: &grafanav1alpha1.GrafanaHttpProxy{
						Enabled:   true,
						URL:       "http://1.2.3.4",
						SecureURL: "http://1.2.3.4",
						NoProxy:   ".svc.cluster.local,.svc",
					},
				},
			},
		}
		deployment := GrafanaDeployment(cr, "", "")
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name != "grafana" {
				continue
			}

			noProxyExist := false
			for _, env := range container.Env {
				if env.Name == "NO_PROXY" {
					noProxyExist = true
					assert.Equal(t, ".svc.cluster.local,.svc", env.Value)
				}
			}
			assert.True(t, noProxyExist)
		}
	})

	t.Run("noProxy is not setting", func(t *testing.T) {
		cr := &grafanav1alpha1.Grafana{
			Spec: grafanav1alpha1.GrafanaSpec{
				Deployment: &grafanav1alpha1.GrafanaDeployment{
					HttpProxy: &grafanav1alpha1.GrafanaHttpProxy{
						Enabled:   true,
						URL:       "http://1.2.3.4",
						SecureURL: "http://1.2.3.4",
					},
				},
			},
		}
		deployment := GrafanaDeployment(cr, "", "")
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name != "grafana" {
				continue
			}

			noProxyExist := false
			for _, env := range container.Env {
				if env.Name == "NO_PROXY" {
					noProxyExist = true
				}
			}
			assert.False(t, noProxyExist)
		}
	})
}
