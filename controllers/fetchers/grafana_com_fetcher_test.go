package fetchers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestFetchDashboardFromGrafanaCom(t *testing.T) {
	dashboard := &v1beta1.GrafanaDashboard{
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaCom: &v1beta1.GrafanaComDashboardReference{
				Id: 1860,
			},
		},
		Status: v1beta1.GrafanaDashboardStatus{},
	}

	fetchedDashboard, err := FetchDashboardFromGrafanaCom(dashboard)
	assert.Nil(t, err)
	assert.NotNil(t, fetchedDashboard, "Fetched dashboard shouldn't be empty")
	assert.GreaterOrEqual(t, *dashboard.Spec.GrafanaCom.Revision, 30, "At least 30 revisions exist for dashboard 1860 as of 2023-03-29")

	assert.False(t, dashboard.Status.ContentTimestamp.Time.IsZero(), "ContentTimestamp should have been set")
	assert.NotEmpty(t, dashboard.Status.ContentUrl, "ContentUrl should have been set")
}
