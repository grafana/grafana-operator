package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaDatasourcesConfig(cr *v1alpha1.Grafana) (*v1.ConfigMap) {
	configMap := &v1.ConfigMap{}
	configMap.ObjectMeta = v12.ObjectMeta{
		Name:      config.GrafanaDatasourcesConfigMapName,
		Namespace: cr.Namespace,
	}

	return configMap
}

func GrafanaDatasourcesConfigReconciled(cr *v1alpha1.Grafana, currentState *v1.ConfigMap) (*v1.ConfigMap) {
	reconciled := currentState.DeepCopy()

	return reconciled
}

func GrafanaDatasourceConfigSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      config.GrafanaDatasourcesConfigMapName,
	}
}
