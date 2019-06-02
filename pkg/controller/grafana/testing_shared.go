package grafana

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MockCR = v1alpha1.Grafana{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test",
		Namespace: "dummy",
	},
	Spec: v1alpha1.GrafanaSpec{
		Containers: []v12.Container{},
	},
}

var Templates = []string{
	common.GrafanaDeploymentName,
	common.GrafanaDashboardsConfigMapName,
	common.GrafanaDatasourcesConfigMapName,
	common.GrafanaRouteName,
	common.GrafanaProvidersConfigMapName,
	common.GrafanaServiceAccountName,
	common.GrafanaServiceName,
}

var Mockplugina100 = v1alpha1.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.0",
}

var Mockplugina101 = v1alpha1.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.1",
}

var Mockplugina102 = v1alpha1.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.2",
}

var Mockpluginb100 = v1alpha1.GrafanaPlugin{
	Name:    "b",
	Version: "1.0.0",
}

var Mockpluginc100 = v1alpha1.GrafanaPlugin{
	Name:    "c",
	Version: "1.0.0",
}

var MockPluginList = v1alpha1.PluginList{Mockplugina100, Mockplugina101, Mockpluginb100}

var MockDashboard = v1alpha1.GrafanaDashboard{
	Status: v1alpha1.GrafanaDashboardStatus{
		Messages: []v1alpha1.GrafanaDashboardStatusMessage{},
	},
}

var MockGrafana = v1alpha1.Grafana{
	Status: v1alpha1.GrafanaStatus{
		Phase:            0,
		InstalledPlugins: v1alpha1.PluginList{},
	},
}
