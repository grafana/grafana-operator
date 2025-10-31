package model

import (
	"fmt"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	v2 "sigs.k8s.io/gateway-api/apis/v1"
)

func GetCommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "grafana-operator",
	}
}

func GetGrafanaConfigMap(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ConfigMap {
	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ini", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, config, scheme) //nolint:errcheck

	return config
}

func GetGrafanaAdminSecret(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-admin-credentials", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}

	if scheme != nil {
		controllerutil.SetControllerReference(cr, secret, scheme) //nolint:errcheck
	}

	return secret
}

func GetGrafanaDataPVC(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-pvc", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	// using OwnerReference specifically here to allow admins to change storage variables without the operator complaining
	controllerutil.SetOwnerReference(cr, pvc, scheme) //nolint:errcheck

	return pvc
}

func GetGrafanaServiceAccount(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ServiceAccount {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sa", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, sa, scheme) //nolint:errcheck

	return sa
}

func GetGrafanaService(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, service, scheme) //nolint:errcheck

	return service
}

func GetGrafanaHeadlessService(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-alerting", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, service, scheme) //nolint:errcheck

	return service
}

func GetGrafanaIngress(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v12.Ingress {
	ingress := &v12.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, ingress, scheme) //nolint:errcheck

	return ingress
}

func GetGrafanaHTTPRoute(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v2.HTTPRoute {
	httpRoute := &v2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, httpRoute, scheme) //nolint:errcheck

	return httpRoute
}

func GetGrafanaRoute(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, route, scheme) //nolint:errcheck

	return route
}

func GetGrafanaDeployment(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v13.Deployment {
	deployment := &v13.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-deployment", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	if scheme != nil {
		controllerutil.SetControllerReference(cr, deployment, scheme) //nolint:errcheck
	}

	return deployment
}
