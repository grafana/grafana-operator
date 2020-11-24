package model



import (
	"strconv"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getLokiServiceLabels(cr *v1alpha1.Loki) map[string]string {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service.Labels
}

func getLokiServiceAnnotations(cr *v1alpha1.Loki, existing map[string]string) map[string]string {
	if cr.Spec.Service == nil {
		return existing
	}

	return MergeAnnotations(cr.Spec.Service.Annotations, existing)
}

func getLokiServiceType(cr *v1alpha1.Loki) v1.ServiceType {
	if cr.Spec.Service == nil {
		return v1.ServiceTypeClusterIP
	}
	if cr.Spec.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cr.Spec.Service.Type
}

func getLokiClusterIP(cr *v1alpha1.Loki) string {
	if cr.Spec.Service == nil {
		return ""
	}
	return cr.Spec.Service.ClusterIP
}

func getLokiPort(cr *v1alpha1.Loki) int {
	if cr.Spec.Config.Server == nil {
		return LokiHttpPort
	}

	if cr.Spec.Config.Server.HttpPort == "" {
		return GrafanaHttpPort
	}

	port, err := strconv.Atoi(cr.Spec.Config.Server.HttpPort)
	if err != nil {
		return GrafanaHttpPort
	}

	return port
}

func getLokiPrefix(cr *v1alpha1.Loki) string {
	if cr.Spec.Config.Server == nil {
		return LokiHttpPrefix
	}

	if cr.Spec.Config.Server.HttpPrefix == "" {
		return LokiHttpPrefix
	}
	return cr.Spec.Config.Server.HttpPrefix
}

func getLokiServicePorts(cr *v1alpha1.Loki, currentState *v1.Service) []v1.ServicePort {
	intPort := int32(getLokiPort(cr))

	defaultPorts := []v1.ServicePort{
		{
			Name:       LokiHttpPortName,
			Protocol:   "TCP",
			Port:       intPort,
			TargetPort: intstr.FromString("loki-http"),
		},
	}

	if cr.Spec.Service == nil {
		return defaultPorts
	}

	// Re-assign existing node port
	if cr.Spec.Service != nil &&
		currentState != nil {
		for _, port := range currentState.Spec.Ports {
			if port.Name == LokiHttpPortName {
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
		if port.Name == GrafanaHttpPortName || port.Port == intPort {
			continue
		}
		defaultPorts = append(defaultPorts, port)
	}

	return defaultPorts
}

func LokiService(cr *v1alpha1.Loki) *v1.Service {
	return &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        LokiServiceName,
			Namespace:   cr.Namespace,
			Labels:      getLokiServiceLabels(cr),
			Annotations: getLokiServiceAnnotations(cr, nil),
		},
		Spec: v1.ServiceSpec{
			Ports: getLokiServicePorts(cr, nil),
			Selector: map[string]string{
				"app": LokiPodLabel,
			},
			ClusterIP: getLokiClusterIP(cr),
			Type:      getLokiServiceType(cr),
		},
	}
}

func LokiServiceReconciled(cr *v1alpha1.Loki, currentState *v1.Service) *v1.Service {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getLokiServiceLabels(cr)
	reconciled.Annotations = getLokiServiceAnnotations(cr, currentState.Annotations)
	reconciled.Spec.Ports = getLokiServicePorts(cr, currentState)
	reconciled.Spec.Type = getLokiServiceType(cr)
	return reconciled
}

func LokiServiceSelector(cr *v1alpha1.Loki) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiServiceName,
	}
}
