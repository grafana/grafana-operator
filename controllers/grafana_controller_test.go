package controllers

import (
	"testing"

	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
