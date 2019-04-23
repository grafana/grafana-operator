package grafana

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	gr "github.com/integr8ly/grafana-operator/pkg/client/versioned"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	InitContainerName = "grafana-plugins-init"
)

type KubeHelperImpl struct {
	k8client *kubernetes.Clientset
	grclient *gr.Clientset
}

func newKubeHelper() *KubeHelperImpl {
	config := config.GetConfigOrDie()

	k8client := kubernetes.NewForConfigOrDie(config)
	grclient := gr.NewForConfigOrDie(config)

	helper := new(KubeHelperImpl)
	helper.k8client = k8client
	helper.grclient = grclient
	return helper
}

func (h KubeHelperImpl) getMonitoringNamespaces(ls *metav1.LabelSelector) ([]v1.Namespace, error) {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return nil, err
	}

	var selectorString string
	metav1.Convert_labels_Selector_To_string(&selector, &selectorString, nil)
	opts := metav1.ListOptions{
		LabelSelector: selectorString,
	}

	namespaces, err := h.k8client.CoreV1().Namespaces().List(opts)
	if err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}

func (h KubeHelperImpl) getNamespaceDashboards(namespaceName string) (*v1alpha1.GrafanaDashboardList, error) {
	selector := metav1.ListOptions{}
	dashboards, err := h.grclient.IntegreatlyV1alpha1().GrafanaDashboards(namespaceName).List(selector)

	if err != nil {
		return nil, err
	}

	return dashboards, nil
}

func (h KubeHelperImpl) getDashboardsConfigMap(namespaceName string) (*v1.ConfigMap, error) {
	opts := metav1.GetOptions{}
	return h.k8client.CoreV1().ConfigMaps(namespaceName).Get("grafana-dashboards", opts)
}

func (h KubeHelperImpl) updateDashboard(monitoringNamespace string, dashboardNamespace string, dashboard *v1alpha1.GrafanaDashboard) error {
	configMap, err := h.getDashboardsConfigMap(monitoringNamespace)
	if err != nil {
		return err
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[dashboard.Spec.Name] = dashboard.Spec.Json
	configMap, err = h.k8client.CoreV1().ConfigMaps(monitoringNamespace).Update(configMap)
	return err
}

func (h KubeHelperImpl) getGrafanaDeployment(namespaceName string) (*apps.Deployment, error) {
	opts := metav1.GetOptions{}
	return h.k8client.AppsV1().Deployments(namespaceName).Get(GrafanaDeploymentName, opts)
}

func (h KubeHelperImpl) updateGrafanaDeployment(monitoringNamespace string, newEnv string) error {
	deployment, err := h.getGrafanaDeployment(monitoringNamespace)
	if err != nil {
		return err
	}

	// Leave the deployment alone when it's busy with another operation
	if deployment.Status.Replicas != deployment.Status.ReadyReplicas {
		return nil
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
