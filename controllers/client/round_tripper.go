package client

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type instrumentedRoundTripper struct {
	relatedResource string
	wrapped         http.RoundTripper
	metric          *prometheus.CounterVec
}

func NewInstrumentedRoundTripper(relatedResource string, metric *prometheus.CounterVec, useProxy bool) http.RoundTripper {
	transport := &http.Transport{TLSClientConfig: &tls.Config{}}

	// Default transport respects proxy settings
	if useProxy {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}

	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1
	transport.TLSClientConfig.InsecureSkipVerify = true //nolint

	return &instrumentedRoundTripper{
		relatedResource: relatedResource,
		wrapped:         transport,
		metric:          metric,
	}
}

func (in *instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
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
