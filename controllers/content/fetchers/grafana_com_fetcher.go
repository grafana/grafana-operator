package fetchers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const grafanaComDashboardsAPIEndpoint = "https://grafana.com/api/dashboards"

func FetchFromGrafanaCom(ctx context.Context, cr v1beta1.GrafanaContentResource, c client.Client) ([]byte, error) {
	cache := cache.GetContentCache(cr)
	if len(cache) > 0 {
		return cache, nil
	}

	spec := cr.GrafanaContentSpec()
	if spec == nil {
		return nil, fmt.Errorf("missing content spec definition on resource")
	}

	source := spec.GrafanaCom

	tlsConfig := client2.DefaultTLSConfiguration

	if source.Revision == nil {
		rev, err := getLatestGrafanaComRevision(cr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest revision for dashboard id %d: %w", source.ID, err)
		}

		source.Revision = &rev
	}

	spec.URL = fmt.Sprintf("%s/%d/revisions/%d/download", grafanaComDashboardsAPIEndpoint, source.ID, *source.Revision)

	return FetchFromURL(ctx, cr, c, tlsConfig)
}

func getLatestGrafanaComRevision(cr v1beta1.GrafanaContentResource, tlsConfig *tls.Config) (int, error) {
	spec := cr.GrafanaContentSpec()
	if spec == nil {
		return -1, nil
	}

	source := spec.GrafanaCom
	url := fmt.Sprintf("%s/%d/revisions", grafanaComDashboardsAPIEndpoint, source.ID)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return -1, err
	}

	client := client2.NewInstrumentedRoundTripper(true, tlsConfig, metrics.GrafanaComAPIRevisionRequests.MustCurryWith(prometheus.Labels{
		"kind":     cr.GetObjectKind().GroupVersionKind().Kind,
		"resource": fmt.Sprintf("%v/%v", cr.GetNamespace(), cr.GetName()),
	}))

	response, err := client.RoundTrip(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected status code when requesting revisions, got %v for dashboard %v", response.StatusCode, cr.GetName())
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
