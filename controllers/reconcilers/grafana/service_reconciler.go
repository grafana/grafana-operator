package grafana

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func GetGrafanaServiceMeta(cr *v1beta1.Grafana) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-grafana", cr.Name),
			Namespace: cr.Namespace,
		},
	}
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) error {
	service := GetGrafanaServiceMeta(cr)
	if err := controllerutil.SetControllerReference(cr, service, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
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
		return fmt.Errorf("failed to create or update: %w", err)
	}

	// try to assign the admin url
	if !cr.PreferIngress() {
		cr.Status.AdminUrl = fmt.Sprintf("%v://%v.%v:%d", getGrafanaServerProtocol(cr), service.Name, cr.Namespace,
			int32(GetGrafanaPort(cr)))
		// TODO: update status?
	}

	return nil
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
