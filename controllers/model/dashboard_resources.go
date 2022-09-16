package model

import (
	"fmt"

	grafanav1beta1 "github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetPluginsConfigMap(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ConfigMap {
	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-plugins", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, config, scheme)
	return config
}

func GetDashboardsConfigMap(cr *grafanav1beta1.Grafana, scheme *runtime.Scheme) *v1.ConfigMap {
	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-dashboards", cr.Name),
			Namespace: cr.Namespace,
		},
	}
	controllerutil.SetOwnerReference(cr, config, scheme)
	return config
}
