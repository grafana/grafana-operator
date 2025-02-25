package fetchers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaClient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"

	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FetchFromUrl(ctx context.Context, cr v1beta1.GrafanaContentResource, c client.Client, tlsConfig *tls.Config) ([]byte, error) {
	spec := cr.GrafanaContentSpec()

	url, err := url.Parse(spec.Url)
	if err != nil {
		return nil, err
	}

	cached := cache.GetContentCache(cr)
	if len(cached) > 0 {
		return cached, nil
	}

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	contentMetric, err := metrics.ContentUrlRequests.CurryWith(prometheus.Labels{
		"kind":     cr.GetObjectKind().GroupVersionKind().Kind,
		"resource": fmt.Sprintf("%v/%v", cr.GetNamespace(), cr.GetName()),
	})
	if err != nil {
		return nil, fmt.Errorf("building dashboards metric: %w", err)
	}

	// this is a documented deprecated metric but we don't want to fail lint
	//nolint:staticcheck
	dashboardMetric, err := metrics.DashboardUrlRequests.CurryWith(prometheus.Labels{
		"dashboard": fmt.Sprintf("%v/%v", cr.GetNamespace(), cr.GetName()),
	})
	if err != nil {
		return nil, fmt.Errorf("building dashboards metric: %w", err)
	}

	client := client2.NewInstrumentedRoundTripper(true, tlsConfig, contentMetric, dashboardMetric)
	// basic auth is supported for dashboards from url
	if spec.UrlAuthorization != nil && spec.UrlAuthorization.BasicAuth != nil {
		username, err := grafanaClient.GetValueFromSecretKey(ctx, spec.UrlAuthorization.BasicAuth.Username, c, cr.GetNamespace())
		if err != nil {
			return nil, err
		}

		password, err := grafanaClient.GetValueFromSecretKey(ctx, spec.UrlAuthorization.BasicAuth.Password, c, cr.GetNamespace())
		if err != nil {
			return nil, err
		}

		if username != nil && password != nil {
			request.SetBasicAuth(string(username), string(password))
		} else {
			return nil, fmt.Errorf("basic auth username and/or password are missing for dashboard %s/%s", cr.GetNamespace(), cr.GetName())
		}
	}

	response, err := client.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from dashboard url request, get %v for dashboard %v", response.StatusCode, cr.GetName())
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}

	gz, err := cache.Gzip(content)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to gzip dashboard %v", cr.GetName())
	}

	status := cr.GrafanaContentStatus()
	status.ContentCache = gz
	status.ContentTimestamp = v1.Time{Time: time.Now()}
	status.ContentUrl = spec.Url

	return content, nil
}
