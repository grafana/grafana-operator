package model

import (
	"testing"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
)

func TestGrafanaDeployment_httpProxy(t *testing.T) {
	t.Run("noProxy is setting", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				Deployment: &v1alpha1.GrafanaDeployment{
					HttpProxy: &v1alpha1.GrafanaHttpProxy{
						Enabled:   true,
						URL:       "http://1.2.3.4",
						SecureURL: "http://1.2.3.4",
						NoProxy:   ".svc.cluster.local,.svc",
					},
				},
			},
		}
		deployment := GrafanaDeployment(cr, "", "", "")
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
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				Deployment: &v1alpha1.GrafanaDeployment{
					HttpProxy: &v1alpha1.GrafanaHttpProxy{
						Enabled:   true,
						URL:       "http://1.2.3.4",
						SecureURL: "http://1.2.3.4",
					},
				},
			},
		}
		deployment := GrafanaDeployment(cr, "", "", "")
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

func Test_getLivenessProbe(t *testing.T) {
	t.Run("Default probe", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{},
		}

		got := getLivenessProbe(cr)
		want := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   constants.GrafanaHealthEndpoint,
					Port:   intstr.FromInt(GetGrafanaPort(cr)),
					Scheme: cr.GetScheme(),
				},
			},
			InitialDelaySeconds: LivenessProbeInitialDelaySeconds,
			TimeoutSeconds:      LivenessProbeTimeoutSeconds,
			PeriodSeconds:       LivenessProbePeriodSeconds,
			SuccessThreshold:    LivenessProbeSuccessThreshold,
			FailureThreshold:    LivenessProbeFailureThreshold,
		}

		assert.Equal(t, want, got)
	})

	t.Run("Custom probe", func(t *testing.T) {
		var (
			delay   int32 = 101
			timeout int32 = 102
			period  int32 = 103
			success int32 = 104
			failure int32 = 105
		)

		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				LivenessProbeSpec: &v1alpha1.LivenessProbeSpec{
					InitialDelaySeconds: &delay,
					TimeOutSeconds:      &timeout,
					PeriodSeconds:       &period,
					SuccessThreshold:    &success,
					FailureThreshold:    &failure,
				},
			},
		}

		got := getLivenessProbe(cr)
		want := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   constants.GrafanaHealthEndpoint,
					Port:   intstr.FromInt(GetGrafanaPort(cr)),
					Scheme: cr.GetScheme(),
				},
			},
			InitialDelaySeconds: delay,
			TimeoutSeconds:      timeout,
			PeriodSeconds:       period,
			SuccessThreshold:    success,
			FailureThreshold:    failure,
		}

		assert.Equal(t, want, got)
	})
}

func Test_getReadinessProbe(t *testing.T) {
	t.Run("Default probe", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{},
		}

		got := getReadinessProbe(cr)
		want := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   constants.GrafanaHealthEndpoint,
					Port:   intstr.FromInt(GetGrafanaPort(cr)),
					Scheme: cr.GetScheme(),
				},
			},
			InitialDelaySeconds: ReadinessProbeInitialDelaySeconds,
			TimeoutSeconds:      ReadinessProbeTimeoutSeconds,
			PeriodSeconds:       ReadinessProbePeriodSeconds,
			SuccessThreshold:    ReadinessProbeSuccessThreshold,
			FailureThreshold:    ReadinessProbeFailureThreshold,
		}

		assert.Equal(t, want, got)
	})

	t.Run("Custom probe", func(t *testing.T) {
		var (
			delay   int32 = 101
			timeout int32 = 102
			period  int32 = 103
			success int32 = 104
			failure int32 = 105
		)

		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				ReadinessProbeSpec: &v1alpha1.ReadinessProbeSpec{
					InitialDelaySeconds: &delay,
					TimeOutSeconds:      &timeout,
					PeriodSeconds:       &period,
					SuccessThreshold:    &success,
					FailureThreshold:    &failure,
				},
			},
		}

		got := getReadinessProbe(cr)
		want := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path:   constants.GrafanaHealthEndpoint,
					Port:   intstr.FromInt(GetGrafanaPort(cr)),
					Scheme: cr.GetScheme(),
				},
			},
			InitialDelaySeconds: delay,
			TimeoutSeconds:      timeout,
			PeriodSeconds:       period,
			SuccessThreshold:    success,
			FailureThreshold:    failure,
		}

		assert.Equal(t, want, got)
	})
}

func Test_getTopologySpreadConstraints(t *testing.T) {
	t.Run("Default empty topology constraints", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{},
		}

		got := getTopologySpreadConstraints(cr)
		want := []v1.TopologySpreadConstraint{}

		assert.Equal(t, want, got)
	})

	t.Run("Specified empty constraints", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				Deployment: &v1alpha1.GrafanaDeployment{
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{},
				},
			},
		}

		got := getTopologySpreadConstraints(cr)
		want := []v1.TopologySpreadConstraint{}

		assert.Equal(t, want, got)
	})

	t.Run("Specified non-empty constraints", func(t *testing.T) {
		cr := &v1alpha1.Grafana{
			Spec: v1alpha1.GrafanaSpec{
				Deployment: &v1alpha1.GrafanaDeployment{
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{
						{
							MaxSkew:     1,
							TopologyKey: "topology.kubernetes.io/zone",
						},
					},
				},
			},
		}

		got := getTopologySpreadConstraints(cr)
		want := []v1.TopologySpreadConstraint{
			{
				MaxSkew:     1,
				TopologyKey: "topology.kubernetes.io/zone",
			},
		}

		assert.Equal(t, want, got)
	})
}
