package model

import (
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getDatasourcesConfigName(cr *v1alpha1.Grafana) string {
	if cr.Spec.DatasourceConfig != nil && cr.Spec.DatasourceConfig.Name != "" {
		return cr.Spec.DatasourceConfig.Name
	}
	return constants.GrafanaDatasourcesConfigMapName
}

func GrafanaDatasourcesConfig(cr *v1alpha1.Grafana) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      getDatasourcesConfigName(cr),
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				constants.LastConfigAnnotation: "",
			},
		},
	}
}

func GrafanaDatasourceConfigSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      getDatasourcesConfigName(cr),
	}
}
