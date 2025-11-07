package model

import (
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetPluginsConfigMap(cr *v1beta1.Grafana, scheme *runtime.Scheme) *corev1.ConfigMap {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-plugins", cr.Name),
			Namespace: cr.Namespace,
			Labels:    GetCommonLabels(),
		},
	}
	controllerutil.SetControllerReference(cr, config, scheme) //nolint:errcheck

	return config
}
