package grafana

import (
	"context"
	"fmt"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RouteKind = "Route"
)

type IngressReconciler struct {
	client    client.Client
	discovery discovery.DiscoveryInterface
}

func NewIngressReconciler(client client.Client, discovery discovery.DiscoveryInterface) reconcilers.OperatorGrafanaReconciler {
	return &IngressReconciler{
		client:    client,
		discovery: discovery,
	}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx)

	openshift, err := r.isOpenShift()
	if err != nil {
		logger.Error(err, "error determining platform")
		return v1beta1.OperatorStageResultFailed, err
	}

	if openshift {
		logger.Info("platform is OpenShift, creating Route")
		return r.reconcileRoute(ctx, cr, status, vars, scheme)
	} else {
		logger.Info("platform is Kubernetes, creating Ingress")
		return r.reconcileIngress(ctx, cr, status, vars, scheme)
	}
}

func (r *IngressReconciler) reconcileIngress(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	ingress := model.GetGrafanaIngress(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, ingress, func() error {
		ingress.Spec = getIngressSpec(cr, scheme)
		return v1beta1.Merge(ingress, cr.Spec.Ingress)
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			ingress := ingress.Status.LoadBalancer.Ingress[0]
			if ingress.Hostname != "" {
				status.AdminUrl = fmt.Sprintf("https://%v", ingress.Hostname)
			}
			status.AdminUrl = fmt.Sprintf("https://%v", ingress.IP)
		}
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) reconcileRoute(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	route := model.GetGrafanaRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, route, func() error {
		route.Spec = getRouteSpec(cr, scheme)
		err := v1beta1.Merge(route, cr.Spec.Route)
		return err
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		if route.Spec.Host != "" {
			status.AdminUrl = fmt.Sprintf("https://%v", route.Spec.Host)
		}
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) isOpenShift() (bool, error) {
	apiGroupVersion := routev1.SchemeGroupVersion.String()

	apiList, err := r.discovery.ServerResourcesForGroupVersion(apiGroupVersion)
	if apiList == nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, r := range apiList.APIResources {
		if r.Kind == RouteKind {
			return true, nil
		}
	}
	return false, nil
}

func getRouteTLS(cr *v1beta1.Grafana) *routev1.TLSConfig {
	return &routev1.TLSConfig{
		Certificate:                   "",
		Key:                           "",
		CACertificate:                 "",
		DestinationCACertificate:      "",
		InsecureEdgeTerminationPolicy: "",
	}
}

func GetIngressTargetPort(cr *v1beta1.Grafana) intstr.IntOrString {
	return intstr.FromInt(GetGrafanaPort(cr))
}

func getRouteSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) routev1.RouteSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetIngressTargetPort(cr)

	return routev1.RouteSpec{
		To: routev1.RouteTargetReference{
			Kind: "Service",
			Name: service.Name,
		},
		AlternateBackends: nil,
		Port: &routev1.RoutePort{
			TargetPort: port,
		},
		TLS:            getRouteTLS(cr),
		WildcardPolicy: "None",
	}
}

func getIngressSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) v1.IngressSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetIngressTargetPort(cr)
	var assignedPort v1.ServiceBackendPort
	if port.IntVal > 0 {
		assignedPort.Number = port.IntVal
	}
	if port.StrVal != "" {
		assignedPort.Name = port.StrVal
	}

	pathType := v1.PathTypePrefix
	return v1.IngressSpec{
		Rules: []v1.IngressRule{
			{
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: service.Name,
										Port: assignedPort,
									},
									Resource: nil,
								},
							},
						},
					},
				},
			},
		},
	}
}
