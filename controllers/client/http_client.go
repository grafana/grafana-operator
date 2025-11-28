package client

import (
	"context"
	"net/http"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewHTTPClient(ctx context.Context, c client.Client, cr *v1beta1.Grafana) (*http.Client, error) {
	var timeout time.Duration
	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		timeout = max(time.Duration(*cr.Spec.Client.TimeoutSeconds), 0)
	} else {
		timeout = 10
	}

	tlsConfig, err := buildTLSConfiguration(ctx, c, cr)
	if err != nil {
		return nil, err
	}

	transport := NewInstrumentedRoundTripper(cr.IsExternal(), tlsConfig, metrics.GrafanaAPIRequests.MustCurryWith(prometheus.Labels{
		"instance_namespace": cr.Namespace,
		"instance_name":      cr.Name,
	}))
	if cr.Spec.Client != nil && cr.Spec.Client.Headers != nil {
		transport.(*instrumentedRoundTripper).addHeaders(cr.Spec.Client.Headers) //nolint:errcheck
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * timeout,
	}, nil
}
