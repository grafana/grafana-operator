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
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Dashboard Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaDashboardSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecSuspended,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecNoMatchingInstances,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecApplyFailed,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionDashboardSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: LogMsgApplyErrors,
		},
		{
			name: "Invalid JSON",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  commonSpecInvalidSpec,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{]"}, // Invalid json
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidModelResolution,
			},
			wantErr: "resolving dashboard contents",
		},
		{
			name: "No model can be resolved, no model source is defined",
			meta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "invalid-spec-no-model-source",
			},
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonInvalidModelResolution,
			},
			wantErr: "resolving dashboard contents",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JSON: `{
							"title": "Minimal Dashboard",
							"links": []
						}`,
				},
			},
			want: metav1.Condition{
				Type:   conditionDashboardSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
		{
			name: "Successfully applied shared dashboard",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JSON: `{
							"title": "Minimal Dashboard With sharing",
							"links": []
						}`,
				},
				PublicSharing: &v1beta1.GrafanaDashboardPublicSharing{},
			},
			want: metav1.Condition{
				Type:   conditionDashboardSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaDashboard{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaDashboardReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("Dashboard Reconciler", Ordered, func() {
	t := GinkgoT()

	const (
		uid       = "url-based-dashboard"
		title1    = "title1"
		title2    = "title2"
		endpoint1 = "/endpoint1"
		endpoint2 = "/endpoint2"
	)

	dash1 := fmt.Sprintf(`{ "title": "%s", "uid": "%s", "links": [] }`, title1, uid)
	dash2 := fmt.Sprintf(`{ "title": "%s", "uid": "%s", "links": [] }`, title2, uid)

	data := map[string]string{
		endpoint1: dash1,
		endpoint2: dash2,
	}

	mux := tk8s.GetJSONmux(t, data)

	ts := httptest.NewServer(mux)

	AfterAll(func() {
		ts.Close()
	})

	It("updates dashboard in Grafana upon .spec.url change", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "url-based-dashboard",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: externalGrafanaCr.GetLabels(),
					},
				},
			},
		}

		key := types.NamespacedName{
			Namespace: cr.Namespace,
			Name:      cr.Name,
		}

		r := &GrafanaDashboardReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, cr)

		// First revision
		cr.Spec.URL = ts.URL + endpoint1

		err = cl.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err := gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, dash.String(), title1)

		// Second revision
		cr = &v1beta1.GrafanaDashboard{}
		err = cl.Get(testCtx, key, cr)
		require.NoError(t, err)

		cr.Spec.URL = ts.URL + endpoint2

		err = cl.Update(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err = gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.NotContains(t, dash.String(), title1)
		assert.Contains(t, dash.String(), title2)

		// Cleanup
		err = cl.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})

	It("mitigates dashboard drift when it occurs", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "url-based-dashboard-drift",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: externalGrafanaCr.GetLabels(),
					},
				},
			},
		}

		// Make it long enough, so we can play with reconciliation
		cr.Spec.ResyncPeriod.Duration = 5 * time.Minute

		r := &GrafanaDashboardReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, cr)

		// Create dashboard
		cr.Spec.URL = ts.URL + endpoint1

		err = cl.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err := gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, dash.String(), title1)

		// Modify the dashboard to simulate remote drift
		model, ok := dash.GetPayload().Dashboard.(map[string]any)
		require.True(t, ok)

		model["title"] = title2

		_, err = gClient.Dashboards.PostDashboard( //nolint:errcheck
			&models.SaveDashboardCommand{
				Dashboard: model,
				FolderUID: dash.Payload.Meta.FolderUID,
				Overwrite: true,
			})
		require.NoError(t, err)

		dash, err = gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, dash.String(), title2) // Make sure the drift is there

		// Reconcile again to fix the drift
		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err = gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, dash.String(), title1) // Make sure the drift is gone now

		// Cleanup
		err = cl.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})
})

var _ = Describe("Dashboard Reconciler", Ordered, func() {
	t := GinkgoT()

	const uid = "has-public-dashboard"

	It("reconciles drift of publicDashboard", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "updates-shared-dashboard",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: externalGrafanaCr.GetLabels(),
					},
				},
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					CustomUID: uid,
					JSON:      `{ "title": "title", "links": [] }`,
				},
				PublicSharing: &v1beta1.GrafanaDashboardPublicSharing{},
			},
		}

		r := &GrafanaDashboardReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, cr)

		// Create dashboard
		err = cl.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		// Create dashboard
		err = cl.Get(testCtx, req.NamespacedName, cr)
		require.NoError(t, err)

		dash, err := gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)
		assert.NotNil(t, dash)

		remoteDTO, err := gClient.Dashboards.GetPublicDashboard(uid)
		require.NoError(t, err)
		assert.NotNil(t, remoteDTO)
		assert.NotNil(t, remoteDTO.Payload)

		expectedDTO := &models.PublicDashboardDTO{
			UID:                  string(cr.UID),
			AccessToken:          string(cr.UID),
			IsEnabled:            new(true),
			TimeSelectionEnabled: new(false),
			AnnotationsEnabled:   new(false),
		}

		matches, recreate := r.publicSharingMatchesStateInGrafana(cr, expectedDTO, remoteDTO)
		assert.True(t, matches)
		assert.False(t, recreate)

		// Invert all booleans and update ignored field UID
		_, err = gClient.Dashboards.UpdatePublicDashboard(&dashboards.UpdatePublicDashboardParams{ //nolint:errcheck
			Body: &models.PublicDashboardDTO{
				UID:                  "ignored-by-grafana",
				IsEnabled:            new(false),
				AnnotationsEnabled:   new(true),
				TimeSelectionEnabled: new(true),
			},
			DashboardUID: uid,
			UID:          string(cr.UID),
		})
		require.NoError(t, err)

		// Make sure the drift is there
		remoteDTO, err = gClient.Dashboards.GetPublicDashboard(uid)
		require.NoError(t, err)
		assert.NotNil(t, remoteDTO)
		assert.NotNil(t, remoteDTO.Payload)

		// Should not match
		// Recreate is unnecessary as only mutable fields changed (UID ignored).
		matches, recreate = r.publicSharingMatchesStateInGrafana(cr, expectedDTO, remoteDTO)
		assert.False(t, matches)
		assert.False(t, recreate)

		// Reconcile again to fix the drift
		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		// Make sure the drift is gone now
		remoteDTO, err = gClient.Dashboards.GetPublicDashboard(uid)
		require.NoError(t, err)
		assert.NotNil(t, remoteDTO)
		assert.NotNil(t, remoteDTO.Payload)

		matches, _ = r.publicSharingMatchesStateInGrafana(cr, expectedDTO, remoteDTO)
		assert.True(t, matches)
		assert.False(t, recreate)

		// Cleanup
		err = cl.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})
})

func TestGrafanaDashboardReconcilerMatchesStateInGrafana(t *testing.T) {
	const uid = "myuid"

	tests := []struct {
		name   string
		exists bool
		title1 string
		title2 string
		want   bool
	}{
		{
			name:   "doesn't exist",
			exists: false,
			title1: "title",
			title2: "title",
			want:   false,
		},
		{
			name:   "remote drift",
			exists: true,
			title1: "title",
			title2: "different-title",
			want:   false,
		},
		{
			name:   "no drift",
			exists: true,
			title1: "title",
			title2: "title",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dash1 := map[string]any{
				"uid":   uid,
				"title": tt.title1,
			}

			dash2 := &dashboards.GetDashboardByUIDOK{
				Payload: &models.DashboardFullWithMeta{
					Dashboard: map[string]any{
						"uid":   uid,
						"title": tt.title2,
					},
				},
			}

			r := &GrafanaDashboardReconciler{}

			got, err := r.matchesStateInGrafana(tt.exists, dash1, dash2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("remote dashboard is not a valid object", func(t *testing.T) {
		dash1 := map[string]any{
			"uid":   uid,
			"title": "title",
		}

		dash2 := &dashboards.GetDashboardByUIDOK{
			Payload: &models.DashboardFullWithMeta{
				Dashboard: nil,
			},
		}

		r := &GrafanaDashboardReconciler{}

		got, err := r.matchesStateInGrafana(true, dash1, dash2)
		require.ErrorContains(t, err, "remote dashboard is not a valid object")

		assert.False(t, got)
	})
}

func TestGrafanaDashboardReconcilerPublicDashboardMatchesStateInGrafana(t *testing.T) {
	uid := uuid.NewString()

	defaultPublicDashboard := dashboards.GetPublicDashboardOK{
		Payload: &models.PublicDashboard{
			UID:                  uid,
			AccessToken:          uid,
			IsEnabled:            true,
			AnnotationsEnabled:   false,
			TimeSelectionEnabled: false,
		},
	}

	tests := []struct {
		name         string
		changes      *models.PublicDashboardDTO
		annotations  map[string]string
		wantMatch    bool
		wantRecreate bool
	}{
		{
			name: "no drift",
			changes: &models.PublicDashboardDTO{
				UID:                  uid,
				AccessToken:          uid,
				AnnotationsEnabled:   new(false),
				IsEnabled:            new(true),
				TimeSelectionEnabled: new(false),
			},
			wantMatch:    true,
			wantRecreate: false,
		},
		{
			name: "mutable fields trigger update",
			changes: &models.PublicDashboardDTO{
				UID:                  uid,
				AccessToken:          uid,
				IsEnabled:            new(false),
				AnnotationsEnabled:   new(true),
				TimeSelectionEnabled: new(true),
			},
			wantMatch:    false,
			wantRecreate: false,
		},
		{
			// In case the cr is recreated without running the finalizer or EtcD state is lost, we need to recover
			name: "UID changes should recreate",
			changes: &models.PublicDashboardDTO{
				UID:         "changed-uid",
				AccessToken: uid,
			},
			wantMatch:    false,
			wantRecreate: true,
		},
		{
			name: "AccessToken changes should recreate",
			changes: &models.PublicDashboardDTO{
				UID:         uid,
				AccessToken: uuid.NewString(),
			},
			wantMatch:    false,
			wantRecreate: true,
		},
		{
			name: "status annotation is empty",
			changes: &models.PublicDashboardDTO{
				UID:         uid,
				AccessToken: uid,
			},
			annotations:  map[string]string{annotationSyncedPublicSharing: ""},
			wantMatch:    false,
			wantRecreate: true,
		},
		{
			name: "status annotation does not match",
			changes: &models.PublicDashboardDTO{
				UID:         uid,
				AccessToken: uid,
			},
			annotations:  map[string]string{annotationSyncedPublicSharing: uuid.NewString()},
			wantMatch:    false,
			wantRecreate: true,
		},
		{
			name: "status annotation matches uid",
			changes: &models.PublicDashboardDTO{
				UID:                  uid,
				AccessToken:          uid,
				IsEnabled:            new(true),
				AnnotationsEnabled:   new(false),
				TimeSelectionEnabled: new(false),
			},
			annotations:  map[string]string{annotationSyncedPublicSharing: uid},
			wantMatch:    true,
			wantRecreate: false,
		},
	}

	cr := &v1beta1.GrafanaDashboard{}
	r := &GrafanaDashboardReconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr.Annotations = tt.annotations

			matches, recreate := r.publicSharingMatchesStateInGrafana(cr, tt.changes, &defaultPublicDashboard)
			assert.Equal(t, tt.wantMatch, matches, "'matches' did not match 'wantMatches'")
			assert.Equal(t, tt.wantRecreate, recreate, "'recreate' did not match 'wantRecreate'")
		})
	}

	t.Run("remote model is nil", func(t *testing.T) {
		matches, recreate := r.publicSharingMatchesStateInGrafana(cr, &models.PublicDashboardDTO{UID: uid, AccessToken: uid}, nil)
		assert.False(t, matches)
		assert.True(t, recreate)
	})
}
