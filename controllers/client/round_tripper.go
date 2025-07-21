package client

import (
	"crypto/tls"
	"log/slog"
	"maps"
	"net/http"
	"strconv"

	"github.com/grafana/grafana-operator/v5/embeds"
	"github.com/prometheus/client_golang/prometheus"
)

type RoundTripperOnResponse func(requestMethod string, responseCode int)

type instrumentedRoundTripper struct {
	wrapped http.RoundTripper
	headers map[string]string
	metrics []*prometheus.CounterVec
}

func NewInstrumentedRoundTripper(useProxy bool, tlsConfig *tls.Config, metrics ...*prometheus.CounterVec) http.RoundTripper {
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
		wrapped: transport,
		headers: headers,
		metrics: metrics,
	}
}

func (in *instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if in.headers != nil {
		for k, v := range in.headers {
			r.Header.Add(k, v)
		}
	}

	resp, err := in.wrapped.RoundTrip(r)
	if resp != nil {
		for _, m := range in.metrics {
			c, err := m.GetMetricWith(prometheus.Labels{
				"method": r.Method,
				"status": strconv.Itoa(resp.StatusCode),
			})
			if err != nil {
				slog.WarnContext(r.Context(), "failed constructing metric", "err", err)
				continue
			}

			c.Inc()
		}
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

	maps.Copy(in.headers, headers)
}
