package client

import (
	"context"
	"net/http"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
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

	return &http.Client{
		Transport: NewInstrumentedRoundTripper(grafana.Name, metrics.GrafanaApiRequests, grafana.IsExternal(), tlsConfig),
		Timeout:   time.Second * timeout,
	}, nil
}
