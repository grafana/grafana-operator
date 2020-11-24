package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetLokiHost(cr *v1alpha1.Loki) string {
	if cr.Spec.Ingress == nil {
		return ""
	}
	return cr.Spec.Ingress.Hostname
}

func GetLokiPath(cr *v1alpha1.Loki) string {
	if cr.Spec.Ingress == nil {
		return "/"
	}
	return cr.Spec.Ingress.Path
}

func GetLokiIngressLabels(cr *v1alpha1.Loki) map[string]string {
	if cr.Spec.Ingress == nil {
		return nil
	}
	return cr.Spec.Ingress.Labels
}

func GetLokiIngressAnnotations(cr *v1alpha1.Loki, existing map[string]string) map[string]string {
	if cr.Spec.Ingress == nil {
		return existing
	}
	return MergeAnnotations(cr.Spec.Ingress.Annotations, existing)
}

func GetLokiIngressTargetPort(cr *v1alpha1.Loki) intstr.IntOrString {
	defaultPort := intstr.FromInt(GetLokiPort(cr))

	if cr.Spec.Ingress == nil {
		return defaultPort
	}

	if cr.Spec.Ingress.TargetPort == "" {
		return defaultPort
	}

	return intstr.FromString(cr.Spec.Ingress.TargetPort)
}

func getLokiTermination(cr *v1alpha1.Loki) v1.TLSTerminationType {
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

func getLokiRouteSpec(cr *v1alpha1.Loki) v1.RouteSpec {
	return v1.RouteSpec{
		Host: GetLokiHost(cr),
		Path: GetLokiPath(cr),
		To: v1.RouteTargetReference{
			Kind: "Service",
			Name: LokiServiceName,
		},
		AlternateBackends: nil,
		Port: &v1.RoutePort{
			TargetPort: GetLokiIngressTargetPort(cr),
		},
		TLS: &v1.TLSConfig{
			Termination: getLokiTermination(cr),
		},
		WildcardPolicy: "None",
	}
}

func LokiRoute(cr *v1alpha1.Loki) *v1.Route {
	return &v1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:        LokiRouteName,
			Namespace:   cr.Namespace,
			Labels:      GetLokiIngressLabels(cr),
			Annotations: GetLokiIngressAnnotations(cr, nil),
		},
		Spec: getLokiRouteSpec(cr),
	}
}

func LokiRouteSelector(cr *v1alpha1.Loki) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiRouteName,
	}
}

func LokiRouteReconciled(cr *v1alpha1.Loki, currentState *v1.Route) *v1.Route {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = GetLokiIngressLabels(cr)
	reconciled.Annotations = GetLokiIngressAnnotations(cr, currentState.Annotations)
	reconciled.Spec = getLokiRouteSpec(cr)
	return reconciled
}
