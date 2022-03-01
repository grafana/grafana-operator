package grafana

import (
	"context"
	"fmt"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	config2 "github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
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
	InitMemoryRequest = "128Mi"
	InitCpuRequest    = "250m"
	InitMemoryLimit   = "512Mi"
	InitCpuLimit      = "1000m"
	MemoryRequest     = "256Mi"
	CpuRequest        = "100m"
	MemoryLimit       = "1024Mi"
	CpuLimit          = "500m"
)

type DeploymentReconciler struct {
	client client.Client
}

func NewDeploymentReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &DeploymentReconciler{
		client: client,
	}
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	_ = log.FromContext(ctx)

	deployment := model.GetGrafanaDeployment(cr, scheme)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deployment, func() error {
		deployment.Labels = getDeploymentLabels(cr)
		deployment.Spec = getDeploymentSpec(cr, deployment.Annotations, deployment.Name, scheme, vars)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getReplicas(cr *v1beta1.Grafana) *int32 {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Replicas != nil {
		return cr.Spec.Deployment.Replicas
	}

	return nil
}

func getTerminationGracePeriod(cr *v1beta1.Grafana) *int64 {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.TerminationGracePeriodSeconds != nil {
		return cr.Spec.Deployment.TerminationGracePeriodSeconds
	}
	return nil
}

func getInitResources(cr *v1beta1.Grafana) v1.ResourceRequirements {
	if cr.Spec.InitResources != nil {
		return *cr.Spec.InitResources
	}
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(InitMemoryRequest),
			v1.ResourceCPU:    resource.MustParse(InitCpuRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse(InitMemoryLimit),
			v1.ResourceCPU:    resource.MustParse(InitCpuLimit),
		},
	}
}

func getResources(cr *v1beta1.Grafana) v1.ResourceRequirements {
	if cr.Spec.Resources != nil {
		return *cr.Spec.Resources
	}
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

func getAffinities(cr *v1beta1.Grafana) *v1.Affinity {
	var affinity = v1.Affinity{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Affinity != nil {
		affinity = *cr.Spec.Deployment.Affinity
	}
	return &affinity
}

func getSecurityContext(cr *v1beta1.Grafana) *v1.PodSecurityContext {
	var securityContext = v1.PodSecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SecurityContext != nil {
		securityContext = *cr.Spec.Deployment.SecurityContext
	}
	return &securityContext
}

func getContainerSecurityContext(cr *v1beta1.Grafana) *v1.SecurityContext {
	var containerSecurityContext = v1.SecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.ContainerSecurityContext != nil {
		containerSecurityContext = *cr.Spec.Deployment.ContainerSecurityContext
	}
	return &containerSecurityContext
}

func getRollingUpdateStrategy() *v12.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromInt(25)
	var maxSurge intstr.IntOrString = intstr.FromInt(25)
	return &v12.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func getDeploymentStrategy(cr *v1beta1.Grafana) v12.DeploymentStrategy {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Strategy != nil {
		return *cr.Spec.Deployment.Strategy
	}

	return v12.DeploymentStrategy{
		Type:          "RollingUpdate",
		RollingUpdate: getRollingUpdateStrategy(),
	}
}

func getDeploymentLabels(cr *v1beta1.Grafana) map[string]string {
	var labels = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	return labels
}

func getPodLabels(cr *v1beta1.Grafana) map[string]string {
	var labels = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	labels["app"] = cr.Name
	return labels
}

func getNodeSelectors(cr *v1beta1.Grafana) map[string]string {
	var nodeSelector = map[string]string{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.NodeSelector != nil {
		nodeSelector = cr.Spec.Deployment.NodeSelector
	}
	return nodeSelector
}

func getPodPriorityClassName(cr *v1beta1.Grafana) string {
	if cr.Spec.Deployment != nil {
		return cr.Spec.Deployment.PriorityClassName
	}
	return ""
}

func getTolerations(cr *v1beta1.Grafana) []v1.Toleration {
	tolerations := []v1.Toleration{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Tolerations != nil {
		tolerations = append(tolerations, cr.Spec.Deployment.Tolerations...)
	}
	return tolerations
}

func getVolumes(cr *v1beta1.Grafana, scheme *runtime.Scheme) []v1.Volume { // nolint
	var volumes []v1.Volume // nolint
	var volumeOptional = true

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

	// Data volume
	if cr.UsePersistentVolume() {
		pvc := model.GetGrafanaDataPVC(cr, scheme)

		volumes = append(volumes, v1.Volume{
			Name: config2.GrafanaDataVolumeName,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		})
	} else {
		volumes = append(volumes, v1.Volume{
			Name: config2.GrafanaDataVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}

	// Extra volumes for secrets
	for _, secret := range cr.Spec.Secrets {
		volumeName := fmt.Sprintf("secret-%s", secret)
		volumes = append(volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
					Optional:   &volumeOptional,
				},
			},
		})
	}

	// Extra volumes for config maps
	for _, configmap := range cr.Spec.ConfigMaps {
		volumeName := fmt.Sprintf("configmap-%s", configmap)
		volumes = append(volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configmap,
					},
				},
			},
		})
	}
	return volumes
}

func getEnvFrom(cr *v1beta1.Grafana) []v1.EnvFromSource {
	var envFrom []v1.EnvFromSource
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.EnvFrom != nil {
		for _, v := range cr.Spec.Deployment.EnvFrom {
			envFrom = append(envFrom, *v.DeepCopy())
		}
	}
	return envFrom
}

func getVolumeMounts(cr *v1beta1.Grafana, scheme *runtime.Scheme) []v1.VolumeMount {
	var mounts []v1.VolumeMount // nolint

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

	for _, secret := range cr.Spec.Secrets {
		mountName := fmt.Sprintf("secret-%s", secret)
		mounts = append(mounts, v1.VolumeMount{
			Name:      mountName,
			MountPath: config2.SecretsMountDir + secret,
		})
	}

	for _, configmap := range cr.Spec.ConfigMaps {
		mountName := fmt.Sprintf("configmap-%s", configmap)
		mounts = append(mounts, v1.VolumeMount{
			Name:      mountName,
			MountPath: config2.ConfigMapsMountDir + configmap,
		})
	}

	return mounts
}

func getContainers(cr *v1beta1.Grafana, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars) []v1.Container { // nolint
	var containers []v1.Container // nolint
	var image string

	if cr.Spec.BaseImage != "" {
		image = cr.Spec.BaseImage
	} else {
		image = fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)
	}

	plugins := model.GetPluginsConfigMap(cr, scheme)

	// env var to restart containers if plugins change
	var t bool = true
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
		EnvFrom:                  getEnvFrom(cr),
		Resources:                getResources(cr),
		VolumeMounts:             getVolumeMounts(cr, scheme),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getContainerSecurityContext(cr),
	})

	// Use auto generated admin account?
	if !cr.SkipCreateAdminAccount() {
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
				}})
		}
	}

	return containers
}

func getDeploymentSpec(cr *v1beta1.Grafana, annotations map[string]string, deploymentName string, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars) v12.DeploymentSpec {
	sa := model.GetGrafanaServiceAccount(cr, scheme)

	return v12.DeploymentSpec{
		Replicas: getReplicas(cr),
		Selector: &v13.LabelSelector{
			MatchLabels: map[string]string{
				"app": cr.Name,
			},
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: v13.ObjectMeta{
				Name:   deploymentName,
				Labels: getPodLabels(cr),
			},
			Spec: v1.PodSpec{
				NodeSelector:                  getNodeSelectors(cr),
				Tolerations:                   getTolerations(cr),
				Affinity:                      getAffinities(cr),
				SecurityContext:               getSecurityContext(cr),
				Volumes:                       getVolumes(cr, scheme),
				Containers:                    getContainers(cr, scheme, vars),
				ServiceAccountName:            sa.Name,
				TerminationGracePeriodSeconds: getTerminationGracePeriod(cr),
				PriorityClassName:             getPodPriorityClassName(cr),
			},
		},
		Strategy: getDeploymentStrategy(cr),
	}
}
