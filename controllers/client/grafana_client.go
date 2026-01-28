package client

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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

func restConfigFor(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (*rest.Config, error) {
	config := &rest.Config{}
	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		config.Timeout = max(time.Duration(*cr.Spec.Client.TimeoutSeconds), 0) * time.Second
	} else {
		config.Timeout = 10 * time.Second
	}

	tlsConfig, err := buildRestTLSConfig(ctx, cl, cr)
	if err != nil {
		return nil, err
	}

	config.TLSClientConfig = tlsConfig

	gURL, err := url.Parse(cr.Status.AdminURL)
	if err != nil {
		return nil, err
	}

	config.Host = (&url.URL{
		Host:   gURL.Host,
		Scheme: gURL.Scheme,
	}).String()
	config.APIPath = path.Join(gURL.Path, "apis")

	// Secrets and ConfigMaps are not cached by default, get credentials as the last step.
	credentials, err := getAdminCredentials(ctx, cl, cr)
	if err != nil {
		return nil, err
	}

	switch {
	case credentials.apikey != "":
		config.BearerToken = credentials.apikey
	case credentials.adminUser != "":
		config.Username = credentials.adminUser
		config.Password = credentials.adminPassword
	}

	return config, nil
}

func NewDynamicClient(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (dynamic.Interface, *discovery.DiscoveryClient, error) {
	config, err := restConfigFor(ctx, cl, cr)
	if err != nil {
		return nil, nil, fmt.Errorf("building rest config for client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("building k8s client: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("building discovery client: %w", err)
	}

	return dynamicClient, dc, nil
}
