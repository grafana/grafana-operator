package grafana

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	config2 "github.com/grafana-operator/grafana-operator/v5/controllers/config"
	"github.com/grafana-operator/grafana-operator/v5/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	MemoryRequest                           = "256Mi"
	CpuRequest                              = "100m"
	MemoryLimit                             = "1024Mi"
	CpuLimit                                = "500m"
	GrafanaHealthEndpoint                   = "/api/health"
	ReadinessProbeFailureThreshold    int32 = 1
	ReadinessProbeInitialDelaySeconds int32 = 5
	ReadinessProbePeriodSeconds       int32 = 10
	ReadinessProbeSuccessThreshold    int32 = 1
	ReadinessProbeTimeoutSeconds      int32 = 3
)

type DeploymentReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	IsOpenShift bool
}

func GetGrafanaDeploymentMeta(cr *v1beta1.Grafana) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-grafana", cr.Name),
			Namespace: cr.Namespace,
		},
	}
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) error {
	openshiftPlatform := r.IsOpenShift

	deployment := GetGrafanaDeploymentMeta(cr)
	if err := controllerutil.SetControllerReference(cr, deployment, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.Spec = getDeploymentSpec(cr, deployment.Name, openshiftPlatform)
		err := v1beta1.Merge(deployment, cr.Spec.Deployment)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	return nil
}

func getResources() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(MemoryRequest),
			v1.ResourceCPU:    resource.MustParse(CpuRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(MemoryLimit),
			v1.ResourceCPU:    resource.MustParse(CpuLimit),
		},
	}
}

func getVolumes(cr *v1beta1.Grafana) []v1.Volume {
	var volumes []v1.Volume

	config := GetGrafanaIniMeta(cr)

	// Volume to mount the config file from a config map
	volumes = append(volumes, v1.Volume{
		Name: config.Name,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: config.Name,
				},
			},
		},
	})

	// Volume to store the logs
	volumes = append(volumes, v1.Volume{
		Name: config2.GrafanaLogsVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, v1.Volume{
		Name: config2.GrafanaDataVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, v1.Volume{
		Name: "tmp",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	return volumes
}

func getVolumeMounts(cr *v1beta1.Grafana) []v1.VolumeMount {
	var mounts []v1.VolumeMount

	config := GetGrafanaIniMeta(cr)

	mounts = append(mounts, v1.VolumeMount{
		Name:      config.Name,
		MountPath: "/etc/grafana/",
	})

	mounts = append(mounts, v1.VolumeMount{
		Name:      config2.GrafanaDataVolumeName,
		MountPath: config2.GrafanaDataPath,
	})

	mounts = append(mounts, v1.VolumeMount{
		Name:      config2.GrafanaLogsVolumeName,
		MountPath: config2.GrafanaLogsPath,
	})

	mounts = append(mounts, v1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp",
	})

	return mounts
}

func getContainers(cr *v1beta1.Grafana, openshiftPlatform bool) []v1.Container {
	var containers []v1.Container

	image := fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)

	// env var to restart containers if plugins change
	var envVars []v1.EnvVar

	// env var to restart container if config changes
	envVars = append(envVars, v1.EnvVar{
		Name: "CONFIG_HASH",
		// Value: vars.ConfigHash, // TODO
	})

	// env var to restart container if plugins change
	envVars = append(envVars, v1.EnvVar{
		Name:  "GF_INSTALL_PLUGINS",
		Value: cr.Status.Plugins.String(),
	})

	containers = append(containers, v1.Container{
		Name:       "grafana",
		Image:      image,
		Args:       []string{"-config=/etc/grafana/grafana.ini"},
		WorkingDir: "",
		Ports: []v1.ContainerPort{
			{
				Name:          "grafana-http",
				ContainerPort: int32(GetGrafanaPort(cr)),
				Protocol:      "TCP",
			},
		},
		Env:                      envVars,
		Resources:                getResources(),
		VolumeMounts:             getVolumeMounts(cr),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getGrafanaContainerSecurityContext(openshiftPlatform),
		ReadinessProbe:           getReadinessProbe(cr),
	})

	// Use auto generated admin account?
	secret := GetGrafanaAdminSecretMeta(cr)

	for i := 0; i < len(containers); i++ {
		containers[i].Env = append(containers[i].Env, v1.EnvVar{
			Name: config2.GrafanaAdminUserEnvVar,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: config2.GrafanaAdminUserEnvVar,
				},
			},
		})
		containers[i].Env = append(containers[i].Env, v1.EnvVar{
			Name: config2.GrafanaAdminPasswordEnvVar,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: config2.GrafanaAdminPasswordEnvVar,
				},
			},
		})
	}

	return containers
}

// getGrafanaContainerSecurityContext provides default securityContext for grafana container
func getGrafanaContainerSecurityContext(openshiftPlatform bool) *v1.SecurityContext {
	capability := &v1.Capabilities{
		Drop: []v1.Capability{
			"ALL",
		},
	}
	if openshiftPlatform {
		return &v1.SecurityContext{
			AllowPrivilegeEscalation: util.BoolPtr(false),
			ReadOnlyRootFilesystem:   util.BoolPtr(true),
			Privileged:               util.BoolPtr(false),
			RunAsNonRoot:             util.BoolPtr(true),
			Capabilities:             capability,
		}
	}
	return &v1.SecurityContext{
		AllowPrivilegeEscalation: util.BoolPtr(false),
		ReadOnlyRootFilesystem:   util.BoolPtr(true),
		Privileged:               util.BoolPtr(false),
		RunAsNonRoot:             util.BoolPtr(true),
		RunAsUser:                util.IntPtr(10001),
		RunAsGroup:               util.IntPtr(10001),
		Capabilities:             capability,
	}
}

func getReadinessProbe(cr *v1beta1.Grafana) *v1.Probe {
	return &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   GrafanaHealthEndpoint,
				Port:   intstr.FromInt(GetGrafanaPort(cr)),
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: ReadinessProbeInitialDelaySeconds,
		TimeoutSeconds:      ReadinessProbeTimeoutSeconds,
		PeriodSeconds:       ReadinessProbePeriodSeconds,
		SuccessThreshold:    ReadinessProbeSuccessThreshold,
		FailureThreshold:    ReadinessProbeFailureThreshold,
	}
}

func getPodSecurityContext() *v1.PodSecurityContext {
	return &v1.PodSecurityContext{
		SeccompProfile: &v1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
}

func getDeploymentSpec(cr *v1beta1.Grafana, deploymentName string, openshiftPlatform bool) appsv1.DeploymentSpec {
	sa := GetGrafanaServiceAccountMeta(cr) // todo: inlinemodel

	return appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": cr.Name,
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Spec: v1.PodSpec{
				Volumes:            getVolumes(cr),
				Containers:         getContainers(cr, openshiftPlatform),
				SecurityContext:    getPodSecurityContext(),
				ServiceAccountName: sa.Name,
			},
		},
	}
}
