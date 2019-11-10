package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	v1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRouteSpec(cr *v1alpha1.Grafana) v1.RouteSpec {
	return v1.RouteSpec{
		Path: cr.Spec.Ingress.Path,
		To: v1.RouteTargetReference{
			Kind: "Service",
			Name: GrafanaServiceName,
		},
		AlternateBackends: nil,
		Port: &v1.RoutePort{
			TargetPort: intstr.FromInt(GrafanaHttpPort),
		},
		TLS: &v1.TLSConfig{
			Termination: "edge",
		},
		WildcardPolicy: "None",
	}
}

func GrafanaRoute(cr *v1alpha1.Grafana) *v1.Route {
	return &v1.Route{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaRouteName,
			Namespace:   cr.Namespace,
			Labels:      cr.Spec.Ingress.Labels,
			Annotations: cr.Spec.Ingress.Annotations,
		},
		Spec: getRouteSpec(cr),
	}
}

func GrafanaRouteSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaRouteName,
	}
}

func GrafanaRouteReconciled(cr *v1alpha1.Grafana, currentState *v1.Route) *v1.Route {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = cr.Spec.Ingress.Labels
	reconciled.Annotations = cr.Spec.Ingress.Annotations
	reconciled.Spec = getRouteSpec(cr)
	return reconciled
}
