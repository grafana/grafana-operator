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
	MemoryRequest = "256Mi"
	CpuRequest    = "100m"
	MemoryLimit   = "1024Mi"
	CpuLimit      = "500m"
)

func getProxyResources(cr *v1alpha1.GrafanaProxy) v13.ResourceRequirements {
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

func getProxyReplicas(cr *v1alpha1.GrafanaProxy) *int32 {
	var replicas int32 = 1
	return &replicas
}

func getProxyRollingUpdateStrategy() *v1.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromInt(25)
	var maxSurge intstr.IntOrString = intstr.FromInt(25)
	return &v1.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func getProxyPodAnnotations(cr *v1alpha1.GrafanaProxy) map[string]string {
	var annotations = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Annotations != nil {
		annotations = cr.Spec.Deployment.Annotations
	}

	// Add fixed annotations
	annotations["vice-president"] = "true"
	//annotations["prometheus.io/scrape"] = "true"
	//annotations["prometheus.io/port"] = fmt.Sprintf("%v", 80)
	return annotations
}

func getProxyPodLabels(cr *v1alpha1.GrafanaProxy) map[string]string {
	var labels = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	labels["app"] = GrafanaProxyPodLabel
	return labels
}

func getProxyVolumes(cr *v1alpha1.GrafanaProxy) []v13.Volume {
	var volumes []v13.Volume

	// Volume to mount the config file from a config map
	volumes = append(volumes, v13.Volume{
		Name: GrafanaProxyConfigName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: GrafanaProxyConfigName,
				},
			},
		},
	})

	return volumes
}

func getProxyVolumeMounts(cr *v1alpha1.GrafanaProxy) []v13.VolumeMount {
	var mounts []v13.VolumeMount

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaProxyConfigName,
		MountPath: "/etc/dex/cfg",
	})

	return mounts
}

func getProxyProbe(cr *v1alpha1.GrafanaProxy, delay, timeout, failure int32) *v13.Probe {
	return &v13.Probe{
		Handler: v13.Handler{
			HTTPGet: &v13.HTTPGetAction{
				Path: GrafanaProxyHealthEndpoint,
				Port: intstr.FromInt(80),
			},
		},
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		FailureThreshold:    failure,
	}
}

func getProxyContainers(cr *v1alpha1.GrafanaProxy, configHash string) []v13.Container {
	var containers []v13.Container

	cfg := config.GetControllerConfig()
	image := cfg.GetConfigString(config.ConfigGrafanaProxyImage, GrafanaProxyImage)
	tag := cfg.GetConfigString(config.ConfigGrafanaProxyImageTag, GrafanaProxyVersion)

	containers = append(containers, v13.Container{
		Name:       "grafana-proxy",
		Image:      fmt.Sprintf("%s:%s", image, tag),
		Command:    []string{"/usr/local/bin/dex", "serve", "/etc/dex/cfg/config.yaml"},
		WorkingDir: "",
		Ports: []v13.ContainerPort{
			{
				Name:          "http",
				ContainerPort: int32(80),
				Protocol:      "TCP",
			},
		},
		Env: []v13.EnvVar{
			{
				Name:  LastConfigEnvVar,
				Value: configHash,
			},
		},
		Resources:                getProxyResources(cr),
		VolumeMounts:             getProxyVolumeMounts(cr),
		LivenessProbe:            getProxyProbe(cr, 3, 2, 10),
		ReadinessProbe:           getProxyProbe(cr, 3, 2, 10),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
	})

	return containers
}

func getProxyDeploymentSpec(cr *v1alpha1.GrafanaProxy, configHash string) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Replicas: getProxyReplicas(cr),
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": GrafanaProxyPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name:        GrafanaProxyDeploymentName,
				Labels:      getProxyPodLabels(cr),
				Annotations: getProxyPodAnnotations(cr),
			},
			Spec: v13.PodSpec{
				Volumes:            getProxyVolumes(cr),
				Containers:         getProxyContainers(cr, configHash),
				ServiceAccountName: GrafanaProxyServiceAccountName,
			},
		},
		Strategy: v1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}
}

func GrafanaProxyDeployment(cr *v1alpha1.GrafanaProxy, configHash string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      GrafanaProxyDeploymentName,
			Namespace: cr.Namespace,
		},
		Spec: getProxyDeploymentSpec(cr, configHash),
	}
}

func GrafanaProxyDeploymentSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyDeploymentName,
	}
}

func GrafanaProxyDeploymentReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1.Deployment, configHash string) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getProxyDeploymentSpec(cr, configHash)
	return reconciled
}
