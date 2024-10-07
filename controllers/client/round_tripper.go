package client

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/grafana/grafana-operator/v5/embeds"
	"github.com/prometheus/client_golang/prometheus"
)

type instrumentedRoundTripper struct {
	relatedResource string
	wrapped         http.RoundTripper
	metric          *prometheus.CounterVec
}

func NewInstrumentedRoundTripper(relatedResource string, metric *prometheus.CounterVec, useProxy bool, tlsConfig *tls.Config) http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	if !useProxy {
		transport.Proxy = nil
	}

	return &instrumentedRoundTripper{
		relatedResource: relatedResource,
		wrapped:         transport,
		metric:          metric,
	}
}

func (in *instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("user-agent", "grafana-operator/"+embeds.Version)
	resp, err := in.wrapped.RoundTrip(r)
	if resp != nil {
		in.metric.WithLabelValues(
			in.relatedResource,
			r.Method,
			strconv.Itoa(resp.StatusCode)).
			Inc()
	}
	return resp, err
}
