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
	"time"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
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
			wantErr: "failed to apply to all instances",
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
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaDashboard{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaDashboardReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

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

	mux := getJSONmux(
		map[string]string{
			endpoint1: dash1,
			endpoint2: dash2,
		},
	)

	ts := httptest.NewServer(mux)
	AfterAll(func() {
		ts.Close()
	})

	It("updates dashboard in Grafana upon .spec.url change", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
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

		r := &GrafanaDashboardReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		req := requestFromMeta(cr.Metadata())

		// First revision
		cr.Spec.URL = ts.URL + endpoint1

		err = k8sClient.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err := gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, dash.String(), title1)

		// Second revision
		cr = &v1beta1.GrafanaDashboard{}
		err = k8sClient.Get(testCtx, key, cr)
		require.NoError(t, err)

		cr.Spec.URL = ts.URL + endpoint2

		err = k8sClient.Update(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		dash, err = gClient.Dashboards.GetDashboardByUID(uid)
		require.NoError(t, err)

		assert.NotContains(t, dash.String(), title1)
		assert.Contains(t, dash.String(), title2)

		// Cleanup
		err = k8sClient.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})

	It("mitigates dashboard drift when it occurs", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
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

		r := &GrafanaDashboardReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		req := requestFromMeta(cr.Metadata())

		// Create dashboard
		cr.Spec.URL = ts.URL + endpoint1

		err = k8sClient.Create(testCtx, cr)
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
		err = k8sClient.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})
})
