package testing

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha2"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MockCR = v1alpha2.Grafana{
	ObjectMeta: v1.ObjectMeta{
		Name:      "test",
		Namespace: "dummy",
	},
	Spec: v1alpha2.GrafanaSpec{
		Containers: []v12.Container{},
	},
}

var Mockplugina100 = v1alpha2.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.0",
}

var Mockplugina101 = v1alpha2.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.1",
}

var Mockplugina102 = v1alpha2.GrafanaPlugin{
	Name:    "a",
	Version: "1.0.2",
}

var Mockpluginb100 = v1alpha2.GrafanaPlugin{
	Name:    "b",
	Version: "1.0.0",
}

var Mockpluginc100 = v1alpha2.GrafanaPlugin{
	Name:    "c",
	Version: "1.0.0",
}

var MockPluginList = v1alpha2.PluginList{Mockplugina100, Mockplugina101, Mockpluginb100}

var MockDashboard = v1alpha2.GrafanaDashboard{
	Status: v1alpha2.GrafanaDashboardStatus{},
}

var MockGrafana = v1alpha2.Grafana{
	Status: v1alpha2.GrafanaStatus{
		Phase:            v1alpha2.PhaseReconciling,
		InstalledPlugins: v1alpha2.PluginList{},
	},
}
