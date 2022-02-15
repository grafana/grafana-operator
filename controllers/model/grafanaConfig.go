package model

import (
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaConfig(cr *v1alpha1.Grafana) *v1.ConfigMap {
	ini := config.NewGrafanaIni(&cr.Spec.Config)
	config, hash := ini.Write()

	configMap := &v1.ConfigMap{}
	configMap.ObjectMeta = v12.ObjectMeta{
		Name:      constants.GrafanaConfigName,
		Namespace: cr.Namespace,
	}

	// Store the hash of the current configuration for later
	// comparisons
	configMap.Annotations = map[string]string{
		"lastConfig": hash,
	}

	configMap.Data = map[string]string{}
	configMap.Data[constants.GrafanaConfigFileName] = config
	return configMap
}

func GrafanaConfigReconciled(cr *v1alpha1.Grafana, currentState *v1.ConfigMap) *v1.ConfigMap {
	reconciled := currentState.DeepCopy()

	ini := config.NewGrafanaIni(&cr.Spec.Config)
	config, hash := ini.Write()

	reconciled.Annotations = map[string]string{
		constants.LastConfigAnnotation: hash,
	}

	reconciled.Data[constants.GrafanaConfigFileName] = config
	return reconciled
}

func GrafanaConfigSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      constants.GrafanaConfigName,
	}
}
