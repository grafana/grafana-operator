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

type ServiceReconciler struct {
	client        client.Client
	clusterDomain string
}

func NewServiceReconciler(client client.Client, clusterDomain string) reconcilers.OperatorGrafanaReconciler {
	return &ServiceReconciler{
		client:        client,
		clusterDomain: clusterDomain,
	}
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	_ = logf.FromContext(ctx)

	service := model.GetGrafanaService(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, service, func() error {
		service.Spec = v1.ServiceSpec{
			Ports: getServicePorts(cr),
			Selector: map[string]string{
				"app": cr.Name,
			},
			Type: v1.ServiceTypeClusterIP,
		}

		err := v1beta1.Merge(service, cr.Spec.Service)
		if err != nil {
			setInvalidMergeCondition(cr, "Service", err)
			return err
		}

		removeInvalidMergeCondition(cr, "Service")

		if scheme != nil {
			err = controllerutil.SetControllerReference(cr, service, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(service, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if !cr.PreferIngress() {
		// default empty clusterDomain supports automatic openshift certificates:
		// https://docs.openshift.com/container-platform/4.17/security/certificates/service-serving-certificate.html#add-service-certificate_service-serving-certificate
		adminHost := fmt.Sprintf("%v.%v.svc", service.Name, cr.Namespace)
		if r.clusterDomain != "" {
			adminHost += "." + r.clusterDomain
		}

		cr.Status.AdminURL = fmt.Sprintf("%v://%v:%d", getGrafanaServerProtocol(cr), adminHost, int32(GetGrafanaPort(cr))) // #nosec G115
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
	protocol := cr.GetConfigSectionValue("server", "protocol")
	if protocol != "" {
		return protocol
	}

	return config.GrafanaServerProtocol
}

func GetGrafanaPort(cr *v1beta1.Grafana) int {
	port := cr.GetConfigSectionValue("server", "http_port")

	intPort, err := strconv.Atoi(port)
	if err != nil {
		return config.GrafanaHTTPPort
	}

	return intPort
}

func getServicePorts(cr *v1beta1.Grafana) []v1.ServicePort {
	intPort := int32(GetGrafanaPort(cr)) // #nosec G115

	defaultPorts := []v1.ServicePort{
		{
			Name:       config.GrafanaHTTPPortName,
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
