package fetchers

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FetchDashboardFromConfigMap(cr v1beta1.GrafanaContentResource, cl client.Client) ([]byte, error) {
	spec := cr.GrafanaContentSpec()
	if spec == nil {
		return nil, nil // TODO
	}

	ref := spec.ConfigMapRef
	dashboardConfigMap := &corev1.ConfigMap{}
	selector := client.ObjectKey{
		Namespace: cr.GetNamespace(),
		Name:      ref.Name,
	}

	err := cl.Get(context.Background(), selector, dashboardConfigMap)
	if err != nil {
		return nil, err
	}

	if content, ok := dashboardConfigMap.Data[ref.Key]; ok {
		return []byte(content), nil
	}

	return nil, fmt.Errorf("cannot find key '%v' in config map '%v' for dashboard %v/%v",
		ref.Key, ref.Name, cr.GetNamespace(), cr.GetName())
}
