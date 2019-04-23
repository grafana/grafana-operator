package grafana

import (
	"testing"
)

func TestPluginsList(t *testing.T) {
	result := true
	result = result && MockPluginList.HasSomeVersionOf(&Mockplugina100)
	result = result && MockPluginList.HasSomeVersionOf(&Mockplugina101)
	result = result && MockPluginList.HasSomeVersionOf(&Mockpluginb100)
	result = result && MockPluginList.HasSomeVersionOf(&Mockplugina102)

	if !result {
		t.Errorf("Error in `HasSomeVersionOf`")
	}

	result = result && MockPluginList.HasExactVersionOf(&Mockplugina100)
	result = result && MockPluginList.HasExactVersionOf(&Mockplugina101)
	result = result && MockPluginList.HasExactVersionOf(&Mockpluginb100)
	result = result && (MockPluginList.HasExactVersionOf(&Mockplugina102) == false)
	result = result && (MockPluginList.HasExactVersionOf(&Mockpluginc100) == false)

	if !result {
		t.Errorf("Error in `HasExactVersionOf`")
	}

	result, err := MockPluginList.HasNewerVersionOf(&Mockplugina100)
	if err != nil {
		t.Error(err)
	}

	if !result {
		t.Errorf("Error in `HasNewerVersionOf`")
	}

	result, err = MockPluginList.HasNewerVersionOf(&Mockplugina101)
	if err != nil {
		t.Error(err)
	}

	if result {
		t.Errorf("Error in `HasNewerVersionOf`")
	}
}

func TestPluginsHelperImpl_PickLatestVersions(t *testing.T) {
	var h PluginsHelperImpl

	latestVersions, err := h.PickLatestVersions(MockPluginList)
	if err != nil {
		t.Error(err)
	}

	if latestVersions.HasExactVersionOf(&Mockplugina100) {
		t.Errorf("Expected %s but got %s", Mockplugina101.Version, Mockplugina100.Version)
	}

	result, err := latestVersions.HasNewerVersionOf(&Mockplugina101)
	if err != nil {
		t.Error(err)
	}

	if result {
		t.Errorf("Expected no newer version than %s", Mockplugina101.Version)
	}
}

func TestPluginsHelperImpl_FilterPlugins(t *testing.T) {
	var h PluginsHelperImpl

	MockPluginList.SetOrigin(&MockDashboard)
	installed, updated := h.FilterPlugins(&MockGrafana, MockPluginList)

	if !updated {
		t.Errorf("Expected plugins to be installed")
	}

	result := true
	result = result && installed.HasExactVersionOf(&Mockplugina101)
	result = result && installed.HasExactVersionOf(&Mockpluginb100)
	result = result && (len(MockDashboard.Status.Messages) == 2)
	result = result && (len(installed) == 2)

	if !result {
		t.Errorf("Unexpected plugins got installed")
	}
}
