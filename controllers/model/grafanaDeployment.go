package model

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
)

const (
	InitMemoryRequest                       = "128Mi"
	InitCpuRequest                          = "250m"
	InitMemoryLimit                         = "512Mi"
	InitCpuLimit                            = "1000m"
	MemoryRequest                           = "256Mi"
	CpuRequest                              = "100m"
	MemoryLimit                             = "1024Mi"
	CpuLimit                                = "500m"
	LivenessProbeFailureThreshold     int32 = 10
	LivenessProbeInitialDelaySeconds  int32 = 60
	LivenessProbePeriodSeconds        int32 = 10
	LivenessProbeSuccessThreshold     int32 = 1
	LivenessProbeTimeoutSeconds       int32 = 30
	ReadinessProbeFailureThreshold    int32 = 1
	ReadinessProbeInitialDelaySeconds int32 = 5
	ReadinessProbePeriodSeconds       int32 = 10
	ReadinessProbeSuccessThreshold    int32 = 1
	ReadinessProbeTimeoutSeconds      int32 = 3
)

func getSkipCreateAdminAccount(cr *v1alpha1.Grafana) bool {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SkipCreateAdminAccount != nil {
		return *cr.Spec.Deployment.SkipCreateAdminAccount
	}

	return false
}

func getReplicas(cr *v1alpha1.Grafana) *int32 {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Replicas != nil {
		return cr.Spec.Deployment.Replicas
	}

	return nil
}

func getTerminationGracePeriod(cr *v1alpha1.Grafana) *int64 {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.TerminationGracePeriodSeconds != nil {
		return cr.Spec.Deployment.TerminationGracePeriodSeconds
	}

	return nil
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
	affinity := v13.Affinity{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Affinity != nil {
		affinity = *cr.Spec.Deployment.Affinity
	}
	return &affinity
}

func getSecurityContext(cr *v1alpha1.Grafana) *v13.PodSecurityContext {
	securityContext := v13.PodSecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SecurityContext != nil {
		securityContext = *cr.Spec.Deployment.SecurityContext
	}
	return &securityContext
}

func getContainerSecurityContext(cr *v1alpha1.Grafana) *v13.SecurityContext {
	containerSecurityContext := v13.SecurityContext{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.ContainerSecurityContext != nil {
		containerSecurityContext = *cr.Spec.Deployment.ContainerSecurityContext
	}
	return &containerSecurityContext
}

func getDeploymentStrategy(cr *v1alpha1.Grafana) v1.DeploymentStrategy {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Strategy != nil {
		return *cr.Spec.Deployment.Strategy
	}

	return v1.DeploymentStrategy{
		Type:          "RollingUpdate",
		RollingUpdate: getRollingUpdateStrategy(),
	}
}

func getDeploymentLabels(cr *v1alpha1.Grafana) map[string]string {
	labels := map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	return labels
}

func getDeploymentAnnotations(cr *v1alpha1.Grafana, existing map[string]string) map[string]string {
	annotations := map[string]string{}
	// Add fixed annotations
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = fmt.Sprintf("%v", GetGrafanaPort(cr))
	annotations = MergeAnnotations(annotations, existing)

	if cr.Spec.Deployment != nil {
		annotations = MergeAnnotations(cr.Spec.Deployment.Annotations, annotations)
	}
	return annotations
}

func getRollingUpdateStrategy() *v1.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromString("25%")
	var maxSurge intstr.IntOrString = intstr.FromString("25%")
	return &v1.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func getPodAnnotations(cr *v1alpha1.Grafana, existing map[string]string) map[string]string {
	annotations := map[string]string{}
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
	labels := map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	labels["app"] = constants.GrafanaPodLabel
	return labels
}

func getNodeSelectors(cr *v1alpha1.Grafana) map[string]string {
	nodeSelector := map[string]string{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.NodeSelector != nil {
		nodeSelector = cr.Spec.Deployment.NodeSelector
	}
	return nodeSelector
}

func getPodPriorityClassName(cr *v1alpha1.Grafana) string {
	if cr.Spec.Deployment != nil {
		return cr.Spec.Deployment.PriorityClassName
	}
	return ""
}

func getTopologySpreadConstraints(cr *v1alpha1.Grafana) []v13.TopologySpreadConstraint {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.TopologySpreadConstraints != nil {
		return cr.Spec.Deployment.TopologySpreadConstraints
	}
	return []v13.TopologySpreadConstraint{}
}

func getTolerations(cr *v1alpha1.Grafana) []v13.Toleration {
	tolerations := []v13.Toleration{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Tolerations != nil {
		tolerations = append(tolerations, cr.Spec.Deployment.Tolerations...)
	}
	return tolerations
}

func getVolumes(cr *v1alpha1.Grafana) []v13.Volume { // nolint
	var volumes []v13.Volume // nolint
	volumeOptional := true

	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaProvisionPluginVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaProvisionDashboardVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaProvisionNotifierVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Volume to mount the config file from a config map
	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaConfigName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: constants.GrafanaConfigName,
				},
			},
		},
	})

	// Volume to store the logs
	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaLogsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Data volume
	if cr.UsedPersistentVolume() {
		volumes = append(volumes, v13.Volume{
			Name: constants.GrafanaDataVolumeName,
			VolumeSource: v13.VolumeSource{
				PersistentVolumeClaim: &v13.PersistentVolumeClaimVolumeSource{
					ClaimName: constants.GrafanaDataStorageName,
				},
			},
		})
	} else {
		volumes = append(volumes, v13.Volume{
			Name: constants.GrafanaDataVolumeName,
			VolumeSource: v13.VolumeSource{
				EmptyDir: &v13.EmptyDirVolumeSource{},
			},
		})
	}

	// Volume to store the plugins
	appendIfContainsPlugin := func() bool {
		var foundGrafanaPluginsPath bool
		if cr.Spec.Deployment != nil {
			for _, item := range cr.Spec.Deployment.ExtraVolumeMounts {
				if item.MountPath == config.GrafanaPluginsPath {
					foundGrafanaPluginsPath = true
					break
				}
			}
		}

		if cr.Spec.Deployment != nil {
			volumes = append(volumes, cr.Spec.Deployment.ExtraVolumes...)
		}
		return foundGrafanaPluginsPath
	}
	if !appendIfContainsPlugin() {
		volumes = append(volumes, v13.Volume{
			Name: constants.GrafanaPluginsVolumeName,
			VolumeSource: v13.VolumeSource{
				EmptyDir: &v13.EmptyDirVolumeSource{},
			},
		})
	}

	// Volume to store the datasources
	volumes = append(volumes, v13.Volume{
		Name: constants.GrafanaDatasourcesConfigMapName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: constants.GrafanaDatasourcesConfigMapName,
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
	var mounts []v13.VolumeMount // nolint

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaConfigName,
		MountPath: "/etc/grafana/",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaDataVolumeName,
		MountPath: config.GrafanaDataPath,
	})

	appendIfContainsPlugin := func() bool {
		var foundGrafanaPluginsPath bool
		if cr.Spec.Deployment != nil {
			for _, item := range cr.Spec.Deployment.ExtraVolumeMounts {
				if item.MountPath == config.GrafanaPluginsPath {
					foundGrafanaPluginsPath = true
					break
				}
			}
		}

		if cr.Spec.Deployment != nil {
			mounts = append(mounts, cr.Spec.Deployment.ExtraVolumeMounts...)
		}
		return foundGrafanaPluginsPath
	}
	if !appendIfContainsPlugin() {
		mounts = append(mounts, v13.VolumeMount{
			Name:      constants.GrafanaPluginsVolumeName,
			MountPath: config.GrafanaPluginsPath,
		})
	}

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaProvisionPluginVolumeName,
		MountPath: config.GrafanaProvisioningPluginsPath,
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaProvisionDashboardVolumeName,
		MountPath: config.GrafanaProvisioningDashboardsPath,
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaProvisionNotifierVolumeName,
		MountPath: config.GrafanaProvisioningNotifiersPath,
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaLogsVolumeName,
		MountPath: config.GrafanaLogsPath,
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      constants.GrafanaDatasourcesConfigMapName,
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

func getLivenessProbe(cr *v1alpha1.Grafana) *v13.Probe {
	spec := &v1alpha1.LivenessProbeSpec{}
	if cr.Spec.LivenessProbeSpec != nil {
		spec = cr.Spec.LivenessProbeSpec
	}

	return &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			HTTPGet: &v13.HTTPGetAction{
				Path:   constants.GrafanaHealthEndpoint,
				Port:   intstr.FromInt(GetGrafanaPort(cr)),
				Scheme: cr.GetScheme(),
			},
		},
		InitialDelaySeconds: getDefaultInt32(spec.InitialDelaySeconds, LivenessProbeInitialDelaySeconds),
		TimeoutSeconds:      getDefaultInt32(spec.TimeOutSeconds, LivenessProbeTimeoutSeconds),
		PeriodSeconds:       getDefaultInt32(spec.PeriodSeconds, LivenessProbePeriodSeconds),
		SuccessThreshold:    getDefaultInt32(spec.SuccessThreshold, LivenessProbeSuccessThreshold),
		FailureThreshold:    getDefaultInt32(spec.FailureThreshold, LivenessProbeFailureThreshold),
	}
}

func getReadinessProbe(cr *v1alpha1.Grafana) *v13.Probe {
	spec := &v1alpha1.ReadinessProbeSpec{}
	if cr.Spec.ReadinessProbeSpec != nil {
		spec = cr.Spec.ReadinessProbeSpec
	}

	return &v13.Probe{
		ProbeHandler: v13.ProbeHandler{
			HTTPGet: &v13.HTTPGetAction{
				Path:   constants.GrafanaHealthEndpoint,
				Port:   intstr.FromInt(GetGrafanaPort(cr)),
				Scheme: cr.GetScheme(),
			},
		},
		InitialDelaySeconds: getDefaultInt32(spec.InitialDelaySeconds, ReadinessProbeInitialDelaySeconds),
		TimeoutSeconds:      getDefaultInt32(spec.TimeOutSeconds, ReadinessProbeTimeoutSeconds),
		PeriodSeconds:       getDefaultInt32(spec.PeriodSeconds, ReadinessProbePeriodSeconds),
		SuccessThreshold:    getDefaultInt32(spec.SuccessThreshold, ReadinessProbeSuccessThreshold),
		FailureThreshold:    getDefaultInt32(spec.FailureThreshold, ReadinessProbeFailureThreshold),
	}
}

func getContainers(cr *v1alpha1.Grafana, configHash, dsHash, credentialsHash string) []v13.Container { // nolint
	var containers []v13.Container // nolint
	var image string

	if cr.Spec.BaseImage != "" {
		image = cr.Spec.BaseImage
	} else {
		cfg := config.GetControllerConfig()
		img := cfg.GetConfigString(config.ConfigGrafanaImage, constants.GrafanaImage)
		tag := cfg.GetConfigString(config.ConfigGrafanaImageTag, constants.GrafanaVersion)
		image = fmt.Sprintf("%s:%s", img, tag)
	}

	envVars := []v13.EnvVar{
		{
			Name:  constants.LastConfigEnvVar,
			Value: configHash,
		},
		{
			Name:  constants.LastDatasourcesConfigEnvVar,
			Value: dsHash,
		},
		{
			Name:  constants.LastCredentialsEnvVar,
			Value: credentialsHash,
		},
	}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.HttpProxy != nil && cr.Spec.Deployment.HttpProxy.Enabled {
		envVars = append(envVars, v13.EnvVar{
			Name:  "HTTP_PROXY",
			Value: cr.Spec.Deployment.HttpProxy.URL,
		})
		if cr.Spec.Deployment.HttpProxy.SecureURL != "" {
			envVars = append(envVars, v13.EnvVar{
				Name:  "HTTPS_PROXY",
				Value: cr.Spec.Deployment.HttpProxy.SecureURL,
			})
		}
		if cr.Spec.Deployment.HttpProxy.NoProxy != "" {
			envVars = append(envVars, v13.EnvVar{
				Name:  "NO_PROXY",
				Value: cr.Spec.Deployment.HttpProxy.NoProxy,
			})
		}
	}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Env != nil {
		envVars = append(envVars, cr.Spec.Deployment.Env...)
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
		Env:                      envVars,
		EnvFrom:                  getEnvFrom(cr),
		Resources:                getResources(cr),
		VolumeMounts:             getVolumeMounts(cr),
		LivenessProbe:            getLivenessProbe(cr),
		ReadinessProbe:           getReadinessProbe(cr),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getContainerSecurityContext(cr),
	})

	// Use auto generated admin account?
	if !getSkipCreateAdminAccount(cr) {
		for i := 0; i < len(containers); i++ {
			containers[i].Env = append(containers[i].Env, v13.EnvVar{
				Name: constants.GrafanaAdminUserEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: constants.GrafanaAdminSecretName,
						},
						Key: constants.GrafanaAdminUserEnvVar,
					},
				},
			})
			containers[i].Env = append(containers[i].Env, v13.EnvVar{
				Name: constants.GrafanaAdminPasswordEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: constants.GrafanaAdminSecretName,
						},
						Key: constants.GrafanaAdminPasswordEnvVar,
					},
				},
			})
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
	var image string

	if cr.Spec.InitImage != "" {
		image = cr.Spec.InitImage
	} else {
		cfg := config.GetControllerConfig()
		img := cfg.GetConfigString(config.ConfigPluginsInitContainerImage, config.PluginsInitContainerImage)
		tag := cfg.GetConfigString(config.ConfigPluginsInitContainerTag, config.PluginsInitContainerTag)
		image = fmt.Sprintf("%s:%s", img, tag)
	}

	envVars := []v13.EnvVar{
		{
			Name:  "GRAFANA_PLUGINS",
			Value: plugins,
		},
	}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.HttpProxy != nil && cr.Spec.Deployment.HttpProxy.Enabled {
		envVars = append(envVars, v13.EnvVar{
			Name:  "HTTP_PROXY",
			Value: cr.Spec.Deployment.HttpProxy.URL,
		})
		if cr.Spec.Deployment.HttpProxy.SecureURL != "" {
			envVars = append(envVars, v13.EnvVar{
				Name:  "HTTPS_PROXY",
				Value: cr.Spec.Deployment.HttpProxy.SecureURL,
			})
		}
	}

	volumeName := constants.GrafanaPluginsVolumeName

	if cr.Spec.Deployment != nil {
		for _, item := range cr.Spec.Deployment.ExtraVolumeMounts {
			if item.MountPath == config.GrafanaPluginsPath {
				volumeName = item.Name
			}
		}
	}

	return []v13.Container{
		{
			Name:      constants.GrafanaInitContainerName,
			Image:     image,
			Env:       envVars,
			Resources: getInitResources(cr),
			VolumeMounts: []v13.VolumeMount{
				{
					Name:      volumeName,
					ReadOnly:  false,
					MountPath: "/opt/plugins",
				},
			},
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: "File",
			ImagePullPolicy:          "IfNotPresent",
			SecurityContext:          getContainerSecurityContext(cr),
		},
	}
}

func getDeploymentSpec(cr *v1alpha1.Grafana, annotations map[string]string, configHash, plugins, dsHash, credentialsHash string) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Replicas: getReplicas(cr),
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": constants.GrafanaPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name:        constants.GrafanaDeploymentName,
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
				Containers:                    getContainers(cr, configHash, dsHash, credentialsHash),
				ServiceAccountName:            constants.GrafanaServiceAccountName,
				TerminationGracePeriodSeconds: getTerminationGracePeriod(cr),
				PriorityClassName:             getPodPriorityClassName(cr),
				TopologySpreadConstraints:     getTopologySpreadConstraints(cr),
			},
		},
		Strategy: getDeploymentStrategy(cr),
	}
}

func GrafanaDeployment(cr *v1alpha1.Grafana, configHash, dsHash, credentialsHash string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:        constants.GrafanaDeploymentName,
			Namespace:   cr.Namespace,
			Labels:      getDeploymentLabels(cr),
			Annotations: getDeploymentAnnotations(cr, nil),
		},
		Spec: getDeploymentSpec(cr, nil, configHash, "", dsHash, credentialsHash),
	}
}

func GrafanaDeploymentSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      constants.GrafanaDeploymentName,
	}
}

func GrafanaDeploymentReconciled(cr *v1alpha1.Grafana, currentState *v1.Deployment, configHash, plugins, dshash, credentialsHash string) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getDeploymentSpec(cr,
		currentState.Spec.Template.Annotations,
		configHash,
		plugins,
		dshash,
		credentialsHash)
	return reconciled
}
