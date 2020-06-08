package testing

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
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

var MockGrafana = v1alpha1.Grafana{
	Status: v1alpha1.GrafanaStatus{
		Phase:            v1alpha1.PhaseReconciling,
		InstalledPlugins: v1alpha1.PluginList{},
	},
}
