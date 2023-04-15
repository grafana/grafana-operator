package grafana

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RouteKind = "Route"
)

type IngressReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	IsOpenShift bool
}

func GetGrafanaIngressMeta(cr *v1beta1.Grafana) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
}

func GetGrafanaRouteMeta(cr *v1beta1.Grafana) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error {
	logger := log.FromContext(ctx)

	if r.IsOpenShift {
		logger.Info("reconciling route", "platform", "openshift")
		return r.reconcileRoute(ctx, cr, next)
	} else {
		logger.Info("reconciling ingress", "platform", "kubernetes")
		return r.reconcileIngress(ctx, cr, next)
	}
}

func (r *IngressReconciler) reconcileIngress(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error {
	if cr.Spec.Ingress == nil || len(cr.Spec.Ingress.Spec.Rules) == 0 {
		return nil
	}

	ingress := GetGrafanaIngressMeta(cr)
	if err := controllerutil.SetControllerReference(cr, ingress, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		ingress.Spec = mergeIngressSpec(cr, r.Scheme)
		return v1beta1.Merge(&ingress.ObjectMeta, cr.Spec.Ingress.ObjectMeta)
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		adminURL := r.getIngressAdminURL(ingress)

		if len(ingress.Status.LoadBalancer.Ingress) == 0 {
			return fmt.Errorf("ingress is not ready yet")
		}

		if adminURL == "" {
			return fmt.Errorf("ingress spec is incomplete")
		}

		next.Status.AdminUrl = adminURL
	}

	return nil
}

func (r *IngressReconciler) reconcileRoute(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error {
	route := GetGrafanaRouteMeta(cr)
	if err := controllerutil.SetControllerReference(cr, route, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, route, func() error {
		route.Spec = getRouteSpec(cr, r.Scheme)
		err := v1beta1.Merge(route, cr.Spec.Route)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	// try to assign the admin url
	if cr.PreferIngress() {
		if route.Spec.Host != "" {
			next.Status.AdminUrl = fmt.Sprintf("https://%s", route.Spec.Host)
		}
	}

	return nil
}

// getIngressAdminURL returns the first valid URL (Host field is set) from the ingress spec
func (r *IngressReconciler) getIngressAdminURL(ingress *networkingv1.Ingress) string {
	if ingress == nil {
		return ""
	}

	protocol := "http"
	var hostname string
	var adminURL string

	// An ingress rule might not have the field Host specified, better not to consider such rules
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			hostname = rule.Host
			break
		}
	}

	// If we can find the target host in any of the IngressTLS, then we should use https protocol
	for _, tls := range ingress.Spec.TLS {
		for _, h := range tls.Hosts {
			if h == hostname {
				protocol = "https"
				break
			}
		}
	}

	// if all fails, try to get access through the load balancer
	if hostname == "" {
		loadBalancerIp := ""
		for _, lb := range ingress.Status.LoadBalancer.Ingress {
			if lb.Hostname != "" {
				hostname = lb.Hostname
				break
			}
			if lb.IP != "" {
				loadBalancerIp = lb.IP
			}
		}

		if hostname == "" && loadBalancerIp != "" {
			hostname = loadBalancerIp
		}
	}

	// adminUrl should not be empty only in case hostname is found, otherwise we'll have broken URLs like "http://"
	if hostname != "" {
		adminURL = fmt.Sprintf("%s://%s", protocol, hostname)
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
	service := GetGrafanaServiceMeta(cr)

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

func mergeIngressSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) networkingv1.IngressSpec {
	service := GetGrafanaServiceMeta(cr)

	port := GetIngressTargetPort(cr)
	var assignedPort networkingv1.ServiceBackendPort
	if port.IntVal > 0 {
		assignedPort.Number = port.IntVal
	}
	if port.StrVal != "" {
		assignedPort.Name = port.StrVal
	}

	pathType := networkingv1.PathTypePrefix
	path := networkingv1.HTTPIngressPath{
		Path:     "/",
		PathType: &pathType,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: service.Name,
				Port: assignedPort,
			},
			Resource: nil,
		},
	}

	res := networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{}}

	for _, rule := range cr.Spec.Ingress.Spec.Rules {
		if rule.HTTP == nil || rule.HTTP.Paths == nil {
			rule.HTTP = &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{},
			}
		}

		rule.HTTP.Paths = append(rule.HTTP.Paths, path)
		res.Rules = append(res.Rules, rule)
	}
	return res
}
