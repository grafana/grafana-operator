package grafana

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MemoryRequest                        = "256Mi"
	CPURequest                           = "100m"
	MemoryLimit                          = "1024Mi"
	GrafanaHealthEndpoint                = "/api/health"
	ReadinessProbeFailureThreshold int32 = 1
	ReadinessProbePeriodSeconds    int32 = 10
	ReadinessProbeSuccessThreshold int32 = 1
	ReadinessProbeTimeoutSeconds   int32 = 3
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

func (r *DeploymentReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("DeploymentReconciler")

	openshiftPlatform := r.isOpenShift
	log.Info("reconciling deployment", "openshift", openshiftPlatform)

	deployment := model.GetGrafanaDeployment(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		deployment.Spec = getDeploymentSpec(cr, deployment.Name, scheme, vars, openshiftPlatform)

		err := v1beta1.Merge(deployment, cr.Spec.Deployment)
		if err != nil {
			setInvalidMergeCondition(cr, "Deployment", err)
			return err
		}

		removeInvalidMergeCondition(cr, "Deployment")

		if scheme != nil {
			err = controllerutil.SetControllerReference(cr, deployment, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(deployment, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(MemoryRequest),
			corev1.ResourceCPU:    resource.MustParse(CPURequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(MemoryLimit),
		},
	}
}

func getVolumes(cr *v1beta1.Grafana, scheme *runtime.Scheme) []corev1.Volume {
	var volumes []corev1.Volume

	cm := model.GetGrafanaConfigMap(cr, scheme)

	// Volume to mount the config file from a config map
	volumes = append(volumes, corev1.Volume{
		Name: cm.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.Name,
				},
			},
		},
	})

	// Volume to store the logs
	volumes = append(volumes, corev1.Volume{
		Name: config.GrafanaLogsVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, corev1.Volume{
		Name: config.GrafanaDataVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return volumes
}

func getVolumeMounts(cr *v1beta1.Grafana, scheme *runtime.Scheme) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	cm := model.GetGrafanaConfigMap(cr, scheme)

	mounts = append(mounts, corev1.VolumeMount{
		Name:      cm.Name,
		MountPath: "/etc/grafana/grafana.ini",
		SubPath:   "grafana.ini",
	})

	mounts = append(mounts, corev1.VolumeMount{
		Name:      config.GrafanaDataVolumeName,
		MountPath: config.GrafanaDataPath,
	})

	mounts = append(mounts, corev1.VolumeMount{
		Name:      config.GrafanaLogsVolumeName,
		MountPath: config.GrafanaLogsPath,
	})

	return mounts
}

func getGrafanaImage(cr *v1beta1.Grafana) string {
	if cr.Spec.Version == "" {
		return fmt.Sprintf("%s:%s", config.GrafanaImage, config.GrafanaVersion)
	}

	if strings.ContainsAny(cr.Spec.Version, ":/@") {
		return cr.Spec.Version
	}

	return fmt.Sprintf("%s:%s", config.GrafanaImage, cr.Spec.Version)
}

func getContainers(cr *v1beta1.Grafana, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars, openshiftPlatform bool) []corev1.Container {
	var containers []corev1.Container

	image := getGrafanaImage(cr)
	plugins := model.GetPluginsConfigMap(cr, scheme)

	// env var to restart containers if plugins change
	t := true

	var envVars []corev1.EnvVar

	envVars = append(envVars, corev1.EnvVar{
		Name: "PLUGINS_HASH",
		ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: plugins.Name,
				},
				Key:      "PLUGINS_HASH",
				Optional: &t,
			},
		},
	})

	// env var to restart container if config changes
	envVars = append(envVars, corev1.EnvVar{
		Name:  "CONFIG_HASH",
		Value: vars.ConfigHash,
	})

	// env var to restart container if plugins change
	envVars = append(envVars, corev1.EnvVar{
		Name:  "GF_INSTALL_PLUGINS",
		Value: vars.Plugins,
	})

	// env var to set location where temporary files can be written (e.g. plugin downloads)
	envVars = append(envVars, corev1.EnvVar{
		Name:  "TMPDIR",
		Value: config.GrafanaDataPath,
	})

	// env var to get Pod IP from downward API for gossip (useful for unified alerting).
	envVars = append(envVars, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})

	containers = append(containers, corev1.Container{
		Name:       "grafana",
		Image:      image,
		Args:       []string{"-config=/etc/grafana/grafana.ini"},
		WorkingDir: "",
		Ports: []corev1.ContainerPort{
			{
				Name:          "grafana-http",
				ContainerPort: int32(GetGrafanaPort(cr)), // #nosec G115
				Protocol:      "TCP",
			},
			{
				Name:          config.GrafanaAlertPortName,
				ContainerPort: int32(config.GrafanaAlertPort),
				Protocol:      "TCP",
			},
		},
		Env:                      envVars,
		Resources:                getResources(),
		VolumeMounts:             getVolumeMounts(cr, scheme),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getDefaultContainerSecurityContext(cr.Spec.DisableDefaultSecurityContext, openshiftPlatform),
		ReadinessProbe:           getReadinessProbe(cr),
	})

	// Use auto generated admin account?
	secret := model.GetGrafanaAdminSecret(cr, scheme)

	for i := range containers {
		containers[i].Env = append(containers[i].Env, corev1.EnvVar{
			Name: config.GrafanaAdminUserEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: config.GrafanaAdminUserEnvVar,
				},
			},
		})
		containers[i].Env = append(containers[i].Env, corev1.EnvVar{
			Name: config.GrafanaAdminPasswordEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: config.GrafanaAdminPasswordEnvVar,
				},
			},
		})
	}

	return containers
}

// getDefaultContainerSecurityContext provides securityContext for grafana container unless disabled
func getDefaultContainerSecurityContext(disableSecurityContext string, openshiftPlatform bool) *corev1.SecurityContext {
	if disableSecurityContext == "Container" || disableSecurityContext == "All" {
		return nil
	}

	capability := &corev1.Capabilities{
		Drop: []corev1.Capability{
			"ALL",
		},
	}
	if openshiftPlatform {
		return &corev1.SecurityContext{
			AllowPrivilegeEscalation: model.BoolPtr(false),
			ReadOnlyRootFilesystem:   model.BoolPtr(true),
			Privileged:               model.BoolPtr(false),
			RunAsNonRoot:             model.BoolPtr(true),
			Capabilities:             capability,
		}
	}

	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: model.BoolPtr(false),
		ReadOnlyRootFilesystem:   model.BoolPtr(true),
		Privileged:               model.BoolPtr(false),
		RunAsNonRoot:             model.BoolPtr(true),
		RunAsUser:                model.IntPtr(10001),
		RunAsGroup:               model.IntPtr(10001),
		Capabilities:             capability,
	}
}

func getReadinessProbe(cr *v1beta1.Grafana) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   GrafanaHealthEndpoint,
				Port:   intstr.FromInt(GetGrafanaPort(cr)),
				Scheme: corev1.URISchemeHTTP,
			},
		},
		TimeoutSeconds:   ReadinessProbeTimeoutSeconds,
		PeriodSeconds:    ReadinessProbePeriodSeconds,
		SuccessThreshold: ReadinessProbeSuccessThreshold,
		FailureThreshold: ReadinessProbeFailureThreshold,
	}
}

// getDefaultPodSecurityContext provides securityContext for grafana pod unless disabled
func getDefaultPodSecurityContext(disableSecurityContext string) *corev1.PodSecurityContext {
	if disableSecurityContext == "Pod" || disableSecurityContext == "All" {
		return nil
	}

	return &corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
}

func getDeploymentSpec(cr *v1beta1.Grafana, deploymentName string, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars, openshiftPlatform bool) appsv1.DeploymentSpec {
	sa := model.GetGrafanaServiceAccount(cr, scheme)

	return appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": cr.Name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				Labels: map[string]string{
					"app": cr.Name,
				},
			},
			Spec: corev1.PodSpec{
				Volumes:            getVolumes(cr, scheme),
				Containers:         getContainers(cr, scheme, vars, openshiftPlatform),
				SecurityContext:    getDefaultPodSecurityContext(cr.Spec.DisableDefaultSecurityContext),
				ServiceAccountName: sa.Name,
			},
		},
	}
}
