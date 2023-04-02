package grafana

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	"github.com/grafana-operator/grafana-operator/v5/controllers/model"
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

func (r *ServiceReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {

	service := model.GetGrafanaService(cr, r.Scheme) // todo: inline  model

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
		return nil, err // toodo: error condition
	}

	// try to assign the admin url
	if !cr.PreferIngress() {
		cr.Status.AdminUrl = fmt.Sprintf("%v://%v.%v:%d", getGrafanaServerProtocol(cr), service.Name, cr.Namespace,
			int32(GetGrafanaPort(cr)))
	}

	return nil, nil // todo; success condition
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
