package fetchers

import (
	"errors"
	"fmt"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/metrics"
	"io"
	"net/http"
	"net/url"
)

func FetchDashboardFromUrl(dashboard *v1beta1.GrafanaDashboard) ([]byte, error) {
	url, err := url.Parse(dashboard.Spec.Url)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.DashboardUrlRequests)
	response, err := client.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unexpected status code from dashboard url request, get %v for dashboard %v", response.StatusCode, dashboard.Name))
	}
	return io.ReadAll(response.Body)
}
