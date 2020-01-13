package grafana

import (
	testing2 "github.com/integr8ly/grafana-operator/v3/pkg/controller/testing"
	"testing"
)

func TestPluginsList(t *testing.T) {
	result := true
	result = result && testing2.MockPluginList.HasSomeVersionOf(&testing2.Mockplugina100)
	result = result && testing2.MockPluginList.HasSomeVersionOf(&testing2.Mockplugina101)
	result = result && testing2.MockPluginList.HasSomeVersionOf(&testing2.Mockpluginb100)
	result = result && testing2.MockPluginList.HasSomeVersionOf(&testing2.Mockplugina102)

	if !result {
		t.Errorf("Error in `HasSomeVersionOf`")
	}

	result = result && testing2.MockPluginList.HasExactVersionOf(&testing2.Mockplugina100)
	result = result && testing2.MockPluginList.HasExactVersionOf(&testing2.Mockplugina101)
	result = result && testing2.MockPluginList.HasExactVersionOf(&testing2.Mockpluginb100)
	result = result && (testing2.MockPluginList.HasExactVersionOf(&testing2.Mockplugina102) == false)
	result = result && (testing2.MockPluginList.HasExactVersionOf(&testing2.Mockpluginc100) == false)

	if !result {
		t.Errorf("Error in `HasExactVersionOf`")
	}

	result, err := testing2.MockPluginList.HasNewerVersionOf(&testing2.Mockplugina100)
	if err != nil {
		t.Error(err)
	}

	if !result {
		t.Errorf("Error in `HasNewerVersionOf`")
	}

	result, err = testing2.MockPluginList.HasNewerVersionOf(&testing2.Mockplugina101)
	if err != nil {
		t.Error(err)
	}

	if result {
		t.Errorf("Error in `HasNewerVersionOf`")
	}
}

func TestPluginsHelperImpl_PickLatestVersions(t *testing.T) {
	var h PluginsHelperImpl

	latestVersions, err := h.pickLatestVersions(testing2.MockPluginList)
	if err != nil {
		t.Error(err)
	}

	if latestVersions.HasExactVersionOf(&testing2.Mockplugina100) {
		t.Errorf("Expected %s but got %s", testing2.Mockplugina101.Version, testing2.Mockplugina100.Version)
	}

	result, err := latestVersions.HasNewerVersionOf(&testing2.Mockplugina101)
	if err != nil {
		t.Error(err)
	}

	if result {
		t.Errorf("Expected no newer version than %s", testing2.Mockplugina101.Version)
	}
}

func TestPluginsHelperImpl_FilterPlugins(t *testing.T) {
	var h PluginsHelperImpl

	installed, updated := h.FilterPlugins(&testing2.MockGrafana, testing2.MockPluginList)

	if !updated {
		t.Errorf("Expected plugins to be installed")
	}

	result := true
	result = result && installed.HasExactVersionOf(&testing2.Mockplugina101)
	result = result && installed.HasExactVersionOf(&testing2.Mockpluginb100)
	result = result && (len(installed) == 2)

	if !result {
		t.Errorf("Unexpected plugins got installed")
	}
}
