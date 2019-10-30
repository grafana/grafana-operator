package common

import (
	"context"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

type ResponseStatus struct {
	Succeeded bool
}

type GrafanaData struct {
	kubeHelper *KubeHelperImpl
}

type Grafana interface {
	UpdateDashboard(context.Context, *v1alpha1.GrafanaDashboard) (ResponseStatus, error)
	GetDashboard(context.Context, *v1alpha1.GrafanaDashboard) (ResponseStatus, error)
	DeleteDashboard(context.Context, *v1alpha1.GrafanaDashboard) (ResponseStatus, error)
	DashboardIsKnown(context.Context, *v1alpha1.GrafanaDashboard) (bool, error)

	UpdateDataSource(context.Context, *v1alpha1.GrafanaDataSource) (ResponseStatus, error)
	GetDataSource(context.Context, *v1alpha1.GrafanaDataSource) (ResponseStatus, error)
	DeleteDataSource(context.Context, *v1alpha1.GrafanaDataSource) (ResponseStatus, error)
	DataSourceIsKnown(context.Context, *v1alpha1.GrafanaDataSource) (bool, error)
}

func NewGrafana(kubeHelper *KubeHelperImpl) Grafana {
	config := GetControllerConfig()
	grafanaData := &GrafanaData{kubeHelper: kubeHelper}
	if config.GetConfigBool(ConfigGrafanaApi, false) {
		return NewGrafanaApiImpl(grafanaData)
	} else {
		return NewConfigMapImpl(grafanaData)
	}
}
