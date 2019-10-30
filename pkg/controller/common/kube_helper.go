package common

import (
	"fmt"
	v1alpha1client "github.com/integr8ly/grafana-operator/pkg/clientset/v1alpha1"
	"strings"
	"time"

	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("kube_helper")

type KubeHelperImpl struct {
	k8client      *kubernetes.Clientset
	grafanaClient *v1alpha1client.GrafanaV1Alpha1Client
	config        *ControllerConfig
}

func NewKubeHelper() *KubeHelperImpl {
	config := config.GetConfigOrDie()

	k8client := kubernetes.NewForConfigOrDie(config)
	grafanaClient := v1alpha1client.NewForConfigOrDie(config)

	helper := new(KubeHelperImpl)
	helper.k8client = k8client
	helper.grafanaClient = grafanaClient
	helper.config = GetControllerConfig()
	return helper
}

func (h KubeHelperImpl) GetConfigMapKey(namespace, name string, suffix string) string {
	return fmt.Sprintf("%s_%s.%s", namespace, strings.ToLower(name), suffix)
}

func (h KubeHelperImpl) GetConfigMap(name string) (*v1.ConfigMap, error) {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	return h.k8client.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
}

func (h KubeHelperImpl) UpdateConfigMap(c *v1.ConfigMap) error {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	_, err := h.k8client.CoreV1().ConfigMaps(namespace).Update(c)
	return err
}

func (h KubeHelperImpl) getGrafanaDeployment() (*apps.Deployment, error) {
	namespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	return h.k8client.AppsV1().Deployments(namespace).Get(GrafanaDeploymentName, metav1.GetOptions{})
}

func (h KubeHelperImpl) UpdateGrafanaConfig(config string, cr *v1alpha1.Grafana) error {
	configMap, err := h.GetConfigMap(GrafanaConfigMapName)
	if err != nil {
		return err
	}
	configMap.Data[GrafanaConfigFileName] = config
	return h.UpdateConfigMap(configMap)
}

func (h KubeHelperImpl) GetGrafana(name string, namespace string) (*v1alpha1.Grafana, error) {
	grafana, err := h.grafanaClient.Grafanas(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return grafana, nil
}

func (h KubeHelperImpl) ListGrafanas(namespace string) (*v1alpha1.GrafanaList, error) {
	return h.grafanaClient.Grafanas(namespace).List(metav1.ListOptions{})
}

func (h KubeHelperImpl) getGrafanaPods(namespaceName string) ([]core.Pod, error) {
	podLabel := h.config.GetConfigString(ConfigPodLabelValue, PodLabelDefaultValue)

	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", podLabel),
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

	return pods.Items, nil
}

// UpdateGrafanaDeployment is responsible for update the config hash on the deployment
func (h *KubeHelperImpl) UpdateGrafanaDeployment(hash string) error {
	ns := h.config.GetConfigString(ConfigOperatorNamespace, "")
	deployment, err := h.getGrafanaDeployment()
	if err != nil {
		return err
	}
	// update the configuration hash
	deployment.Spec.Template.Spec.Containers[0].Env[0].Value = hash

	if _, err = h.k8client.AppsV1().Deployments(ns).Update(deployment); err != nil {
		return err
	}

	return nil
}

// UpdateGrafanaInitContainersDeployment updates the initcontainer config environment vars
func (h KubeHelperImpl) UpdateGrafanaInitContainersDeployment(newEnv string) error {
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

// RestartGrafana is responsible for restarting the pods
func (h KubeHelperImpl) RestartGrafana() error {
	monitoringNamespace := h.config.GetConfigString(ConfigOperatorNamespace, "")
	pods, err := h.getGrafanaPods(monitoringNamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			// No need to restart if grafana has not yet been deployed
			return nil
		}

		return err
	}

	if len(pods) == 1 {
		return h.k8client.CoreV1().Pods(monitoringNamespace).Delete(pods[0].Name, nil)
	}

	// @step: iterate and delete the pods
	for _, pod := range pods {
		if err := h.k8client.CoreV1().Pods(monitoringNamespace).Delete(pod.Name, nil); err != nil {
			return err
		}
		time.Sleep(time.Second * 30)
	}

	return nil
}

func (h KubeHelperImpl) RestartGrafanaIfNeeded() error {
	if h.config.GetConfigBool(ConfigGrafanaApi, false) {
		return nil
	} else {
		return h.RestartGrafana()
	}
}

// AppendMessage a status message to the origin dashboard of a plugin
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
