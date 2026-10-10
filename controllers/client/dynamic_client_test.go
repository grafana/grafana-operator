package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// desiredDashboard is a dashboard as a GrafanaManifest template describes it.
func desiredDashboard() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "dashboard.grafana.app/v2",
		"kind":       "Dashboard",
		"metadata": map[string]any{
			"name":        "availability",
			"namespace":   "default",
			"labels":      map[string]any{"team": "sre"},
			"annotations": map[string]any{"grafana.app/folder": "folder-uid"},
		},
		"spec": map[string]any{
			"title": "Availability",
			"steps": []any{
				map[string]any{"color": "green"},
				map[string]any{"color": "red", "value": int64(80)},
			},
		},
	}}
}

// storedDashboard is the same dashboard as Grafana returns it: with a resource version, the
// labels and annotations Grafana adds, a status, and the base step's value stored as null.
func storedDashboard() *unstructured.Unstructured {
	stored := desiredDashboard().DeepCopy()
	stored.SetResourceVersion("42")
	stored.SetLabels(map[string]string{"team": "sre", "grafana.app/deprecatedInternalID": "123"})
	stored.SetAnnotations(map[string]string{
		"grafana.app/folder":           "folder-uid",
		"grafana.app/createdBy":        "user:abc",
		"grafana.app/updatedTimestamp": "2026-10-08T10:21:08Z",
	})
	stored.Object["status"] = map[string]any{"conversion": map[string]any{"failed": false}}
	stored.Object["spec"].(map[string]any)["steps"].([]any)[0].(map[string]any)["value"] = nil //nolint:errcheck,forcetypeassert

	return stored
}

func TestMatchesExisting(t *testing.T) {
	spec := func(u *unstructured.Unstructured) map[string]any {
		return u.Object["spec"].(map[string]any) //nolint:errcheck,forcetypeassert
	}

	tests := []struct {
		name    string
		desired func() *unstructured.Unstructured
		stored  func() *unstructured.Unstructured
		want    bool
	}{
		{
			name:    "unchanged, with what Grafana adds of its own",
			desired: desiredDashboard,
			stored:  storedDashboard,
			want:    true,
		},
		{
			name:    "a value of the spec changed",
			desired: func() *unstructured.Unstructured { d := desiredDashboard(); spec(d)["title"] = "Renamed"; return d },
			stored:  storedDashboard,
			want:    false,
		},
		{
			name:    "a key of the spec only Grafana has, which an update removes",
			desired: desiredDashboard,
			stored: func() *unstructured.Unstructured {
				s := storedDashboard()
				spec(s)["description"] = "edited in the UI"

				return s
			},
			want: false,
		},
		{
			name:    "a key of the spec only the template has",
			desired: func() *unstructured.Unstructured { d := desiredDashboard(); spec(d)["description"] = "new"; return d },
			stored:  storedDashboard,
			want:    false,
		},
		{
			name:    "a null in the template is a key Grafana does not have",
			desired: func() *unstructured.Unstructured { d := desiredDashboard(); spec(d)["description"] = nil; return d },
			stored:  storedDashboard,
			want:    true,
		},
		{
			name:    "a number decoded as int64 and the same number from a patch",
			desired: func() *unstructured.Unstructured { d := desiredDashboard(); spec(d)["refresh"] = 30; return d },
			stored:  func() *unstructured.Unstructured { s := storedDashboard(); spec(s)["refresh"] = int64(30); return s },
			want:    true,
		},
		{
			name:    "a number decoded as int64 and the same number as float64",
			desired: func() *unstructured.Unstructured { d := desiredDashboard(); spec(d)["refresh"] = float64(30); return d },
			stored:  func() *unstructured.Unstructured { s := storedDashboard(); spec(s)["refresh"] = int64(30); return s },
			want:    true,
		},
		{
			name: "two large numbers that a float64 cannot tell apart",
			desired: func() *unstructured.Unstructured {
				d := desiredDashboard()
				spec(d)["id"] = int64(9007199254740993)

				return d
			},
			stored: func() *unstructured.Unstructured {
				s := storedDashboard()
				spec(s)["id"] = int64(9007199254740992)

				return s
			},
			want: false,
		},
		{
			name: "an element of a list changed",
			desired: func() *unstructured.Unstructured {
				d := desiredDashboard()
				spec(d)["steps"] = []any{map[string]any{"color": "blue"}}

				return d
			},
			stored: storedDashboard,
			want:   false,
		},
		{
			name: "an annotation of the template changed (the dashboard moved to another folder)",
			desired: func() *unstructured.Unstructured {
				d := desiredDashboard()
				d.SetAnnotations(map[string]string{"grafana.app/folder": "other-folder"})

				return d
			},
			stored: storedDashboard,
			want:   false,
		},
		{
			name: "a label of the template is missing in Grafana",
			desired: func() *unstructured.Unstructured {
				d := desiredDashboard()
				d.SetLabels(map[string]string{"team": "sre", "tier": "1"})

				return d
			},
			stored: storedDashboard,
			want:   false,
		},
		{
			name:    "a top-level field other than status only Grafana has",
			desired: desiredDashboard,
			stored:  func() *unstructured.Unstructured { s := storedDashboard(); s.Object["data"] = "x"; return s },
			want:    false,
		},
		{
			name:    "the status is Grafana's own",
			desired: desiredDashboard,
			stored:  func() *unstructured.Unstructured { s := storedDashboard(); s.Object["status"] = "anything"; return s },
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchesExisting(tt.desired(), tt.stored()))
		})
	}
}

func TestApplyUpdatesOnlyWhatChanged(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "dashboard.grafana.app", Version: "v2", Resource: "dashboards"}
	gvrKey := "dashboard.grafana.app/v2.Dashboard"
	gvrCache.Store(gvrKey, gvr)
	t.Cleanup(func() { gvrCache.Delete(gvrKey) })

	fake := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: "DashboardList"}, storedDashboard())
	cl := &DynamicClient{Interface: fake, defaultResourceNamespace: "default"}

	verbs := func() []string {
		actions := fake.Actions()

		out := make([]string, 0, len(actions))
		for _, action := range actions {
			out = append(out, action.GetVerb())
		}

		fake.ClearActions()

		return out
	}

	require.NoError(t, cl.Apply(t.Context(), desiredDashboard()))
	assert.Equal(t, []string{"get"}, verbs(), "an object Grafana already holds is not sent again")

	renamed := desiredDashboard()
	renamed.Object["spec"].(map[string]any)["title"] = "Renamed" //nolint:errcheck,forcetypeassert
	require.NoError(t, cl.Apply(t.Context(), renamed))
	assert.Equal(t, []string{"get", "update"}, verbs(), "a changed object is updated")

	created := desiredDashboard()
	created.SetName("new-dashboard")
	require.NoError(t, cl.Apply(t.Context(), created))
	assert.Equal(t, []string{"get", "create"}, verbs(), "a missing object is created")
}
