package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func LokiConfig(cr *v1alpha1.Loki) (*v1.ConfigMap, error) {
	// TODO
	return nil, nil
}
