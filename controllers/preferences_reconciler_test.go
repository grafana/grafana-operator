/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	grafanareconciler "github.com/grafana/grafana-operator/v5/controllers/reconcilers/grafana"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
)

// Unit tests for the indexer + watch map function. These don't need the Grafana
// testcontainer.
func TestPreferencesIndexAndMap(t *testing.T) {
	t.Run("indexHomeDashboardUID returns nil when preferences unset", func(t *testing.T) {
		r := &GrafanaReconciler{}
		idx := r.indexHomeDashboardUID()

		got := idx(&v1beta1.Grafana{})
		assert.Nil(t, got)
	})

	t.Run("indexHomeDashboardUID returns nil when home dashboard UID empty", func(t *testing.T) {
		r := &GrafanaReconciler{}
		idx := r.indexHomeDashboardUID()

		got := idx(&v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				Preferences: &v1beta1.GrafanaPreferences{},
			},
		})
		assert.Nil(t, got)
	})

	t.Run("indexHomeDashboardUID indexes the configured UID", func(t *testing.T) {
		r := &GrafanaReconciler{}
		idx := r.indexHomeDashboardUID()

		got := idx(&v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				Preferences: &v1beta1.GrafanaPreferences{HomeDashboardUID: "abc-123"},
			},
		})
		assert.Equal(t, []string{"abc-123"}, got)
	})

	t.Run("requestsForDashboardHomeRef ignores dashboards without a resolved status.UID", func(t *testing.T) {
		fakeClient := fake.NewClientBuilder().Build()
		r := &GrafanaReconciler{Client: fakeClient}

		mapFn := r.requestsForDashboardHomeRef("ignored-index-key")
		dash := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "dash"},
		}

		got := mapFn(context.Background(), dash)
		assert.Empty(t, got)
	})

	t.Run("requestsForDashboardHomeRef enqueues only Grafanas matching by index", func(t *testing.T) {
		matching := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns-a", Name: "matching"},
			Spec: v1beta1.GrafanaSpec{
				Preferences: &v1beta1.GrafanaPreferences{HomeDashboardUID: "home-uid"},
			},
		}
		other := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns-b", Name: "other"},
			Spec: v1beta1.GrafanaSpec{
				Preferences: &v1beta1.GrafanaPreferences{HomeDashboardUID: "different-uid"},
			},
		}

		const indexKey = ".spec.preferences.homeDashboardUID"

		fakeClient := fake.NewClientBuilder().
			WithScheme(testScheme()).
			WithObjects(matching, other).
			WithIndex(&v1beta1.Grafana{}, indexKey, func(o client.Object) []string {
				g, ok := o.(*v1beta1.Grafana)
				if !ok || g.Spec.Preferences == nil {
					return nil
				}

				return []string{g.Spec.Preferences.HomeDashboardUID}
			}).
			Build()
		r := &GrafanaReconciler{Client: fakeClient}

		mapFn := r.requestsForDashboardHomeRef(indexKey)
		dash := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "dash"},
			Status: v1beta1.GrafanaDashboardStatus{
				GrafanaContentStatus: v1beta1.GrafanaContentStatus{UID: "home-uid"},
			},
		}

		got := mapFn(context.Background(), dash)
		assert.Equal(t, []reconcile.Request{
			{NamespacedName: types.NamespacedName{Namespace: "ns-a", Name: "matching"}},
		}, got)
	})
}

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	if err := v1beta1.AddToScheme(s); err != nil {
		panic(err)
	}

	return s
}

// Integration tests against the real Grafana running in a testcontainer
// (started by suite_test.go). Each test creates a Grafana CR pointing to the
// container and invokes the reconciler directly.
var _ = Describe("Preferences Reconciler", Ordered, func() {
	t := GinkgoT()

	const (
		homeDashUID   = "preferences-home-dash"
		dashTitle     = "Home Dashboard For Preferences Test"
		grafanaCRName = "preferences-grafana"
		grafanaNS     = "default"
		dashboardCR   = "preferences-home-dashboard"
		labelKey      = "preferences-test"
		labelValue    = "true"
	)

	matchLabels := map[string]string{labelKey: labelValue}

	var (
		grafanaCR *v1beta1.Grafana
		gReq      = reconcile.Request{NamespacedName: types.NamespacedName{Namespace: grafanaNS, Name: grafanaCRName}}
	)

	BeforeAll(func() {
		grafanaCR = &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: grafanaNS,
				Name:      grafanaCRName,
				Labels:    matchLabels,
			},
			Spec: v1beta1.GrafanaSpec{
				External: externalGrafanaCr.Spec.External,
				Config:   externalGrafanaCr.Spec.Config,
				Client:   externalGrafanaCr.Spec.Client,
				// Preferences set/cleared per-test via Patch
			},
		}

		err := cl.Create(testCtx, grafanaCR)
		require.NoError(t, err)
	})

	AfterAll(func() {
		// Clean up the Grafana CR. Dashboard cleanup happens inline.
		require.NoError(t, client.IgnoreNotFound(cl.Delete(testCtx, grafanaCR)))
	})

	It("stays Ready and marks PreferencesApplied=False when home dashboard is missing", func() {
		// Set spec.preferences to a UID that doesn't yet exist in Grafana.
		fresh := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, fresh))

		fresh.Spec.Preferences = &v1beta1.GrafanaPreferences{HomeDashboardUID: "definitely-does-not-exist"}
		require.NoError(t, cl.Update(testCtx, fresh))

		r := &GrafanaReconciler{Client: cl, Scheme: cl.Scheme()}
		_, err := r.Reconcile(testCtx, gReq)
		require.NoError(t, err, "missing home dashboard should not hard-fail the reconcile")

		got := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, got))

		assert.True(t, tk8s.HasCondition(t, got, metav1.Condition{
			Type:   conditionTypeGrafanaReady,
			Reason: "GrafanaReady",
		}), "Grafana should still report Ready when home dashboard is missing")

		assert.True(t, tk8s.HasCondition(t, got, metav1.Condition{
			Type:   grafanareconciler.ConditionPreferencesApplied,
			Reason: "HomeDashboardMissing",
		}), "PreferencesApplied condition should reflect HomeDashboardMissing")
	})

	It("applies the home dashboard preference once the dashboard exists in Grafana", func() {
		// Create a GrafanaDashboard CR matching grafanaCR's labels and reconcile it,
		// which creates the dashboard in Grafana.
		dash := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: grafanaNS,
				Name:      dashboardCR,
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{MatchLabels: matchLabels},
				},
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JSON: fmt.Sprintf(`{"uid": %q, "title": %q, "links": []}`, homeDashUID, dashTitle),
				},
			},
		}
		require.NoError(t, cl.Create(testCtx, dash))

		dr := &GrafanaDashboardReconciler{Client: cl, Scheme: cl.Scheme(), Cfg: &Config{}}
		_, err := dr.Reconcile(testCtx, tk8s.GetRequest(t, dash))
		require.NoError(t, err)

		// Sanity: dashboard now exists in Grafana. Use externalGrafanaCr (which
		// has Status.AdminURL populated by the suite bootstrap reconcile) since
		// both CRs target the same testcontainer.
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		_, err = gClient.Dashboards.GetDashboardByUID(homeDashUID) //nolint:errcheck
		require.NoError(t, err, "dashboard should exist in Grafana after dashboard reconcile")

		// Point the Grafana CR's preferences at the now-existing dashboard and
		// also exercise the additional preference fields wired through the PATCH.
		fresh := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, fresh))

		const (
			wantOrgName  = "Custom Org"
			wantTheme    = "dark"
			wantTimezone = "UTC"
			wantWeek     = "monday"
			wantLanguage = "en-US"
		)

		fresh.Spec.Preferences = &v1beta1.GrafanaPreferences{
			OrganizationName: wantOrgName,
			HomeDashboardUID: homeDashUID,
			Theme:            wantTheme,
			Timezone:         wantTimezone,
			WeekStart:        wantWeek,
			Language:         wantLanguage,
		}
		require.NoError(t, cl.Update(testCtx, fresh))

		r := &GrafanaReconciler{Client: cl, Scheme: cl.Scheme()}
		_, err = r.Reconcile(testCtx, gReq)
		require.NoError(t, err)

		got := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, got))

		assert.True(t, tk8s.HasCondition(t, got, metav1.Condition{
			Type:   grafanareconciler.ConditionPreferencesApplied,
			Reason: "PreferencesApplied",
		}), "PreferencesApplied condition should flip to PreferencesApplied/True after PATCH succeeds")

		// Verify all preference fields landed in Grafana itself.
		prefs, err := gClient.Org.GetOrgPreferences()
		require.NoError(t, err)

		payload := prefs.GetPayload()
		assert.Equal(t, homeDashUID, payload.HomeDashboardUID, "homeDashboardUID")
		assert.Equal(t, wantTheme, payload.Theme, "theme")
		assert.Equal(t, wantTimezone, payload.Timezone, "timezone")
		assert.Equal(t, wantWeek, payload.WeekStart, "weekStart")
		assert.Equal(t, wantLanguage, payload.Language, "language")

		// Verify the org name was applied via PUT /api/org.
		org, err := gClient.Org.GetCurrentOrg()
		require.NoError(t, err)
		assert.Equal(t, wantOrgName, org.GetPayload().Name, "organization name")

		// Cleanup. Reset the org name so the shared testcontainer Grafana
		// looks the way other specs expect.
		require.NoError(t, cl.Delete(testCtx, dash))

		_, err = dr.Reconcile(testCtx, tk8s.GetRequest(t, dash))
		require.NoError(t, err)

		_, err = gClient.Org.UpdateCurrentOrg(&models.UpdateOrgForm{Name: "Main Org."}) //nolint:errcheck
		require.NoError(t, err)
	})

	It("clears the PreferencesApplied condition when spec.preferences is unset", func() {
		fresh := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, fresh))

		fresh.Spec.Preferences = nil
		require.NoError(t, cl.Update(testCtx, fresh))

		r := &GrafanaReconciler{Client: cl, Scheme: cl.Scheme()}
		_, err := r.Reconcile(testCtx, gReq)
		require.NoError(t, err)

		got := &v1beta1.Grafana{}
		require.NoError(t, cl.Get(testCtx, gReq.NamespacedName, got))

		assert.False(t, tk8s.HasCondition(t, got, metav1.Condition{
			Type:   grafanareconciler.ConditionPreferencesApplied,
			Reason: "HomeDashboardMissing",
		}), "PreferencesApplied condition should be removed when preferences are unset")
		assert.False(t, tk8s.HasCondition(t, got, metav1.Condition{
			Type:   grafanareconciler.ConditionPreferencesApplied,
			Reason: "PreferencesApplied",
		}), "PreferencesApplied condition should be removed when preferences are unset")
	})
})
