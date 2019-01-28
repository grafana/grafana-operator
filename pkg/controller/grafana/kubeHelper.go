package grafana

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	gr "github.com/integr8ly/grafana-operator/pkg/client/versioned"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type KubeHelper interface {
	listNamespaces()
}

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

func (h KubeHelperImpl) getMonitoringNamespaces() ([]v1.Namespace, error) {
	selector := metav1.ListOptions{
		LabelSelector: "monitoring=enabled",
	}

	namespaces, err := h.k8client.CoreV1().Namespaces().List(selector)
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
	_, err = h.k8client.CoreV1().ConfigMaps(monitoringNamespace).Update(configMap)

	if err == nil {
		dashboard.Status.Created = true
		_, err = h.grclient.IntegreatlyV1alpha1().GrafanaDashboards(dashboardNamespace).Update(dashboard)
	}

	return err
}
