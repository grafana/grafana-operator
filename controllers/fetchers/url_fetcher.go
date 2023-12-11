package fetchers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FetchDashboardFromUrl(dashboard *v1beta1.GrafanaDashboard) ([]byte, error) {
	url, err := url.Parse(dashboard.Spec.Url)
	if err != nil {
		return nil, err
	}

	cache := dashboard.GetContentCache()
	if len(cache) > 0 {
		return cache, nil
	}

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.DashboardUrlRequests, true)
	response, err := client.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from dashboard url request, get %v for dashboard %v", response.StatusCode, dashboard.Name)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	gz, err := v1beta1.Gzip(content)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to gzip dashboard %v", dashboard.Name)
	}

	dashboard.Status.ContentCache = gz
	dashboard.Status.ContentTimestamp = v1.Time{Time: time.Now()}
	dashboard.Status.ContentUrl = dashboard.Spec.Url

	return content, nil
}
