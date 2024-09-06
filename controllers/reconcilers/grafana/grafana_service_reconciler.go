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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// .svc suffix needed for automatic openshift certificates: https://docs.openshift.com/container-platform/4.17/security/certificates/service-serving-certificate.html#add-service-certificate_service-serving-certificate
const defaultClusterLocalDomain = ".svc"

type ServiceReconciler struct {
	client             client.Client
	clusterLocalDomain string
}

func NewServiceReconciler(client client.Client, clusterLocalDomain string) reconcilers.OperatorGrafanaReconciler {
	if clusterLocalDomain == "" {
		clusterLocalDomain = defaultClusterLocalDomain
	}
	return &ServiceReconciler{
		client:             client,
		clusterLocalDomain: clusterLocalDomain,
	}
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	_ = logf.FromContext(ctx)

	service := model.GetGrafanaService(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		model.SetInheritedLabels(service, cr.Labels)
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
		adminHost := service.Name + "." + cr.Namespace + r.clusterLocalDomain
		status.AdminUrl = fmt.Sprintf("%v://%v:%d", getGrafanaServerProtocol(cr), adminHost, int32(GetGrafanaPort(cr))) // #nosec G115
	}

	// Headless service for grafana unified alerting
	headlessService := model.GetGrafanaHeadlessService(cr, scheme)
	_, err = controllerutil.CreateOrUpdate(ctx, r.client, headlessService, func() error {
		model.SetInheritedLabels(headlessService, cr.Labels)
		headlessService.Spec = v1.ServiceSpec{
			ClusterIP: "None",
			Ports:     getHeadlessServicePorts(cr),
			Selector: map[string]string{
				"app": cr.Name,
			},
			Type: v1.ServiceTypeClusterIP,
		}
		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
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
	intPort := int32(GetGrafanaPort(cr)) // #nosec G115

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

func getHeadlessServicePorts(_ *v1beta1.Grafana) []v1.ServicePort {
	intPort := int32(config.GrafanaAlertPort)

	defaultPorts := []v1.ServicePort{
		{
			Name:       config.GrafanaAlertPortName,
			Protocol:   "TCP",
			Port:       intPort,
			TargetPort: intstr.FromInt32(intPort),
		},
	}

	return defaultPorts
}
