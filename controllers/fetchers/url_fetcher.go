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
	"github.com/grafana/grafana-operator/v5/controllers/metrics"

	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FetchDashboardFromUrl(ctx context.Context, dashboard *v1beta1.GrafanaDashboard, c client.Client, tlsConfig *tls.Config) ([]byte, error) {
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

	client := client2.NewInstrumentedRoundTripper(fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Name), metrics.DashboardUrlRequests, true, tlsConfig)
	// basic auth is supported for dashboards from url
	if dashboard.Spec.UrlAuthorization != nil && dashboard.Spec.UrlAuthorization.BasicAuth != nil {
		username, err := grafanaClient.GetValueFromSecretKey(ctx, dashboard.Spec.UrlAuthorization.BasicAuth.Username, c, dashboard.Namespace)
		if err != nil {
			return nil, err
		}

		password, err := grafanaClient.GetValueFromSecretKey(ctx, dashboard.Spec.UrlAuthorization.BasicAuth.Password, c, dashboard.Namespace)
		if err != nil {
			return nil, err
		}

		if username != nil && password != nil {
			request.SetBasicAuth(string(username), string(password))
		} else {
			return nil, fmt.Errorf("basic auth username and/or password are missing for dashboard %s/%s", dashboard.Namespace, dashboard.Name)
		}
	}

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
