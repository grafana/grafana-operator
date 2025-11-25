package model

import (
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func GetCommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "grafana-operator",
	}
}

func GetGrafanaConfigMap(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.ConfigMap {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ini", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, config, scheme) //nolint:errcheck

	return config
}

func GetGrafanaAdminSecret(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.Secret {
	secret := &corev1.Secret{
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

func GetGrafanaDataPVC(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
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

func GetGrafanaServiceAccount(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sa", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, sa, scheme) //nolint:errcheck

	return sa
}

func GetGrafanaService(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, service, scheme) //nolint:errcheck

	return service
}

func GetGrafanaHeadlessService(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-alerting", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, service, scheme) //nolint:errcheck

	return service
}

func GetGrafanaIngress(cr *v1beta1.Grafana, scheme *runtime.Scheme) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, ingress, scheme) //nolint:errcheck

	return ingress
}

func GetGrafanaRoute(cr *v1beta1.Grafana, scheme *runtime.Scheme) *routev1.Route {
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

func GetGrafanaHTTPRoute(cr *v1beta1.Grafana, scheme *runtime.Scheme) *gwapiv1.HTTPRoute {
	httpRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, httpRoute, scheme) //nolint:errcheck

	return httpRoute
}

func GetGrafanaDeployment(cr *v1beta1.Grafana, scheme *runtime.Scheme) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
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
