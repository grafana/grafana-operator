package model

import (
	"fmt"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	v13 "k8s.io/api/core/v1"
)

import (
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getExternal(cr *v1alpha1.Loki) *v1alpha1.LokiExternal {
	if cr.Spec.External != nil {
		return cr.Spec.External
	}
	return nil
}

func getLokiDeploymentSpec(cr *v1alpha1.Loki,configHash,dsHash string) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": LokiPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name: LokiDeploymentName,
			},
			//TODO ASK PETER ABOUT THIS
			Spec: v13.PodSpec{
				NodeSelector:                  getLokiNodeSelectors(cr),
				Tolerations:                   getLokiTolerations(cr),
				Affinity:                      getLokiAffinities(cr),
				SecurityContext:               getLokiSecurityContext(cr),
				Volumes:                       getLokiVolumes(cr),
				//TODO Probably not applicable |V|
				//InitContainers:                getLokiInitContainers(cr),
				Containers:                    getLokiContainers(cr, configHash, dsHash),
				ServiceAccountName:            LokiServiceAccountName,
				TerminationGracePeriodSeconds: getLokiTerminationGracePeriod(cr),
			},
		},
		Strategy: v1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}
}

func getLokiContainers(cr *v1alpha1.Loki, configHash string) []v13.Container {
	var containers []v13.Container
	var image string


	if cr.Spec.BaseImage != "" {
		image = cr.Spec.BaseImage
	} else {
		cfg := config.GetControllerConfig()
		img := cfg.GetConfigString(config.ConfigLokiImage, LokiImage)
		tag := cfg.GetConfigString(config.ConfigLokiImageTag, LokiVersion)
		image = fmt.Sprintf("%s:%s", img, tag)
	}


	containers = append(containers, v13.Container{
		Name:       "Loki",
		Image:      image,
		Args:       []string{"-config=/etc/grafana/grafana.ini"},
		WorkingDir: "",
		Ports: []v13.ContainerPort{
			{
				Name:          "loki-http",
				ContainerPort: int32(GetLokiPort(cr)),
				Protocol:      "TCP",
			},
		},
		Env: []v13.EnvVar{
			{
				Name:  LastConfigEnvVar,
				Value: configHash,
			},
		},
		VolumeMounts:             getLokiVolumeMounts(cr),
		LivenessProbe:            getLokiProbe(cr, 60, 30, 10),
		ReadinessProbe:           getLokiProbe(cr, 5, 3, 1),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		SecurityContext:          getLokiContainerSecurityContext(cr),
	})

	return containers
}





func getLokiVolumes(cr *v1alpha1.Loki) []v13.Volume {
	var volumes []v13.Volume

	volumes = append(volumes, v13.Volume{
		Name: LokiConfigName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: LokiConfigName,
				},
			},
		},
	})

	volumes = append(volumes, v13.Volume{
		Name: LokiLogsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Data volume
	if cr.UsedPersistentVolume() {
		volumes = append(volumes, v13.Volume{
			Name: LokiDataVolumeName,
			VolumeSource: v13.VolumeSource{
				PersistentVolumeClaim: &v13.PersistentVolumeClaimVolumeSource{
					ClaimName: LokiDataStorageName,
				},
			},
		})
	}

	return volumes

}

func getLokiSecurityContext(cr *v1alpha1.Loki) *v13.PodSecurityContext {
	securityContext := v13.PodSecurityContext{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SecurityContext != nil {
		securityContext = * cr.Spec.Deployment.SecurityContext
	}

	return &securityContext

}

func getLokiAffinities(cr *v1alpha1.Loki) *v13.Affinity {
	affinity := v13.Affinity{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Affinity != nil {
		affinity = *cr.Spec.Deployment.Affinity
	}
	return &affinity

}

func getLokiTolerations(cr *v1alpha1.Loki) []v13.Toleration {
	tolerations := []v13.Toleration{}

	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Tolerations != nil {
		for _, val := range cr.Spec.Deployment.Tolerations {
			tolerations = append(tolerations, val)
		}
	}
	return tolerations

}

func getLokiNodeSelectors(cr *v1alpha1.Loki) map[string]string {
	nodeSelector := map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.NodeSelector != nil {
		nodeSelector = cr.Spec.Deployment.NodeSelector
	}

	return nodeSelector
}

func LokiDeployment(cr *v1alpha1.Loki) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      LokiDeploymentName,
			Namespace: cr.Namespace,
		},
		Spec: getLokiDeploymentSpec(cr),
	}
}

func LokiDeploymentSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiDeploymentName,
	}
}

func LokiDeploymentReconciled(cr *v1alpha1.Loki, currentState *v1.Deployment) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getLokiDeploymentSpec(cr)
	return reconciled
}
