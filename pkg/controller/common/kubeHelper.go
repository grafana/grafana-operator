package common

import (
	stdErrors "errors"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"time"
)

var log = logf.Log.WithName("kube_helper")

type KubeHelperImpl struct {
	k8client *kubernetes.Clientset
	config   *ControllerConfig
}

func NewKubeHelper() *KubeHelperImpl {
	config := config.GetConfigOrDie()

	k8client := kubernetes.NewForConfigOrDie(config)

	helper := new(KubeHelperImpl)
	helper.k8client = k8client
	helper.config = GetControllerConfig()
	return helper
}

func (h KubeHelperImpl) getConfigMapKey(namespace, name string) string {
	return fmt.Sprintf("%s_%s", namespace, strings.ToLower(name))
}

func (h KubeHelperImpl) getConfigMap(name string) (*v1.ConfigMap, error) {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	return h.k8client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (h KubeHelperImpl) updateConfigMap(c *v1.ConfigMap) error {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	_, err := h.k8client.CoreV1().ConfigMaps(namespace).Update(c)
	return err
}

func (h KubeHelperImpl) getGrafanaDeployment() (*apps.Deployment, error) {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	return h.k8client.AppsV1().Deployments(namespace).Get(GrafanaDeploymentName, metav1.GetOptions{})
}

func (h KubeHelperImpl) UpdateGrafanaConfig(config string, cr *v1alpha1.Grafana) error {
	configMap, err := h.getConfigMap(GrafanaConfigMapName)
	if err != nil {
		return err
	}
	configMap.Data[GrafanaConfigFileName] = config
	return h.updateConfigMap(configMap)
}

func (h KubeHelperImpl) UpdateDashboard(d *v1alpha1.GrafanaDashboard) (bool, error) {
	configMap, err := h.getConfigMap(GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	dashboardName := h.getConfigMapKey(d.Namespace, d.Spec.Name)
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[dashboardName] = d.Spec.Json
	err = h.updateConfigMap(configMap)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (h KubeHelperImpl) isKnown(config, namespace, name string) (bool, error) {
	configMap, err := h.getConfigMap(config)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if configMap.Data == nil {
		return false, nil
	}

	key := h.getConfigMapKey(namespace, name)
	_, found := configMap.Data[key]
	return found, nil
}

func (h KubeHelperImpl) IsKnown(kind string, o runtime.Object) (bool, error) {
	switch kind {
	case v1alpha1.GrafanaDashboardKind:
		d := o.(*v1alpha1.GrafanaDashboard)
		return h.isKnown(GrafanaDashboardsConfigMapName, d.Namespace, d.Spec.Name)
	case v1alpha1.GrafanaDataSourceKind:
		d := o.(*v1alpha1.GrafanaDataSource)
		return h.isKnown(GrafanaDatasourcesConfigMapName, d.Namespace, d.Spec.Name)
	default:
		return false, stdErrors.New(fmt.Sprintf("unknown kind '%v'", kind))
	}
}

func (h KubeHelperImpl) UpdateDataSources(name, namespace, ds string) (bool, error) {
	configMap, err := h.getConfigMap(GrafanaDatasourcesConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Prefix the data source filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := h.getConfigMapKey(namespace, name)

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[key] = ds
	err = h.updateConfigMap(configMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h KubeHelperImpl) DeleteDataSources(name, namespace string) error {
	configMap, err := h.getConfigMap(GrafanaDatasourcesConfigMapName)
	if err != nil {
		// Grafana may already be uninstalled
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := h.getConfigMapKey(namespace, name)

	if configMap.Data == nil {
		return nil
	}

	if _, ok := configMap.Data[key]; !ok {
		// Resource deleted but no such key in the configmap
		return nil
	}

	delete(configMap.Data, key)
	err = h.updateConfigMap(configMap)
	return err
}

func (h KubeHelperImpl) DeleteDashboard(d *v1alpha1.GrafanaDashboard) error {
	configMap, err := h.getConfigMap(GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			// Grafana may already be uninstalled
			return nil
		}
		return err
	}

	dashboardName := h.getConfigMapKey(d.Namespace, d.Spec.Name)
	if configMap.Data == nil {
		return nil
	}

	if _, ok := configMap.Data[dashboardName]; !ok {
		// Resource deleted but no such key in the configmap
		return nil
	}

	delete(configMap.Data, dashboardName)
	err = h.updateConfigMap(configMap)
	if err != nil {
		return err
	}

	return nil
}

func (h KubeHelperImpl) getGrafanaPod(namespaceName string) (*core.Pod, error) {
	opts := metav1.ListOptions{
		LabelSelector: "app=grafana",
	}

	pods, err := h.k8client.CoreV1().Pods(namespaceName).List(opts)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, errors.NewNotFound(struct {
			Group    string
			Resource string
		}{Group: "core", Resource: "pod"}, GrafanaDeploymentName)
	}

	return &pods.Items[0], nil
}

func (h KubeHelperImpl) UpdateGrafanaDeployment(newEnv string) error {
	monitoringNamespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	deployment, err := h.getGrafanaDeployment()

	if err != nil {
		return err
	}

	updated := false

	// find and update the init container env var
	for i, container := range deployment.Spec.Template.Spec.InitContainers {
		if container.Name == InitContainerName {
			for j, env := range deployment.Spec.Template.Spec.InitContainers[i].Env {
				if env.Name == PluginsEnvVar {
					deployment.Spec.Template.Spec.InitContainers[i].Env[j].Value = newEnv
					updated = true
					break
				}
			}
		}
	}

	if updated {
		_, err := h.k8client.AppsV1().Deployments(monitoringNamespace).Update(deployment)
		return err
	}

	return nil
}

func (h KubeHelperImpl) RestartGrafana() error {
	monitoringNamespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	pod, err := h.getGrafanaPod(monitoringNamespace)

	if err != nil {
		if errors.IsNotFound(err) {
			// No need to restart if grafana has not yet been deployed
			return nil
		}

		return err
	}

	return h.k8client.CoreV1().Pods(monitoringNamespace).Delete(pod.Name, nil)
}

// Append a status message to the origin dashboard of a plugin
func AppendMessage(message string, dashboard *v1alpha1.GrafanaDashboard) {
	if dashboard == nil {
		return
	}

	status := v1alpha1.GrafanaDashboardStatusMessage{
		Message:   message,
		Timestamp: time.Now().Format(time.RFC850),
	}

	dashboard.Status.Messages = append(dashboard.Status.Messages, status)
}
