package grafana

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		service.Spec = v1.ServiceSpec{
			Ports: getServicePorts(cr),
			Selector: map[string]string{
				"app": cr.Name,
			},
			Type: v1.ServiceTypeClusterIP,
		}
		return v1beta1.Merge(service, cr.Spec.Service)
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if !cr.PreferIngress() {
		status.AdminUrl = fmt.Sprintf("%v://%v.%v:%d", getGrafanaServerProtocol(cr), service.Name, cr.Namespace,
			int32(GetGrafanaPort(cr)))
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getGrafanaServerProtocol(cr *v1beta1.Grafana) string {
	if cr.Spec.Config != nil && cr.Spec.Config["server"] != nil && cr.Spec.Config["server"]["protocol"] != "" {
		return cr.Spec.Config["server"]["protocol"]
	}
	return config.GrafanaServerProtocol
}

func GetGrafanaPort(cr *v1beta1.Grafana) int {
	if cr.Spec.Config["server"] == nil {
		return config.GrafanaHttpPort
	}

	if cr.Spec.Config["server"]["http_port"] == "" {
		return config.GrafanaHttpPort
	}

	port, err := strconv.Atoi(cr.Spec.Config["server"]["http_port"])
	if err != nil {
		return config.GrafanaHttpPort
	}

	return port
}

func getServicePorts(cr *v1beta1.Grafana) []v1.ServicePort {
	intPort := int32(GetGrafanaPort(cr))

	defaultPorts := []v1.ServicePort{
		{
			Name:       config.GrafanaHttpPortName,
			Protocol:   "TCP",
			Port:       intPort,
			TargetPort: intstr.FromString("grafana-http"),
		},
	}

	return defaultPorts
}
