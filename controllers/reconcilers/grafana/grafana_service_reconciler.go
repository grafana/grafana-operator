package grafana

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
)

type ServiceReconciler struct {
	client client.Client
}

func NewServiceReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &ServiceReconciler{
		client: client,
	}
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	_ = log.FromContext(ctx)

	service := model.GetGrafanaService(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		service.Labels = getServiceLabels(cr)
		service.Annotations = getServiceAnnotations(cr, service.Annotations)
		service.Spec = v1.ServiceSpec{
			Ports: getServicePorts(cr, service),
			Selector: map[string]string{
				"app": "grafana",
			},
			ClusterIP: getClusterIP(cr),
			Type:      getServiceType(cr),
		}
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getServiceLabels(cr *v1beta1.Grafana) map[string]string {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service.Labels
}

func getServiceAnnotations(cr *v1beta1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.Service == nil {
		return existing
	}

	return model.MergeAnnotations(cr.Spec.Service.Annotations, existing)
}

func getServiceType(cr *v1beta1.Grafana) v1.ServiceType {
	if cr.Spec.Service == nil {
		return v1.ServiceTypeClusterIP
	}
	if cr.Spec.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cr.Spec.Service.Type
}

func getClusterIP(cr *v1beta1.Grafana) string {
	if cr.Spec.Service == nil {
		return ""
	}
	return cr.Spec.Service.ClusterIP
}

func GetGrafanaPort(cr *v1beta1.Grafana) int {
	if cr.Spec.Config.Server == nil {
		return config.GrafanaHttpPort
	}

	if cr.Spec.Config.Server.HttpPort == "" {
		return config.GrafanaHttpPort
	}

	port, err := strconv.Atoi(cr.Spec.Config.Server.HttpPort)
	if err != nil {
		return config.GrafanaHttpPort
	}

	return port
}

func getServicePorts(cr *v1beta1.Grafana, currentState *v1.Service) []v1.ServicePort {
	intPort := int32(GetGrafanaPort(cr))

	defaultPorts := []v1.ServicePort{
		{
			Name:       config.GrafanaHttpPortName,
			Protocol:   "TCP",
			Port:       intPort,
			TargetPort: intstr.FromString("grafana-http"),
		},
	}

	if cr.Spec.Service == nil {
		return defaultPorts
	}

	// Re-assign existing node port
	if cr.Spec.Service != nil &&
		currentState != nil {
		for _, port := range currentState.Spec.Ports {
			if port.Name == config.GrafanaHttpPortName {
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
		if port.Name == config.GrafanaHttpPortName || port.Port == intPort {
			continue
		}
		defaultPorts = append(defaultPorts, port)
	}

	return defaultPorts
}
