package model

import (
	"fmt"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func getSkipCreateAdminAccount(cr *v1alpha1.Grafana) bool {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SkipCreateAdminAccount != nil {
		return *cr.Spec.Deployment.SkipCreateAdminAccount
	}

	return false
}

func getInitResources(cr *v1alpha1.Grafana) v13.ResourceRequirements {
	if cr.Spec.InitResources != nil {
		return *cr.Spec.InitResources
	}
	return v13.ResourceRequirements{
		Requests: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(InitMemoryRequest),
			v13.ResourceCPU:    resource.MustParse(InitCpuRequest),
		},
		Limits: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(InitMemoryLimit),
			v13.ResourceCPU:    resource.MustParse(InitCpuLimit),
		},
	}
}

func getResources(cr *v1alpha1.Grafana) v13.ResourceRequirements {
	if cr.Spec.Resources != nil {
		return *cr.Spec.Resources
	}
	return v13.ResourceRequirements{
		Requests: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(MemoryRequest),
			v13.ResourceCPU:    resource.MustParse(CpuRequest),
		},
		Limits: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(MemoryLimit),
			v13.ResourceCPU:    resource.MustParse(CpuLimit),
		},
	}
}

func getAffinities(cr *v1alpha1.Grafana) *v13.Affinity {
	var affinity = v13.Affinity{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Affinity != nil {
		affinity = *cr.Spec.Deployment.Affinity
	}
	return &affinity
}

func getSecurityContext(cr *v1alpha1.Grafana) *v13.PodSecurityContext {
	var securityContext = v13.PodSecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SecurityContext != nil {
		securityContext = *cr.Spec.Deployment.SecurityContext
	}
	return &securityContext
}

func getContainerSecurityContext(cr *v1alpha1.Grafana) *v13.SecurityContext {
	var containerSecurityContext = v13.SecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.ContainerSecurityContext != nil {
		containerSecurityContext = *cr.Spec.Deployment.ContainerSecurityContext
	}
	return &containerSecurityContext
}

func getReplicas(cr *v1alpha1.Grafana) *int32 {
	var replicas int32 = 1
	if cr.Spec.Deployment == nil {
		return &replicas
	}
	if cr.Spec.Deployment.Replicas <= 0 {
		return &replicas
	} else {
		return &cr.Spec.Deployment.Replicas
	}
}

func getRollingUpdateStrategy() *v1.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromInt(25)
	var maxSurge intstr.IntOrString = intstr.FromInt(25)
	return &v1.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func getPodAnnotations(cr *v1alpha1.Grafana, existing map[string]string) map[string]string {
	var annotations = map[string]string{}
	// Add fixed annotations
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = fmt.Sprintf("%v", GetGrafanaPort(cr))
	annotations = MergeAnnotations(annotations, existing)

	if cr.Spec.Deployment != nil {
		annotations = MergeAnnotations(cr.Spec.Deployment.Annotations, annotations)
	}
	return annotations
}

func getPodLabels(cr *v1alpha1.Grafana) map[string]string {
	var labels = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	labels["app"] = GrafanaPodLabel
	return labels
}

func getNodeSelectors(cr *v1alpha1.Grafana) map[string]string {
	var nodeSelector = map[string]string{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.NodeSelector != nil {
		nodeSelector = cr.Spec.Deployment.NodeSelector
	}
	return nodeSelector

}

func getTerminationGracePeriod(cr *v1alpha1.Grafana) *int64 {
	var tcp int64 = 30
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.TerminationGracePeriodSeconds != 0 {
		tcp = cr.Spec.Deployment.TerminationGracePeriodSeconds
	}
	return &tcp

}

func getPodPriorityClassName(cr *v1alpha1.Grafana) string {
	if cr.Spec.Deployment != nil {
		return cr.Spec.Deployment.PriorityClassName
	}
	return ""
}

func getTolerations(cr *v1alpha1.Grafana) []v13.Toleration {
	tolerations := []v13.Toleration{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Tolerations != nil {
		for _, val := range cr.Spec.Deployment.Tolerations {
			tolerations = append(tolerations, val)
		}
	}
	return tolerations
}

func getVolumes(cr *v1alpha1.Grafana) []v13.Volume {
	var volumes []v13.Volume
	var volumeOptional bool = true

	// Volume to mount the config file from a config map
	volumes = append(volumes, v13.Volume{
		Name: GrafanaConfigName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: GrafanaConfigName,
				},
			},
		},
	})

	// Volume to store the logs
	volumes = append(volumes, v13.Volume{
		Name: GrafanaLogsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Data volume
	if cr.UsedPersistentVolume() {
		volumes = append(volumes, v13.Volume{
			Name: GrafanaDataVolumeName,
			VolumeSource: v13.VolumeSource{
				PersistentVolumeClaim: &v13.PersistentVolumeClaimVolumeSource{
					ClaimName: GrafanaDataStorageName,
				},
			},
		})
	} else {
		volumes = append(volumes, v13.Volume{
			Name: GrafanaDataVolumeName,
			VolumeSource: v13.VolumeSource{
				EmptyDir: &v13.EmptyDirVolumeSource{},
			},
		})
	}

	// Volume to store the plugins
	volumes = append(volumes, v13.Volume{
		Name: GrafanaPluginsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Volume to store the datasources
	volumes = append(volumes, v13.Volume{
		Name: GrafanaDatasourcesConfigMapName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: GrafanaDatasourcesConfigMapName,
				},
			},
		},
	})

	// Extra volumes for secrets
	for _, secret := range cr.Spec.Secrets {
		volumeName := fmt.Sprintf("secret-%s", secret)
		volumes = append(volumes, v13.Volume{
			Name: volumeName,
			VolumeSource: v13.VolumeSource{
				Secret: &v13.SecretVolumeSource{
					SecretName: secret,
					Optional:   &volumeOptional,
				},
			},
		})
	}

	// Extra volumes for config maps
	for _, configmap := range cr.Spec.ConfigMaps {
		volumeName := fmt.Sprintf("configmap-%s", configmap)
		volumes = append(volumes, v13.Volume{
			Name: volumeName,
			VolumeSource: v13.VolumeSource{
				ConfigMap: &v13.ConfigMapVolumeSource{
					LocalObjectReference: v13.LocalObjectReference{
						Name: configmap,
					},
				},
			},
		})
	}
	return volumes
}

func getEnvFrom(cr *v1alpha1.Grafana) []v13.EnvFromSource {
	var envFrom []v13.EnvFromSource
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.EnvFrom != nil {
		for _, v := range cr.Spec.Deployment.EnvFrom {
			envFrom = append(envFrom, *v.DeepCopy())
		}
	}
	return envFrom
}

// Don't add grafana specific volume mounts to extra containers and preserve
// pre existing ones
func getExtraContainerVolumeMounts(cr *v1alpha1.Grafana, mounts []v13.VolumeMount) []v13.VolumeMount {
	appendIfEmpty := func(mounts []v13.VolumeMount, mount v13.VolumeMount) []v13.VolumeMount {
		for _, existing := range mounts {
			if existing.Name == mount.Name || existing.MountPath == mount.MountPath {
				return mounts
			}
		}
		return append(mounts, mount)
	}

	for _, secret := range cr.Spec.Secrets {
		mountName := fmt.Sprintf("secret-%s", secret)
		mounts = appendIfEmpty(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.SecretsMountDir + secret,
		})
	}

	for _, configmap := range cr.Spec.ConfigMaps {
		mountName := fmt.Sprintf("configmap-%s", configmap)
		mounts = appendIfEmpty(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.ConfigMapsMountDir + configmap,
		})
	}

	return mounts
}

func getVolumeMounts(cr *v1alpha1.Grafana) []v13.VolumeMount {
	var mounts []v13.VolumeMount

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaConfigName,
		MountPath: "/etc/grafana/",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaDataVolumeName,
		MountPath: "/var/lib/grafana",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaPluginsVolumeName,
		MountPath: "/var/lib/grafana/plugins",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaLogsVolumeName,
		MountPath: "/var/log/grafana",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaDatasourcesConfigMapName,
		MountPath: "/etc/grafana/provisioning/datasources",
	})

	for _, secret := range cr.Spec.Secrets {
		mountName := fmt.Sprintf("secret-%s", secret)
		mounts = append(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.SecretsMountDir + secret,
		})
	}

	for _, configmap := range cr.Spec.ConfigMaps {
		mountName := fmt.Sprintf("configmap-%s", configmap)
		mounts = append(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.ConfigMapsMountDir + configmap,
		})
	}

	return mounts
}

func getProbe(cr *v1alpha1.Grafana, delay, timeout, failure int32) *v13.Probe {
	return &v13.Probe{
		Handler: v13.Handler{
			HTTPGet: &v13.HTTPGetAction{
				Path: GrafanaHealthEndpoint,
				Port: intstr.FromInt(GetGrafanaPort(cr)),
			},
		},
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		FailureThreshold:    failure,
	}
}

func getContainers(cr *v1alpha1.Grafana, configHash, dsHash string) []v13.Container {
	var containers []v13.Container
	var image string

	if cr.Spec.BaseImage != "" {
		image = cr.Spec.BaseImage
	} else {
		cfg := config.GetControllerConfig()
		img := cfg.GetConfigString(config.ConfigGrafanaImage, GrafanaImage)
		tag := cfg.GetConfigString(config.ConfigGrafanaImageTag, GrafanaVersion)
		image = fmt.Sprintf("%s:%s", img, tag)
	}

	containers = append(containers, v13.Container{
		Name:       "grafana",
		Image:      image,
		Args:       []string{"-config=/etc/grafana/grafana.ini"},
		WorkingDir: "",
		Ports: []v13.ContainerPort{
			{
				Name:          "grafana-http",
				ContainerPort: int32(GetGrafanaPort(cr)),
				Protocol:      "TCP",
			},
		},
		Env: []v13.EnvVar{
			{
				Name:  LastConfigEnvVar,
				Value: configHash,
			},
			{
				Name:  LastDatasourcesConfigEnvVar,
				Value: dsHash,
			},
		},
		EnvFrom:                  getEnvFrom(cr),
		Resources:                getResources(cr),
		VolumeMounts:             getVolumeMounts(cr),
		LivenessProbe:            getProbe(cr, 60, 30, 10),
		ReadinessProbe:           getProbe(cr, 5, 3, 1),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getContainerSecurityContext(cr),
	})

	// Use auto generated admin account?
	if getSkipCreateAdminAccount(cr) == false {
		for i := 0; i < len(containers); i++ {
			containers[i].Env = append(containers[i].Env, v13.EnvVar{
				Name: GrafanaAdminUserEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: GrafanaAdminSecretName,
						},
						Key: GrafanaAdminUserEnvVar,
					},
				},
			})
			containers[i].Env = append(containers[i].Env, v13.EnvVar{
				Name: GrafanaAdminPasswordEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: GrafanaAdminSecretName,
						},
						Key: GrafanaAdminPasswordEnvVar,
					},
				}})
		}
	}

	// Add extra containers
	for _, container := range cr.Spec.Containers {
		container.VolumeMounts = getExtraContainerVolumeMounts(cr, container.VolumeMounts)
		containers = append(containers, container)
	}

	return containers
}

func getInitContainers(cr *v1alpha1.Grafana, plugins string) []v13.Container {
	cfg := config.GetControllerConfig()
	image := cfg.GetConfigString(config.ConfigPluginsInitContainerImage, config.PluginsInitContainerImage)
	tag := cfg.GetConfigString(config.ConfigPluginsInitContainerTag, config.PluginsInitContainerTag)

	return []v13.Container{
		{
			Name:  GrafanaInitContainerName,
			Image: fmt.Sprintf("%s:%s", image, tag),
			Env: []v13.EnvVar{
				{
					Name:  "GRAFANA_PLUGINS",
					Value: plugins,
				},
			},
			Resources: getInitResources(cr),
			VolumeMounts: []v13.VolumeMount{
				{
					Name:      GrafanaPluginsVolumeName,
					ReadOnly:  false,
					MountPath: "/opt/plugins",
				},
			},
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: "File",
			ImagePullPolicy:          "IfNotPresent",
		},
	}
}

func getDeploymentSpec(cr *v1alpha1.Grafana, annotations map[string]string, configHash, plugins, dsHash string) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Replicas: getReplicas(cr),
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": GrafanaPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name:        GrafanaDeploymentName,
				Labels:      getPodLabels(cr),
				Annotations: getPodAnnotations(cr, annotations),
			},
			Spec: v13.PodSpec{
				NodeSelector:                  getNodeSelectors(cr),
				Tolerations:                   getTolerations(cr),
				Affinity:                      getAffinities(cr),
				SecurityContext:               getSecurityContext(cr),
				Volumes:                       getVolumes(cr),
				InitContainers:                getInitContainers(cr, plugins),
				Containers:                    getContainers(cr, configHash, dsHash),
				ServiceAccountName:            GrafanaServiceAccountName,
				TerminationGracePeriodSeconds: getTerminationGracePeriod(cr),
				PriorityClassName:             getPodPriorityClassName(cr),
			},
		},
		Strategy: v1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}
}

func GrafanaDeployment(cr *v1alpha1.Grafana, configHash, dsHash string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      GrafanaDeploymentName,
			Namespace: cr.Namespace,
		},
		Spec: getDeploymentSpec(cr, nil, configHash, "", dsHash),
	}
}

func GrafanaDeploymentSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaDeploymentName,
	}
}

func GrafanaDeploymentReconciled(cr *v1alpha1.Grafana, currentState *v1.Deployment, configHash, plugins, dshash string) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getDeploymentSpec(cr,
		currentState.Spec.Template.Annotations,
		configHash,
		plugins,
		dshash)
	return reconciled
}
