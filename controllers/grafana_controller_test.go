package controllers

import (
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	tests := []struct {
		name          string
		cr            *v1beta1.Grafana
		wantCondition string
		wantReason    string
	}{
		{
			name: ".spec.suspend=true",
			cr: &v1beta1.Grafana{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaSpec{
					Suspend: true,
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonReconcileSuspended,
		},
		// TODO When InvalidSpec is implemented for external instances admin secret referencing a non-existing secret
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.Grafana{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
