package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LokiServiceAccountSelector(cr *v1alpha1.Loki) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiServiceAccountName,
	}
}
