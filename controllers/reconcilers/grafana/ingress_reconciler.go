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
		ingress.Annotations = GetIngressAnnotations(cr, ingress.Annotations)
		ingress.Labels = GetIngressLabels(cr)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		// if provided use the hostname
		if cr.Spec.Ingress != nil && cr.Spec.Ingress.Hostname != "" {
			status.AdminUrl = fmt.Sprintf("https://%v", cr.Spec.Ingress.Hostname)
		} else {
			// Otherwise try to find something suitable, hostname or IP
			if len(ingress.Status.LoadBalancer.Ingress) > 0 {
				ingress := ingress.Status.LoadBalancer.Ingress[0]
				if ingress.Hostname != "" {
					status.AdminUrl = fmt.Sprintf("https://%v", ingress.Hostname)
				}
				status.AdminUrl = fmt.Sprintf("https://%v", ingress.IP)
			}
		}
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) reconcileRoute(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	route := model.GetGrafanaRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, route, func() error {
		route.Spec = getRouteSpec(cr, scheme)
		route.Annotations = GetIngressAnnotations(cr, route.Annotations)
		route.Labels = GetIngressLabels(cr)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) isOpenShift() (bool, error) {
	apiGroupVersion := routev1.SchemeGroupVersion.String()

	apiList, err := r.discovery.ServerResourcesForGroupVersion(apiGroupVersion)
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

func getIngressTLS(cr *v1beta1.Grafana) []v1.IngressTLS {
	if cr.Spec.Ingress == nil {
		return nil
	}

	if cr.Spec.Ingress.TLSEnabled {
		return []v1.IngressTLS{
			{
				Hosts:      []string{cr.Spec.Ingress.Hostname},
				SecretName: cr.Spec.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func getTermination(cr *v1beta1.Grafana) routev1.TLSTerminationType {
	if cr.Spec.Ingress == nil {
		return routev1.TLSTerminationEdge
	}

	switch cr.Spec.Ingress.Termination {
	case routev1.TLSTerminationEdge:
		return routev1.TLSTerminationEdge
	case routev1.TLSTerminationReencrypt:
		return routev1.TLSTerminationReencrypt
	case routev1.TLSTerminationPassthrough:
		return routev1.TLSTerminationPassthrough
	default:
		return routev1.TLSTerminationEdge
	}
}

func getRouteTLS(cr *v1beta1.Grafana) *routev1.TLSConfig {
	return &routev1.TLSConfig{
		Termination:                   getTermination(cr),
		Certificate:                   "",
		Key:                           "",
		CACertificate:                 "",
		DestinationCACertificate:      "",
		InsecureEdgeTerminationPolicy: "",
	}
}

func GetIngressPathType(cr *v1beta1.Grafana) *v1.PathType {
	defaultPathType := v1.PathTypeExact

	if cr.Spec.Ingress == nil {
		return &defaultPathType
	}

	t := v1.PathType(cr.Spec.Ingress.PathType)
	switch t {
	case v1.PathTypeExact, v1.PathTypePrefix:
		return &t
	case v1.PathTypeImplementationSpecific:
		t = v1.PathTypeImplementationSpecific
		return &t
	default:
		return &defaultPathType
	}
}

func GetIngressClassName(cr *v1beta1.Grafana) *string {
	if cr.Spec.Ingress == nil || cr.Spec.Ingress.IngressClassName == "" {
		return nil
	}

	return &cr.Spec.Ingress.IngressClassName
}

func GetIngressTargetPort(cr *v1beta1.Grafana) intstr.IntOrString {
	defaultPort := intstr.FromInt(GetGrafanaPort(cr))

	if cr.Spec.Ingress == nil {
		return defaultPort
	}

	if cr.Spec.Ingress.TargetPort == "" {
		return defaultPort
	}

	return intstr.FromString(cr.Spec.Ingress.TargetPort)
}

func GetHost(cr *v1beta1.Grafana) string {
	if cr.Spec.Ingress == nil {
		return ""
	}
	return cr.Spec.Ingress.Hostname
}

func GetPath(cr *v1beta1.Grafana) string {
	if cr.Spec.Ingress == nil {
		return "/"
	}

	if cr.Spec.Ingress.Path == "" {
		return "/"
	}

	return cr.Spec.Ingress.Path
}

func getRouteSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) routev1.RouteSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetIngressTargetPort(cr)

	return routev1.RouteSpec{
		Host: GetHost(cr),
		Path: GetPath(cr),
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

	if port.IntVal != 0 {
		return v1.IngressSpec{
			TLS:              getIngressTLS(cr),
			IngressClassName: GetIngressClassName(cr),
			Rules: []v1.IngressRule{
				{
					Host: GetHost(cr),
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     GetPath(cr),
									PathType: GetIngressPathType(cr),
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: service.Name,
											Port: v1.ServiceBackendPort{
												Number: port.IntVal,
											},
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
	return v1.IngressSpec{
		TLS:              getIngressTLS(cr),
		IngressClassName: GetIngressClassName(cr),
		Rules: []v1.IngressRule{
			{
				Host: GetHost(cr),
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path:     GetPath(cr),
								PathType: GetIngressPathType(cr),
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: service.Name,
										Port: v1.ServiceBackendPort{
											Name: port.StrVal,
										},
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

func GetIngressLabels(cr *v1beta1.Grafana) map[string]string {
	if cr.Spec.Ingress == nil {
		return nil
	}
	return cr.Spec.Ingress.Labels
}

func GetIngressAnnotations(cr *v1beta1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.Ingress == nil {
		return existing
	}
	return model.MergeAnnotations(cr.Spec.Ingress.Annotations, existing)
}
