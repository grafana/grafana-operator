package model

import (
	"fmt"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	v1 "k8s.io/api/core/v1"
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
	controllerutil.SetOwnerReference(cr, config, scheme)
	return config
}

func GetGrafanaAdminSecret(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-admin-credentials", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, secret, scheme)
	return secret
}
