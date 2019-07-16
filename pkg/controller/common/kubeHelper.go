package common

import (
	"fmt"
	"strings"

	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"

)

type KubeHelperImpl struct {
	k8client *kubernetes.Clientset
}

func NewKubeHelper() *KubeHelperImpl {
	config := config.GetConfigOrDie()

	k8client := kubernetes.NewForConfigOrDie(config)

	helper := new(KubeHelperImpl)
	helper.k8client = k8client
	return helper
}

func (h KubeHelperImpl) getConfigMap(namespace, name string) (*v1.ConfigMap, error) {
	opts := metav1.GetOptions{}
	return h.k8client.CoreV1().ConfigMaps(namespace).Get(name, opts)
}

func (h KubeHelperImpl) getGrafanaDeployment(namespaceName string) (*apps.Deployment, error) {
	opts := metav1.GetOptions{}
	return h.k8client.AppsV1().Deployments(namespaceName).Get(GrafanaDeploymentName, opts)
}

func (h KubeHelperImpl) UpdateGrafanaConfig(config string, cr *v1alpha1.Grafana) error {
	configMap, err := h.getConfigMap(cr.Namespace, GrafanaConfigMapName)
	if err != nil {
		return err
	}

	configMap.Data[GrafanaConfigFileName] = config
	_, err = h.k8client.CoreV1().ConfigMaps(cr.Namespace).Update(configMap)
	return err
}

func (h KubeHelperImpl) UpdateDashboard(ns string, d *v1alpha1.GrafanaDashboard) (bool, error) {
	configMap, err := h.getConfigMap(ns, GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	dashboardName := fmt.Sprintf("%s_%s", d.Namespace, d.Spec.Name)

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[dashboardName] = d.Spec.Json
	configMap, err = h.k8client.CoreV1().ConfigMaps(ns).Update(configMap)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (h KubeHelperImpl) IsKnownDataSource(ds *v1alpha1.GrafanaDataSource) (bool, error) {
	configMap, err := h.getConfigMap(ds.Namespace, GrafanaDatasourcesConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if configMap.Data == nil {
		return false, nil
	}

	key := fmt.Sprintf("%s_%s", ds.Namespace, strings.ToLower(ds.Spec.Name))
	_, found := configMap.Data[key]

	return found, nil
}

func (h KubeHelperImpl) IsKnownDashboard(ds *v1alpha1.GrafanaDashboard) (bool, error) {
	configMap, err := h.getConfigMap(ds.Namespace, GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if configMap.Data == nil {
		return false, nil
	}

	key := fmt.Sprintf("%s_%s", ds.Namespace, strings.ToLower(ds.Spec.Name))
	_, found := configMap.Data[key]

	return found, nil
}

func (h KubeHelperImpl) UpdateDataSources(name, namespace, ds string) (bool, error) {
	configMap, err := h.getConfigMap(namespace, GrafanaDatasourcesConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	// Prefix the data source filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := fmt.Sprintf("%s_%s", namespace, strings.ToLower(name))

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[key] = ds
	configMap, err = h.k8client.CoreV1().ConfigMaps(namespace).Update(configMap)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h KubeHelperImpl) DeleteDataSources(name, namespace string) error {
	configMap, err := h.getConfigMap(namespace, GrafanaDatasourcesConfigMapName)
	if err != nil {
		// Grafana may already be uninstalled
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := fmt.Sprintf("%s_%s", namespace, strings.ToLower(name))

	if configMap.Data == nil {
		return nil
	}

	if _, ok := configMap.Data[key]; !ok {
		// Resource deleted but no such key in the configmap
		return nil
	}

	delete(configMap.Data, key)
	configMap, err = h.k8client.CoreV1().ConfigMaps(namespace).Update(configMap)
	return err
}

func (h KubeHelperImpl) DeleteDashboard(monitoringNamespace string, dashboardNamespace string, dashboard *v1alpha1.GrafanaDashboard) error {
	configMap, err := h.getConfigMap(monitoringNamespace, GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			// Grafana may already be uninstalled
			return nil
		}
		return err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	dashboardName := fmt.Sprintf("%s_%s", dashboardNamespace, dashboard.Spec.Name)

	if configMap.Data == nil {
		return nil
	}

	if _, ok := configMap.Data[dashboardName]; !ok {
		// Resource deleted but no such key in the configmap
		return nil
	}

	delete(configMap.Data, dashboardName)
	configMap, err = h.k8client.CoreV1().ConfigMaps(monitoringNamespace).Update(configMap)
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

func (h KubeHelperImpl) UpdateGrafanaDeployment(monitoringNamespace string, newEnv string) error {
	deployment, err := h.getGrafanaDeployment(monitoringNamespace)
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

func (h KubeHelperImpl) RestartGrafana(monitoringNamespace string) error {
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
func (h KubeHelperImpl) AppendMessage(message string, dashboard *v1alpha1.GrafanaDashboard) {
	if dashboard == nil {
		return
	}

	status := v1alpha1.GrafanaDashboardStatusMessage{
		Message:   message,
		Timestamp: time.Now().Format(time.RFC850),
	}

	dashboard.Status.Messages = append(dashboard.Status.Messages, status)
}