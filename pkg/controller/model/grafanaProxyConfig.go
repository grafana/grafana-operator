package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaProxyConfig(cr *v1alpha1.GrafanaProxy) (*v1.ConfigMap, error) {
	cfg := config.NewGrafanaProxyConfig(&cr.Spec.Config)
	config, hash := cfg.Write()

	configMap := &v1.ConfigMap{}
	configMap.ObjectMeta = v12.ObjectMeta{
		Name:      GrafanaProxyConfigName,
		Namespace: cr.Namespace,
	}

	// Store the hash of the current configuration for later
	// comparisons
	configMap.Annotations = map[string]string{
		"lastConfig": hash,
	}

	configMap.Data = map[string]string{}
	configMap.Data[GrafanaProxyConfigFileName] = config
	return configMap, nil
}

func GrafanaProxyConfigReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1.ConfigMap) (*v1.ConfigMap, error) {
	reconciled := currentState.DeepCopy()

	cfg := config.NewGrafanaProxyConfig(&cr.Spec.Config)
	config, hash := cfg.Write()

	reconciled.Annotations = map[string]string{
		LastConfigAnnotation: hash,
	}

	reconciled.Data[GrafanaProxyConfigFileName] = config
	return reconciled, nil
}

func GrafanaProxyConfigSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyConfigName,
	}
}
