package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

func TestRemoveMissingCRs(t *testing.T) {
	statusList := v1beta1.NamespacedResourceList{
		"default/present/uid",
		"default/missing/uid",
		"other/missing/uid",
	}

	dashboards := v1beta1.GrafanaDashboardList{
		Items: []v1beta1.GrafanaDashboard{
			{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "present"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "unrelated-dashboard"},
			},
		},
	}

	// Sanity checks before test
	assert.Len(t, statusList, 3)
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/present/uid"))
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/missing/uid"))
	assert.Contains(t, statusList, v1beta1.NamespacedResource("other/missing/uid"))

	updateStatus := false
	removeMissingCRs(&statusList, &dashboards, &updateStatus)

	assert.True(t, updateStatus, "Entries were removed but status change was not detected")

	assert.Len(t, statusList, 1)
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/present/uid"))
	assert.NotContains(t, statusList, v1beta1.NamespacedResource("default/missing/uid"))
	assert.NotContains(t, statusList, v1beta1.NamespacedResource("other/missing/uid"))

	found, _ := statusList.Find("default", "unrelated-dashboard")
	assert.False(t, found, "Dashboard is not in status and should not be")
}

var _ = Describe("Grafana Reconciler: Provoke Conditions", func() {
	t := GinkgoT()

	tests := []struct {
		name string
		meta metav1.ObjectMeta
		spec v1beta1.GrafanaSpec
		want metav1.Condition
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaSpec{
				Suspend: true,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonReconcileSuspended,
			},
		},
		// TODO When InvalidSpec is implemented for external instances admin secret referencing a non-existing secret
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.Grafana{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			err := cl.Create(testCtx, cr)
			require.NoError(t, err)

			r := GrafanaReconciler{Client: cl, Scheme: cl.Scheme()}
			req := tk8s.GetRequest(t, cr)

			_, err = r.Reconcile(testCtx, req)
			require.NoError(t, err)

			cr = &v1beta1.Grafana{}

			err = r.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			hasCondition := tk8s.HasCondition(t, cr, tt.want)
			assert.True(t, hasCondition)
		})
	}
})
