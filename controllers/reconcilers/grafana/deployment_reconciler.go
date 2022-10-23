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
		deployment.Spec = getDeploymentSpec(cr, deployment.Name, scheme, vars)
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
			v1.ResourceCPU:    resource.MustParse(CpuLimit),
		},
	}
}

func getVolumes(cr *v1beta1.Grafana, scheme *runtime.Scheme) []v1.Volume { // nolint
	var volumes []v1.Volume // nolint

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

	return mounts
}

func getContainers(cr *v1beta1.Grafana, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars) []v1.Container { // nolint
	var containers []v1.Container // nolint

	image := fmt.Sprintf("%s:%s", config2.GrafanaImage, config2.GrafanaVersion)
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

func getDeploymentSpec(cr *v1beta1.Grafana, deploymentName string, scheme *runtime.Scheme, vars *v1beta1.OperatorReconcileVars) v12.DeploymentSpec {
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
				Containers:         getContainers(cr, scheme, vars),
				ServiceAccountName: sa.Name,
			},
		},
	}
}
