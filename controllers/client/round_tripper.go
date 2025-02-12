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
	headers         map[string]string
}

func NewInstrumentedRoundTripper(relatedResource string, metric *prometheus.CounterVec, useProxy bool, tlsConfig *tls.Config) http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone() //nolint:errcheck

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1

	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	if !useProxy {
		transport.Proxy = nil
	}

	headers := make(map[string]string)
	headers["user-agent"] = "grafana-operator/" + embeds.Version

	return &instrumentedRoundTripper{
		relatedResource: relatedResource,
		wrapped:         transport,
		metric:          metric,
		headers:         headers,
	}
}

func (in *instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if in.headers != nil {
		for k, v := range in.headers {
			r.Header.Add(k, v)
		}
	}

	resp, err := in.wrapped.RoundTrip(r)
	if resp != nil && in.metric != nil {
		in.metric.WithLabelValues(
			in.relatedResource,
			r.Method,
			strconv.Itoa(resp.StatusCode)).
			Inc()
	}
	return resp, err
}

func (in *instrumentedRoundTripper) addHeaders(headers map[string]string) {
	if headers == nil {
		return
	}

	if in.headers == nil {
		in.headers = make(map[string]string)
	}

	for k, v := range headers {
		in.headers[k] = v
	}
}
