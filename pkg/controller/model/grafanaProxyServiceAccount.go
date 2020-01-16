package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaProxyServiceAccount(cr *v1alpha1.GrafanaProxy) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaProxyServiceAccountName,
			Namespace:   cr.Namespace,
			Labels:      cr.Labels,
			Annotations: cr.Annotations,
		},
	}
}

func GrafanaProxyServiceAccountSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyServiceAccountName,
	}
}

func GrafanaProxyServiceAccountReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1.ServiceAccount) *v1.ServiceAccount {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getProxyPodLabels(cr)
	reconciled.Annotations = getProxyPodAnnotations(cr)
	return reconciled
}
