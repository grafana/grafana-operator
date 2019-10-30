package common

import (
	"context"
	stdErrors "errors"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigMapData struct {
	*GrafanaData
}

func NewConfigMapImpl(grafanaData *GrafanaData) *ConfigMapData {
	return &ConfigMapData{
		GrafanaData: grafanaData,
	}
}

func (h ConfigMapData) UpdateDashboard(ctx context.Context, d *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	configMap, err := h.kubeHelper.GetConfigMap(GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			resp.Succeeded = false
			return resp, nil
		}

		return resp, err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	dashboardName := h.kubeHelper.GetConfigMapKey(d.Namespace, d.Spec.Dashboard.Name, "json")
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	configMap.Data[dashboardName] = d.Spec.Dashboard.Json
	err = h.kubeHelper.UpdateConfigMap(configMap)
	if err != nil {
		return resp, err
	}

	resp.Succeeded = true
	return resp, nil
}

func (h ConfigMapData) DeleteDashboard(ctx context.Context, d *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	configMap, err := h.kubeHelper.GetConfigMap(GrafanaDashboardsConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			// Grafana may already be uninstalled
			resp.Succeeded = true
			return resp, nil
		}
		return resp, err
	}

	dashboardName := h.kubeHelper.GetConfigMapKey(d.Namespace, d.Spec.Dashboard.Name, "json")
	if configMap.Data == nil {
		resp.Succeeded = true
		return resp, nil
	}

	if _, ok := configMap.Data[dashboardName]; !ok {
		// Resource deleted but no such key in the configmap
		resp.Succeeded = true
		return resp, nil
	}

	delete(configMap.Data, dashboardName)
	err = h.kubeHelper.UpdateConfigMap(configMap)
	if err != nil {
		return resp, err
	}
	resp.Succeeded = true
	return resp, nil
}

func (h ConfigMapData) GetDashboard(ctx context.Context, d *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	return resp, stdErrors.New("not implemented")
}

func (h ConfigMapData) UpdateDataSource(ctx context.Context, d *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	configMap, err := h.kubeHelper.GetConfigMap(GrafanaDatasourcesConfigMapName)
	if err != nil {
		if errors.IsNotFound(err) {
			resp.Succeeded = false
			return resp, nil
		}
		return resp, err
	}

	// Prefix the data source filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := h.kubeHelper.GetConfigMapKey(d.Namespace, d.Spec.DataSource.Name, "yaml")

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	ds, err := v1alpha1.ParseDataSource(d)
	if err != nil {
		return resp, err
	}
	configMap.Data[key] = ds
	err = h.kubeHelper.UpdateConfigMap(configMap)
	if err != nil {
		return resp, err
	}
	resp.Succeeded = true
	return resp, nil
}

func (h ConfigMapData) DashboardIsKnown(ctx context.Context, d *v1alpha1.GrafanaDashboard) (bool, error) {
	known, err := h.IsKnown(v1alpha1.GrafanaDashboardKind, d)
	if err != nil {
		return false, err
	}
	return known, nil
}

func (h ConfigMapData) DeleteDataSource(ctx context.Context, d *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	configMap, err := h.kubeHelper.GetConfigMap(GrafanaDatasourcesConfigMapName)
	if err != nil {
		// Grafana may already be uninstalled
		if errors.IsNotFound(err) {
			return resp, nil
		}
		return resp, err
	}

	// Prefix the dashboard filename with the namespace to allow multiple namespaces
	// to import the same dashboard
	key := h.kubeHelper.GetConfigMapKey(d.Namespace, d.Name, "yaml")

	if configMap.Data == nil {
		return resp, nil
	}

	if _, ok := configMap.Data[key]; !ok {
		// Resource deleted but no such key in the configmap
		return resp, nil
	}

	delete(configMap.Data, key)
	err = h.kubeHelper.UpdateConfigMap(configMap)
	return resp, err
}

func (h ConfigMapData) GetDataSource(ctx context.Context, d *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	return resp, fmt.Errorf("not implemented")
}

func (h ConfigMapData) isKnown(config, namespace, name, suffix string) (bool, error) {
	configMap, err := h.kubeHelper.GetConfigMap(config)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if configMap.Data == nil {
		return false, nil
	}

	key := h.kubeHelper.GetConfigMapKey(namespace, name, suffix)
	_, found := configMap.Data[key]
	return found, nil
}

func (h ConfigMapData) IsKnown(kind string, o runtime.Object) (bool, error) {
	switch kind {
	case v1alpha1.GrafanaDashboardKind:
		d := o.(*v1alpha1.GrafanaDashboard)
		return h.isKnown(GrafanaDashboardsConfigMapName, d.Namespace, d.Spec.Dashboard.Name, "json")
	case v1alpha1.GrafanaDataSourceKind:
		d := o.(*v1alpha1.GrafanaDataSource)
		return h.isKnown(GrafanaDatasourcesConfigMapName, d.Namespace, d.Spec.DataSource.Name, "yaml")
	default:
		return false, stdErrors.New(fmt.Sprintf("unknown kind '%v'", kind))
	}
}

func (h ConfigMapData) DataSourceIsKnown(ctx context.Context, d *v1alpha1.GrafanaDataSource) (bool, error) {
	known, err := h.IsKnown(v1alpha1.GrafanaDataSourceKind, d)
	if err != nil {
		return false, err
	}
	return known, nil
}
