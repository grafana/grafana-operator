package content

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDashboardEnvs(t *testing.T) {
	cl := tk8s.GetFakeClient(t)

	dashboard := v1beta1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dashboard",
			Namespace: "grafana-operator-system",
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaContentSpec: v1beta1.GrafanaContentSpec{
				Envs: []v1beta1.GrafanaContentEnv{
					{
						Name:  "TEST_ENV",
						Value: "test-env-value",
					},
				},
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	var contentResource v1beta1.GrafanaContentResource = &dashboard
	assert.NotNil(t, contentResource.GrafanaContentSpec(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")
	assert.NotNil(t, contentResource.GrafanaContentStatus(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")

	resolver := NewResolver(&dashboard, cl)

	envs, err := resolver.getContentEnvs(ctx)

	require.NoError(t, err)
	assert.NotNil(t, envs)
	assert.Len(t, envs, 1)
}

func resolverWithVariables(variables ...v1beta1.GrafanaContentVariable) *Resolver {
	dashboard := &v1beta1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{Name: "test-dashboard", Namespace: "grafana-operator-system"},
		Spec:       v1beta1.GrafanaDashboardSpec{Variables: variables},
	}

	return NewResolver(dashboard, nil)
}

// overrideVariablesIn parses model and applies the overrides declared by resolver to it.
func overrideVariablesIn(t *testing.T, resolver *Resolver, model string) map[string]any {
	t.Helper()

	var parsed map[string]any

	require.NoError(t, json.Unmarshal([]byte(model), &parsed))

	overridden, err := resolver.overrideVariables(parsed)
	require.NoError(t, err)

	return overridden
}

func asMap(t *testing.T, v any) map[string]any {
	t.Helper()

	m, ok := v.(map[string]any)
	require.True(t, ok, "expected a JSON object, got %T", v)

	return m
}

func asSlice(t *testing.T, v any) []any {
	t.Helper()

	s, ok := v.([]any)
	require.True(t, ok, "expected a JSON array, got %T", v)

	return s
}

// templatingVar navigates model["templating"]["list"] and returns the entry with the given name.
func templatingVar(t *testing.T, model map[string]any, name string) map[string]any {
	t.Helper()

	for _, item := range asSlice(t, asMap(t, model["templating"])["list"]) {
		v := asMap(t, item)
		if v["name"] == name {
			return v
		}
	}

	t.Fatalf("variable %q not found", name)

	return nil
}

func TestResolveVariables(t *testing.T) {
	t.Run("overrides a datasource variable without options", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "datasource", Value: "prometheus-uid"},
		), `{
			"templating": {"list": [
				{"name": "datasource", "type": "datasource", "current": {"text": "Old", "value": "old-uid"}}
			]}
		}`)

		current := asMap(t, templatingVar(t, model, "datasource")["current"])
		assert.Equal(t, "prometheus-uid", current["value"])
		assert.Equal(t, "prometheus-uid", current["text"])
		assert.Equal(t, true, current["selected"])
	})

	t.Run("reconciles options when the value already exists", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "default"},
		), `{
			"templating": {"list": [
				{"name": "namespace", "type": "query",
				 "current": {"text": "kube-system", "value": "kube-system", "selected": true},
				 "options": [
					{"text": "kube-system", "value": "kube-system", "selected": true},
					{"text": "default", "value": "default", "selected": false}
				]}
			]}
		}`)

		options := asSlice(t, templatingVar(t, model, "namespace")["options"])
		require.Len(t, options, 2)

		for _, opt := range options {
			option := asMap(t, opt)
			if option["value"] == "default" {
				assert.Equal(t, true, option["selected"])
			} else {
				assert.Equal(t, false, option["selected"])
			}
		}
	})

	t.Run("synthesizes an option when the value is not present", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "cluster", Value: "b"},
		), `{
			"templating": {"list": [
				{"name": "cluster", "type": "custom",
				 "current": {"text": "a", "value": "a", "selected": true},
				 "options": [{"text": "a", "value": "a", "selected": true}]}
			]}
		}`)

		options := asSlice(t, templatingVar(t, model, "cluster")["options"])
		require.Len(t, options, 2)

		first := asMap(t, options[0])
		assert.Equal(t, "b", first["value"])
		assert.Equal(t, true, first["selected"])
		assert.Equal(t, false, asMap(t, options[1])["selected"])
	})

	t.Run("ignores a variable that is not present in the model", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "does-not-exist", Value: "x"},
		), `{
			"templating": {"list": [
				{"name": "namespace", "type": "query", "current": {"text": "default", "value": "default"}}
			]}
		}`)

		current := asMap(t, templatingVar(t, model, "namespace")["current"])
		assert.Equal(t, "default", current["value"])
	})

	t.Run("no-op when templating block is absent", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "x"},
		), `{"title": "no templating"}`)

		assert.Equal(t, "no templating", model["title"])
	})

	t.Run("no-op when the variable list is not shaped like one", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "x"},
		), `{"templating": {"list": "not-a-list"}}`)

		assert.Equal(t, "not-a-list", asMap(t, model["templating"])["list"])
	})

	t.Run("applies multiple variables in a single pass", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "datasource", Value: "prometheus-uid"},
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "default"},
		), `{
			"templating": {"list": [
				{"name": "datasource", "type": "datasource", "current": {"text": "Old", "value": "old-uid"}},
				{"name": "namespace", "type": "query", "current": {"text": "kube-system", "value": "kube-system"}}
			]}
		}`)

		assert.Equal(t, "prometheus-uid", asMap(t, templatingVar(t, model, "datasource")["current"])["value"])
		assert.Equal(t, "default", asMap(t, templatingVar(t, model, "namespace")["current"])["value"])
	})

	t.Run("last duplicate override wins", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "first"},
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "second"},
		), `{
			"templating": {"list": [
				{"name": "namespace", "type": "query", "current": {"text": "kube-system", "value": "kube-system"}}
			]}
		}`)

		assert.Equal(t, "second", asMap(t, templatingVar(t, model, "namespace")["current"])["value"])
	})

	t.Run("collapses a multi-value variable to the single value", func(t *testing.T) {
		model := overrideVariablesIn(t, resolverWithVariables(
			v1beta1.GrafanaContentVariable{Name: "namespace", Value: "default"},
		), `{
			"templating": {"list": [
				{"name": "namespace", "type": "query", "multi": true,
				 "current": {"selected": true, "text": ["a", "b"], "value": ["a", "b"]}}
			]}
		}`)

		current := asMap(t, templatingVar(t, model, "namespace")["current"])
		assert.Equal(t, "default", current["value"])
		assert.Equal(t, "default", current["text"])
	})

	t.Run("leaves content resources without variables untouched", func(t *testing.T) {
		panel := &v1beta1.GrafanaLibraryPanel{
			ObjectMeta: metav1.ObjectMeta{Name: "test-panel", Namespace: "grafana-operator-system"},
		}

		var contentResource v1beta1.GrafanaContentResource = panel
		assert.NotImplements(t, (*v1beta1.GrafanaContentVariableOverrider)(nil), contentResource,
			"variable overrides only apply to the dashboard model")

		model := overrideVariablesIn(t, NewResolver(panel, nil), `{"title": "panel"}`)
		assert.Equal(t, "panel", model["title"])
	})
}

func TestGetContentModelVariablesAffectHash(t *testing.T) {
	contentJSON := []byte(`{
		"templating": {"list": [
			{"name": "datasource", "type": "datasource", "current": {"text": "Old", "value": "old-uid"}}
		]}
	}`)

	_, baseHash, err := resolverWithVariables().getContentModel(contentJSON)
	require.NoError(t, err)

	model, overrideHash, err := resolverWithVariables(
		v1beta1.GrafanaContentVariable{Name: "datasource", Value: "prometheus-uid"},
	).getContentModel(contentJSON)
	require.NoError(t, err)

	assert.NotEqual(t, baseHash, overrideHash, "overriding a variable should change the content hash")

	current := asMap(t, templatingVar(t, model, "datasource")["current"])
	assert.Equal(t, "prometheus-uid", current["value"])
}
