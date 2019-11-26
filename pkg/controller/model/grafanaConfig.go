package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaConfig(cr *v1alpha1.Grafana) (*v1.ConfigMap, error) {
	grafanaIni := config.NewIniConfig(cr)
	err := grafanaIni.Build()
	if err != nil {
		return nil, err
	}

	configMap := &v1.ConfigMap{}
	configMap.ObjectMeta = v12.ObjectMeta{
		Name:      GrafanaConfigName,
		Namespace: cr.Namespace,
	}

	// Store the hash of the current configuration for later
	// comparisons
	configMap.Annotations = map[string]string{
		"lastConfig": grafanaIni.Hash,
	}

	configMap.Data = map[string]string{}
	configMap.Data[GrafanaConfigFileName] = grafanaIni.Contents
	return configMap, nil
}

func GrafanaConfigReconciled(cr *v1alpha1.Grafana, currentState *v1.ConfigMap) (*v1.ConfigMap, error) {
	reconciled := currentState.DeepCopy()

	grafanaIni := config.NewIniConfig(cr)
	err := grafanaIni.Build()
	if err != nil {
		return nil, err
	}

	reconciled.Annotations = map[string]string{
		LastConfigAnnotation: grafanaIni.Hash,
	}

	reconciled.Data[GrafanaConfigFileName] = grafanaIni.Contents
	return reconciled, nil
}

func GrafanaConfigSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaConfigName,
	}
}
