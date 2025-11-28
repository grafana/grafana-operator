package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGeneratedGrafanaClient(ctx context.Context, c client.Client, cr *v1beta1.Grafana) (*genapi.GrafanaHTTPAPI, error) {
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

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
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

	// Secrets and ConfigMaps are not cached by default, get credentials as the last step.
	credentials, err := getAdminCredentials(ctx, c, cr)
	if err != nil {
		return nil, err
	}

	cfg := &genapi.TransportConfig{
		Schemes:  []string{gURL.Scheme},
		BasePath: gURL.Path,
		Host:     gURL.Host,
		// APIKey is an optional API key or service account token.
		APIKey: credentials.apikey,
		// NumRetries contains the optional number of attempted retries
		NumRetries: 0,
		TLSConfig:  tlsConfig,
		Client: &http.Client{
			Transport: transport,
			Timeout:   timeout * time.Second,
		},
	}
	if credentials.adminUser != "" {
		cfg.BasicAuth = url.UserPassword(credentials.adminUser, credentials.adminPassword)
	}

	cl := genapi.NewHTTPClientWithConfig(nil, cfg)

	runtime, ok := cl.Transport.(*httptransport.Runtime)
	if !ok {
		return nil, fmt.Errorf("casting client transport into *httptransport.Runtime to overwrite the default context")
	}

	runtime.Context = ctx

	return cl, nil
}
