package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

const grafanaComDashboardApiUrlRoot = "https://grafana.com/api/dashboards"

func GetDashboardUrl(ref v1beta1.GrafanaComDashboardReference) string {
	return fmt.Sprintf("%s/%d/revisions/%d/download", grafanaComDashboardApiUrlRoot, ref.Id, ref.Revision)
}

func GetLatestGrafanaComRevision(ctx context.Context, id int) (*int, error) {
	url := fmt.Sprintf("%s/%d/revisions", grafanaComDashboardApiUrlRoot, id)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code when requesting revisions, got %v for dashboard id %d", response, id)
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
		return nil, err
	}

	max := 0
	for _, i := range listResponse.Items {
		if i.Revision > max {
			max = i.Revision
		}
	}

	return &max, nil
}
