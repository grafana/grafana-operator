package fetchers

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchDashboardFromGrafanaCom(t *testing.T) {
	cl := tk8s.GetFakeClient(t)

	dashboard := &v1beta1.GrafanaDashboard{
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaContentSpec: v1beta1.GrafanaContentSpec{
				GrafanaCom: &v1beta1.GrafanaComContentReference{
					ID: 1860,
				},
			},
		},
		Status: v1beta1.GrafanaDashboardStatus{},
	}

	fetchedDashboard, err := FetchFromGrafanaCom(context.Background(), dashboard, cl)
	require.NoError(t, err)
	assert.NotNil(t, fetchedDashboard, "Fetched dashboard shouldn't be empty")
	assert.GreaterOrEqual(t, *dashboard.Spec.GrafanaCom.Revision, 42, "At least 42 revisions exist for dashboard 1860 as of 2024-12-22")

	assert.False(t, dashboard.Status.ContentTimestamp.Time.IsZero(), "ContentTimestamp should have been set")
	assert.NotEmpty(t, dashboard.Status.ContentURL, "ContentURL should have been set")
}
