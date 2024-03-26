package grafana

import (
	"context"
	"fmt"
	"os"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	config2 "github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MemoryRequest                           = "256Mi"
	CpuRequest                              = "100m"
	MemoryLimit                             = "1024Mi"
	GrafanaHealthEndpoint                   = "/api/health"
	ReadinessProbeFailureThreshold    int32 = 1
	ReadinessProbeInitialDelaySeconds int32 = 5
	ReadinessProbePeriodSeconds       int32 = 10
	ReadinessProbeSuccessThreshold    int32 = 1
	ReadinessProbeTimeoutSeconds      int32 = 3
)

type DeploymentReconciler struct {
	client      client.Client
	isOpenShift bool
}

func NewDeploymentReconciler(client client.Client, isOpenShift bool) reconcilers.OperatorGrafanaReconciler {
	return &DeploymentReconciler{
		client:      client,
		isOpenShift: isOpenShift,
	}
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx).WithName("DeploymentReconciler")

	openshiftPlatform := r.isOpenShift
	logger.Info("reconciling deployment", "openshift", openshiftPlatform)

	deployment := model.GetGrafanaDeployment(cr, scheme)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		deployment.Spec = getDeploymentSpec(cr, deployment.Name, scheme, vars, openshiftPlatform)
		err := v1beta1.Merge(deployment, cr.Spec.Deployment)
		return err
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getResources() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(MemoryRequest),
			v1.ResourceCPU:    resource.MustParse(CpuRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(MemoryLimit),
		},
	}
}

func getVolumes(cr *v1beta1.Grafana, scheme *runtime.Scheme) []v1.Volume {
	var volumes []v1.Volume

	config := model.GetGrafanaConfigMap(cr, scheme)

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

	return volumes
}

func getVolumeMounts(cr *v1beta1.Grafana, scheme *runtime.Scheme) []v1.VolumeMount {
	var mounts []v1.VolumeMount

	config := model.GetGrafanaConfigMap(cr, scheme)

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

	return mounts
}

func getGrafanaImage(cr *v1beta1.Grafana) string {
	if cr.Spec.Version != "" {
		return fmt.Sprintf("%s:%s", config2.GrafanaImage, cr.Spec.Version)
	}
	grafanaImg := os.Getenv("RELATED_IMAGE_GRAFANA")
	if grafanaImg == "" {
		grafanaImg = fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)
	}
	return grafanaImg
}

func getContainers(cr *v1beta1.Grafana, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars, openshiftPlatform bool) []v1.Container {
	var containers []v1.Container

	image := getGrafanaImage(cr)
	plugins := model.GetPluginsConfigMap(cr, scheme)

	// env var to restart containers if plugins change
	t := true
	var envVars []v1.EnvVar
	envVars = append(envVars, v1.EnvVar{
		Name: "PLUGINS_HASH",
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: plugins.Name,
				},
				Key:      "PLUGINS_HASH",
				Optional: &t,
			},
		},
	})

	// env var to restart container if config changes
	envVars = append(envVars, v1.EnvVar{
		Name:  "CONFIG_HASH",
		Value: vars.ConfigHash,
	})

	// env var to restart container if plugins change
	envVars = append(envVars, v1.EnvVar{
		Name:  "GF_INSTALL_PLUGINS",
		Value: vars.Plugins,
	})

	// env var to set location where temporary files can be written (e.g. plugin downloads)
	envVars = append(envVars, v1.EnvVar{
		Name:  "TMPDIR",
		Value: config2.GrafanaDataPath,
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
		VolumeMounts:             getVolumeMounts(cr, scheme),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getGrafanaContainerSecurityContext(openshiftPlatform),
		ReadinessProbe:           getReadinessProbe(cr),
	})

	// Use auto generated admin account?
	secret := model.GetGrafanaAdminSecret(cr, scheme)

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
			AllowPrivilegeEscalation: model.BoolPtr(false),
			ReadOnlyRootFilesystem:   model.BoolPtr(true),
			Privileged:               model.BoolPtr(false),
			RunAsNonRoot:             model.BoolPtr(true),
			Capabilities:             capability,
		}
	}
	return &v1.SecurityContext{
		AllowPrivilegeEscalation: model.BoolPtr(false),
		ReadOnlyRootFilesystem:   model.BoolPtr(true),
		Privileged:               model.BoolPtr(false),
		RunAsNonRoot:             model.BoolPtr(true),
		RunAsUser:                model.IntPtr(10001),
		RunAsGroup:               model.IntPtr(10001),
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

func getDeploymentSpec(cr *v1beta1.Grafana, deploymentName string, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars, openshiftPlatform bool) v12.DeploymentSpec {
	sa := model.GetGrafanaServiceAccount(cr, scheme)

	return v12.DeploymentSpec{
		Selector: &v13.LabelSelector{
			MatchLabels: map[string]string{
				"app": cr.Name,
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: v13.ObjectMeta{
				Name: deploymentName,
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Spec: v1.PodSpec{
				Volumes:            getVolumes(cr, scheme),
				Containers:         getContainers(cr, scheme, vars, openshiftPlatform),
				SecurityContext:    getPodSecurityContext(),
				ServiceAccountName: sa.Name,
			},
		},
	}
}
