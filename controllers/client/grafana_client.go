package client

import (
	"context"
	"fmt"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGeneratedGrafanaClient(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (*genapi.GrafanaHTTPAPI, error) {
	tlsConfig, err := buildTLSConfiguration(ctx, cl, cr)
	if err != nil {
		return nil, fmt.Errorf("building tls config: %w", err)
	}

	httpClient := NewHTTPClient(cr, tlsConfig)

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return nil, err
	}

	// Secrets and ConfigMaps are not cached by default, get credentials as the last step.
	credentials, err := getAdminCredentials(ctx, cl, cr)
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
		Client:     httpClient,
	}
	if credentials.adminUser != "" {
		cfg.BasicAuth = url.UserPassword(credentials.adminUser, credentials.adminPassword)
	}

	gClient := genapi.NewHTTPClientWithConfig(nil, cfg)

	runtime, ok := gClient.Transport.(*httptransport.Runtime)
	if !ok {
		return nil, fmt.Errorf("casting client transport into *httptransport.Runtime to overwrite the default context")
	}

	runtime.Context = ctx

	return gClient, nil
}
