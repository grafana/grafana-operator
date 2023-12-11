package fetchers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
)

const grafanaComDashboardApiUrlRoot = "https://grafana.com/api/dashboards"

func FetchDashboardFromGrafanaCom(dashboard *v1beta1.GrafanaDashboard) ([]byte, error) {
	cache := dashboard.GetContentCache()
	if len(cache) > 0 {
		return cache, nil
	}

	source := dashboard.Spec.GrafanaCom

	if source.Revision == nil {
		rev, err := getLatestGrafanaComRevision(dashboard)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest revision for dashboard id %d: %w", source.Id, err)
		}
		source.Revision = &rev
	}

	dashboard.Spec.Url = fmt.Sprintf("%s/%d/revisions/%d/download", grafanaComDashboardApiUrlRoot, source.Id, *source.Revision)

	return FetchDashboardFromUrl(dashboard)
}

func getLatestGrafanaComRevision(dashboard *v1beta1.GrafanaDashboard) (int, error) {
	source := dashboard.Spec.GrafanaCom
	url := fmt.Sprintf("%s/%d/revisions", grafanaComDashboardApiUrlRoot, source.Id)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return -1, err
	}

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.GrafanaComApiRevisionRequests, true)
	response, err := client.RoundTrip(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected status code when requesting revisions, got %v for dashboard %v", response.StatusCode, dashboard.Name)
	}

	type dashboardRevisionItem struct {
		Revision int `json:"revision"`
	}

	type listDashboardRevisionsResponse struct {
		Items []dashboardRevisionItem `json:"items"`
	}

	var listResponse listDashboardRevisionsResponse
	err = json.NewDecoder(response.Body).Decode(&listResponse)
	if err != nil {
		return -1, err
	}

	max := 0
	for _, i := range listResponse.Items {
		if i.Revision > max {
			max = i.Revision
		}
	}

	return max, nil
}
