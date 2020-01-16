package model

import (
	"fmt"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getProxyServicePorts(cr *v1alpha1.GrafanaProxy, currentState *v1.Service) []v1.ServicePort {
	defaultPorts := []v1.ServicePort{
		{
			Name:       "grafana-proxy",
			Protocol:   "TCP",
			Port:       80,
			TargetPort: intstr.FromString("http"),
		},
	}
	return defaultPorts
}

func GrafanaProxyService(cr *v1alpha1.GrafanaProxy) *v1.Service {
	return &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaProxyServiceName,
			Namespace:   cr.Namespace,
			Labels:      getProxyPodLabels(cr),
			Annotations: getProxyPodAnnotations(cr),
		},
		Spec: v1.ServiceSpec{
			Ports: getProxyServicePorts(cr, nil),
			Selector: map[string]string{
				"app": GrafanaProxyPodLabel,
			},
			ClusterIP: "",
			Type:      v1.ServiceTypeClusterIP,
		},
	}
}

func GrafanaProxyServiceReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1.Service) *v1.Service {
	fmt.Println(*currentState)
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getProxyPodLabels(cr)
	reconciled.Annotations = getProxyPodAnnotations(cr)
	reconciled.Spec.Ports = getProxyServicePorts(cr, currentState)
	reconciled.Spec.Type = v1.ServiceTypeClusterIP
	return reconciled
}

func GrafanaProxyServiceSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyServiceName,
	}
}
