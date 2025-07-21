package grafana

import (
	"context"
	"fmt"
	"slices"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RouteKind = "Route"
)

type IngressReconciler struct {
	client      client.Client
	isOpenShift bool
}

func NewIngressReconciler(client client.Client, isOpenShift bool) reconcilers.OperatorGrafanaReconciler {
	return &IngressReconciler{
		client:      client,
		isOpenShift: isOpenShift,
	}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("IngressReconciler")

	// On openshift, Fallback to Ingress when spec.route is undefined
	if r.isOpenShift && cr.Spec.Route != nil {
		log.Info("reconciling route")
		return r.reconcileRoute(ctx, cr, vars, scheme)
	}

	log.Info("reconciling ingress")

	return r.reconcileIngress(ctx, cr, vars, scheme)
}

func (r *IngressReconciler) reconcileIngress(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	if cr.Spec.Ingress == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	ingress := model.GetGrafanaIngress(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, ingress, func() error {
		ingress.Spec = getIngressSpec(cr, scheme)

		err := v1beta1.Merge(ingress, cr.Spec.Ingress)
		if err != nil {
			setInvalidMergeCondition(cr, "Ingress", err)
			return err
		}

		removeInvalidMergeCondition(cr, "Ingress")

		err = controllerutil.SetControllerReference(cr, ingress, scheme)
		if err != nil {
			return err
		}

		model.SetInheritedLabels(ingress, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		adminURL := r.getIngressAdminURL(ingress)

		if len(ingress.Status.LoadBalancer.Ingress) == 0 {
			return v1beta1.OperatorStageResultInProgress, fmt.Errorf("ingress is not ready yet")
		}

		if adminURL == "" {
			return v1beta1.OperatorStageResultFailed, fmt.Errorf("ingress spec is incomplete")
		}

		cr.Status.AdminURL = adminURL
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) reconcileRoute(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	if cr.Spec.Route == nil || cr.Spec.Route.Spec == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	route := model.GetGrafanaRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, route, func() error {
		route.Spec = getRouteSpec(cr, scheme)

		err := v1beta1.Merge(route, cr.Spec.Route)
		if err != nil {
			setInvalidMergeCondition(cr, "Route", err)
			return err
		}

		removeInvalidMergeCondition(cr, "Route")

		if scheme != nil {
			err = controllerutil.SetControllerReference(cr, route, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(route, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		if route.Spec.Host != "" {
			cr.Status.AdminURL = fmt.Sprintf("https://%v", route.Spec.Host)
		}
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

// getIngressAdminURL returns the first valid URL (Host field is set) from the ingress spec
func (r *IngressReconciler) getIngressAdminURL(ingress *v1.Ingress) string {
	if ingress == nil {
		return ""
	}

	protocol := "http"

	var (
		hostname string
		adminURL string
	)

	// An ingress rule might not have the field Host specified, better not to consider such rules

	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			hostname = rule.Host
			break
		}
	}

	// If we can find the target host in any of the IngressTLS, then we should use https protocol
	for _, tls := range ingress.Spec.TLS {
		if slices.Contains(tls.Hosts, hostname) {
			protocol = "https"
		}
	}

	// if all fails, try to get access through the load balancer
	if hostname == "" {
		loadBalancerIP := ""

		for _, lb := range ingress.Status.LoadBalancer.Ingress {
			if lb.Hostname != "" {
				hostname = lb.Hostname
				break
			}

			if lb.IP != "" {
				loadBalancerIP = lb.IP
			}
		}

		if hostname == "" && loadBalancerIP != "" {
			hostname = loadBalancerIP
		}
	}

	// adminUrl should not be empty only in case hostname is found, otherwise we'll have broken URLs like "http://"
	if hostname != "" {
		adminURL = fmt.Sprintf("%v://%v", protocol, hostname)
	}

	return adminURL
}

func getRouteTLS() *routev1.TLSConfig {
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
		TLS:            getRouteTLS(),
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
