package grafana

import (
	"context"
	"fmt"
	"slices"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	RouteKind = "Route"
)

type IngressReconciler struct {
	client        client.Client
	isOpenShift   bool
	hasGatewayAPI bool
}

func NewIngressReconciler(client client.Client, isOpenShift bool, hasGatewayAPI bool) reconcilers.OperatorGrafanaReconciler {
	return &IngressReconciler{
		client:        client,
		isOpenShift:   isOpenShift,
		hasGatewayAPI: hasGatewayAPI,
	}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("IngressReconciler")

	if r.isOpenShift {
		err := r.deleteRouteIfNil(ctx, cr, scheme)
		if err != nil {
			return v1beta1.OperatorStageResultFailed, err
		}
	}

	if r.hasGatewayAPI {
		err := r.deleteHTTPRouteIfNil(ctx, cr, scheme)
		if err != nil {
			return v1beta1.OperatorStageResultFailed, err
		}
	}

	err := r.deleteIngressIfNil(ctx, cr, scheme)
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// On openshift, Fallback to Ingress when spec.route is undefined
	if r.isOpenShift && cr.Spec.Route != nil {
		log.Info("reconciling route")
		return r.reconcileRoute(ctx, cr, vars, scheme)
	}

	if cr.Spec.HTTPRoute != nil {
		log.Info("reconciling HTTPRoute")
		return r.reconcileHTTPRoute(ctx, cr, vars, scheme)
	}

	log.Info("reconciling ingress")

	return r.reconcileIngress(ctx, cr, vars, scheme)
}

func (r *IngressReconciler) deleteIngressIfNil(ctx context.Context, cr *v1beta1.Grafana, scheme *runtime.Scheme) error {
	if cr.Spec.Ingress != nil {
		return nil
	}

	ingress := model.GetGrafanaIngress(cr, scheme)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      ingress.Name,
			Namespace: ingress.Namespace,
		},
	}

	err := r.client.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error getting Ingress: %w", err)
	}

	return r.client.Delete(ctx, ingress)
}

func (r *IngressReconciler) deleteHTTPRouteIfNil(ctx context.Context, cr *v1beta1.Grafana, scheme *runtime.Scheme) error {
	if cr.Spec.HTTPRoute != nil {
		return nil
	}

	route := model.GetGrafanaHTTPRoute(cr, scheme)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      route.Name,
			Namespace: route.Namespace,
		},
	}

	err := r.client.Get(ctx, req.NamespacedName, route)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error getting HTTPRoute: %w", err)
	}

	return r.client.Delete(ctx, route)
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

func (r *IngressReconciler) deleteRouteIfNil(ctx context.Context, cr *v1beta1.Grafana, scheme *runtime.Scheme) error {
	if cr.Spec.Route != nil {
		return nil
	}

	route := model.GetGrafanaRoute(cr, scheme)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      route.Name,
			Namespace: route.Namespace,
		},
	}

	err := r.client.Get(ctx, req.NamespacedName, route)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error getting Route: %w", err)
	}

	return r.client.Delete(ctx, route)
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

func (r *IngressReconciler) reconcileHTTPRoute(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	if cr.Spec.HTTPRoute == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	route := model.GetGrafanaHTTPRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, route, func() error {
		route.Spec = getHTTPRouteSpec(cr, scheme)

		err := v1beta1.Merge(route, cr.Spec.HTTPRoute)
		if err != nil {
			setInvalidMergeCondition(cr, "HTTPRoute", err)
			return err
		}

		removeInvalidMergeCondition(cr, "HTTPRoute")

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

	return v1beta1.OperatorStageResultSuccess, nil
}

// getIngressAdminURL returns the first valid URL (Host field is set) from the ingress spec
func (r *IngressReconciler) getIngressAdminURL(ingress *networkingv1.Ingress) string {
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

func getHTTPRouteSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) gwapiv1.HTTPRouteSpec {
	service := model.GetGrafanaService(cr, scheme)
	port := gwapiv1.PortNumber(GetGrafanaPort(cr)) //nolint:gosec
	backendRefs := []gwapiv1.HTTPBackendRef{
		{
			BackendRef: gwapiv1.BackendRef{
				BackendObjectReference: gwapiv1.BackendObjectReference{
					Name: gwapiv1.ObjectName(service.Name),
					Port: &port,
				},
			},
		},
	}

	return gwapiv1.HTTPRouteSpec{
		Rules: []gwapiv1.HTTPRouteRule{
			{
				BackendRefs: backendRefs,
			},
		},
	}
}

func getIngressSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) networkingv1.IngressSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetIngressTargetPort(cr)

	var assignedPort networkingv1.ServiceBackendPort
	if port.IntVal > 0 {
		assignedPort.Number = port.IntVal
	}

	if port.StrVal != "" {
		assignedPort.Name = port.StrVal
	}

	pathType := networkingv1.PathTypePrefix

	return networkingv1.IngressSpec{
		Rules: []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
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
