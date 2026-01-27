package controllers

import (
	"fmt"
	"net/http/httptest"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("LibraryPanel Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaLibraryPanelSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaLibraryPanelSpec{
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
			spec: v1beta1.GrafanaLibraryPanelSpec{
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
			spec: v1beta1.GrafanaLibraryPanelSpec{
				GrafanaCommonSpec:  commonSpecApplyFailed,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
			want: metav1.Condition{
				Type:   conditionLibraryPanelSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: LogMsgApplyErrors,
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaLibraryPanelSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JSON: `{
							"uid": "do-adhv-ank",
							"name": "API docs Example",
							"type": "text",
							"model": {},
							"version": 1
						}`,
				},
			},
			want: metav1.Condition{
				Type:   conditionLibraryPanelSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaLibraryPanel{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaLibraryPanelReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("LibraryPanel Reconciler", Ordered, func() {
	t := GinkgoT()

	const (
		uid       = "url-based-library-panel"
		name1     = "name1"
		name2     = "name2"
		endpoint1 = "/endpoint1"
		endpoint2 = "/endpoint2"
	)

	panel1 := fmt.Sprintf(`{ "name": "%s", "uid": "%s", "type": "text", "model": {} }`, name1, uid)
	panel2 := fmt.Sprintf(`{ "name": "%s", "uid": "%s", "type": "text", "model": {} }`, name2, uid)

	data := map[string]string{
		endpoint1: panel1,
		endpoint2: panel2,
	}

	mux := tk8s.GetJSONmux(t, data)

	ts := httptest.NewServer(mux)
	AfterAll(func() {
		ts.Close()
	})

	It("updates librarypanel in Grafana upon .spec.url change", func() {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		cr := &v1beta1.GrafanaLibraryPanel{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "url-based-library-panel",
			},
			Spec: v1beta1.GrafanaLibraryPanelSpec{
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

		r := &GrafanaLibraryPanelReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, cr)

		// First revision
		cr.Spec.URL = ts.URL + endpoint1

		err = cl.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		panel, err := gClient.LibraryElements.GetLibraryElementByUID(uid)
		require.NoError(t, err)

		assert.Contains(t, panel.String(), name1)

		// Second revision
		cr = &v1beta1.GrafanaLibraryPanel{}
		err = cl.Get(testCtx, key, cr)
		require.NoError(t, err)

		cr.Spec.URL = ts.URL + endpoint2

		err = cl.Update(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)

		panel, err = gClient.LibraryElements.GetLibraryElementByUID(uid)
		require.NoError(t, err)

		assert.NotContains(t, panel.String(), name1)
		assert.Contains(t, panel.String(), name2)

		// Cleanup
		err = cl.Delete(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
	})
})
