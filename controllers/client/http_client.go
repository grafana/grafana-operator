package client

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewHTTPClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*http.Client, error) {
	var timeout time.Duration
	if grafana.Spec.Client != nil && grafana.Spec.Client.TimeoutSeconds != nil {
		timeout = time.Duration(*grafana.Spec.Client.TimeoutSeconds)
		if timeout < 0 {
			timeout = 0
		}
	} else {
		timeout = 10
	}

	tlsConfig, err := buildTLSConfiguration(ctx, c, grafana)
	if err != nil {
		return nil, err
	}

	transport := NewInstrumentedRoundTripper(grafana.IsExternal(), tlsConfig, metrics.GrafanaAPIRequests.MustCurryWith(prometheus.Labels{
		"instance_namespace": grafana.Namespace,
		"instance_name":      grafana.Name,
	}))
	if grafana.Spec.Client != nil && grafana.Spec.Client.Headers != nil {
		transport.(*instrumentedRoundTripper).addHeaders(grafana.Spec.Client.Headers) //nolint:errcheck
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * timeout,
	}, nil
}

// defaultTransport returns a new http.Transport with similar default values to
// http.DefaultTransport, but with idle connections and keepalives disabled.
func defaultTransport() *http.Transport {
	tp := defaultPooledTransport()
	tp.DisableKeepAlives = true
	tp.MaxIdleConnsPerHost = -1
	return tp
}

// defaultPooledTransport returns a new http.Transport with similar default
// values to http.DefaultTransport. Do not use this for transient transports as
// it can leak file descriptors over time. Only use this for transports that
// will be re-used for the same host(s).
func defaultPooledTransport() *http.Transport {
	tp := http.DefaultTransport.(*http.Transport).Clone() //nolint:errcheck
	tp.DisableKeepAlives = false
	tp.MaxIdleConnsPerHost = runtime.GOMAXPROCS(0) + 1
	return tp
}
