package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha2"
	v1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetHost(cr *v1alpha2.Grafana) string {
	if cr.Spec.Ingress == nil {
		return ""
	}
	return cr.Spec.Ingress.Hostname
}

func GetPath(cr *v1alpha2.Grafana) string {
	if cr.Spec.Ingress == nil {
		return "/"
	}
	return cr.Spec.Ingress.Path
}

func GetIngressLabels(cr *v1alpha2.Grafana) map[string]string {
	if cr.Spec.Ingress == nil {
		return nil
	}
	return cr.Spec.Ingress.Labels
}

func GetIngressAnnotations(cr *v1alpha2.Grafana) map[string]string {
	if cr.Spec.Ingress == nil {
		return nil
	}
	return cr.Spec.Ingress.Annotations
}

func GetIngressTargetPort(cr *v1alpha2.Grafana) intstr.IntOrString {
	defaultPort := intstr.FromInt(GetGrafanaPort(cr))

	if cr.Spec.Ingress == nil {
		return defaultPort
	}

	if cr.Spec.Ingress.TargetPort == "" {
		return defaultPort
	}

	return intstr.FromString(cr.Spec.Ingress.TargetPort)
}

func getTermination(cr *v1alpha2.Grafana) v1.TLSTerminationType {
	if cr.Spec.Ingress == nil {
		return v1.TLSTerminationEdge
	}

	switch cr.Spec.Ingress.Termination {
	case v1.TLSTerminationEdge:
		return v1.TLSTerminationEdge
	case v1.TLSTerminationReencrypt:
		return v1.TLSTerminationReencrypt
	case v1.TLSTerminationPassthrough:
		return v1.TLSTerminationPassthrough
	default:
		return v1.TLSTerminationEdge
	}
}

func getRouteSpec(cr *v1alpha2.Grafana) v1.RouteSpec {
	return v1.RouteSpec{
		Host: GetHost(cr),
		Path: GetPath(cr),
		To: v1.RouteTargetReference{
			Kind: "Service",
			Name: GrafanaServiceName,
		},
		AlternateBackends: nil,
		Port: &v1.RoutePort{
			TargetPort: GetIngressTargetPort(cr),
		},
		TLS: &v1.TLSConfig{
			Termination: getTermination(cr),
		},
		WildcardPolicy: "None",
	}
}

func GrafanaRoute(cr *v1alpha2.Grafana) *v1.Route {
	return &v1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaRouteName,
			Namespace:   cr.Namespace,
			Labels:      GetIngressLabels(cr),
			Annotations: GetIngressAnnotations(cr),
		},
		Spec: getRouteSpec(cr),
	}
}

func GrafanaRouteSelector(cr *v1alpha2.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaRouteName,
	}
}

func GrafanaRouteReconciled(cr *v1alpha2.Grafana, currentState *v1.Route) *v1.Route {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = GetIngressLabels(cr)
	reconciled.Annotations = GetIngressAnnotations(cr)
	reconciled.Spec = getRouteSpec(cr)
	return reconciled
}
