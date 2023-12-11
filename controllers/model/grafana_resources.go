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
)

func GetGrafanaConfigMap(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ConfigMap {
	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ini", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, config, scheme) //nolint:errcheck
	return config
}

func GetGrafanaAdminSecret(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-admin-credentials", cr.Name),
			Namespace: cr.Namespace,
		},
	}

	if scheme != nil {
		controllerutil.SetOwnerReference(cr, secret, scheme) //nolint:errcheck
	}
	return secret
}

func GetGrafanaDataPVC(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-pvc", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, pvc, scheme) //nolint:errcheck
	return pvc
}

func GetGrafanaServiceAccount(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ServiceAccount {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sa", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, sa, scheme) //nolint:errcheck
	return sa
}

func GetGrafanaService(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, service, scheme) //nolint:errcheck
	return service
}

func GetGrafanaIngress(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v12.Ingress {
	ingress := &v12.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, ingress, scheme) //nolint:errcheck
	return ingress
}

func GetGrafanaRoute(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, route, scheme) //nolint:errcheck
	return route
}

func GetGrafanaDeployment(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v13.Deployment {
	deployment := &v13.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-deployment", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	if scheme != nil {
		controllerutil.SetOwnerReference(cr, deployment, scheme) //nolint:errcheck
	}
	return deployment
}
