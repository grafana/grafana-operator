package model

import (
	"strconv"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getServiceName(cr *v1alpha1.Grafana) string {
	if cr.Spec.Service != nil && cr.Spec.Service.Name != "" {
		return cr.Spec.Service.Name
	}
	return constants.GrafanaServiceName
}

func getServiceLabels(cr *v1alpha1.Grafana) map[string]string {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service.Labels
}

func getServiceAnnotations(cr *v1alpha1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.Service == nil {
		return existing
	}

	return MergeAnnotations(cr.Spec.Service.Annotations, existing)
}

func getServiceType(cr *v1alpha1.Grafana) v1.ServiceType {
	if cr.Spec.Service == nil {
		return v1.ServiceTypeClusterIP
	}
	if cr.Spec.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cr.Spec.Service.Type
}

func getClusterIP(cr *v1alpha1.Grafana) string {
	if cr.Spec.Service == nil {
		return ""
	}
	return cr.Spec.Service.ClusterIP
}

func GetGrafanaPort(cr *v1alpha1.Grafana) int {
	if cr.Spec.Config.Server == nil {
		return constants.GrafanaHttpPort
	}

	if cr.Spec.Config.Server.HttpPort == "" {
		return constants.GrafanaHttpPort
	}

	port, err := strconv.Atoi(cr.Spec.Config.Server.HttpPort)
	if err != nil {
		return constants.GrafanaHttpPort
	}

	return port
}

func GetGrafanaAlertManagerPort(cr *v1alpha1.Grafana) int {
	return constants.GrafanaAlertManagerPort
}

func unifiedAlertsEnabled(cr *v1alpha1.Grafana) bool {
	if cr.Spec.Config.UnifiedAlerting == nil {
		return false
	}
	if cr.Spec.Config.UnifiedAlerting.Enabled == nil {
		return false
	}

	return *cr.Spec.Config.UnifiedAlerting.Enabled
}
func getServicePorts(cr *v1alpha1.Grafana, currentState *v1.Service) []v1.ServicePort {
	intPort := int32(GetGrafanaPort(cr))

	defaultPorts := []v1.ServicePort{
		{
			Name:       constants.GrafanaHttpPortName,
			Protocol:   "TCP",
			Port:       intPort,
			TargetPort: intstr.FromString("grafana-http"),
		},
	}

	if unifiedAlertsEnabled(cr) {
		alertManagerPort := int32(GetGrafanaAlertManagerPort(cr))
		defaultPorts = append(defaultPorts, v1.ServicePort{
			Name:       constants.GrafanaAlertManagerTCPPortName,
			Protocol:   "TCP",
			Port:       alertManagerPort,
			TargetPort: intstr.FromString("tcp-alert"),
		})
		defaultPorts = append(defaultPorts, v1.ServicePort{
			Name:       constants.GrafanaAlertManagerUDPPortName,
			Protocol:   "UDP",
			Port:       alertManagerPort,
			TargetPort: intstr.FromString("udp-alert"),
		})
	}

	if cr.Spec.Service == nil {
		return defaultPorts
	}

	// Re-assign existing node port
	if cr.Spec.Service != nil &&
		currentState != nil {
		for _, port := range currentState.Spec.Ports {
			if port.Name == constants.GrafanaHttpPortName {
				defaultPorts[0].NodePort = port.NodePort
			}
		}
	}

	if cr.Spec.Service.Ports == nil {
		return defaultPorts
	}

	// Don't allow overriding the default port but allow adding
	// additional ports
	for _, port := range cr.Spec.Service.Ports {
		if port.Name == constants.GrafanaHttpPortName || port.Port == intPort {
			continue
		}
		defaultPorts = append(defaultPorts, port)
	}

	return defaultPorts
}

func GrafanaService(cr *v1alpha1.Grafana) *v1.Service {
	return &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        getServiceName(cr),
			Namespace:   cr.Namespace,
			Labels:      getServiceLabels(cr),
			Annotations: getServiceAnnotations(cr, nil),
		},
		Spec: v1.ServiceSpec{
			Ports: getServicePorts(cr, nil),
			Selector: map[string]string{
				"app": constants.GrafanaPodLabel,
			},
			ClusterIP: getClusterIP(cr),
			Type:      getServiceType(cr),
		},
	}
}

func GrafanaServiceReconciled(cr *v1alpha1.Grafana, currentState *v1.Service) *v1.Service {
	reconciled := currentState.DeepCopy()
	reconciled.Name = getServiceName(cr)
	reconciled.Labels = getServiceLabels(cr)
	reconciled.Annotations = getServiceAnnotations(cr, currentState.Annotations)
	reconciled.Spec.Ports = getServicePorts(cr, currentState)
	reconciled.Spec.Type = getServiceType(cr)
	return reconciled
}

func GrafanaServiceSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      getServiceName(cr),
	}
}
