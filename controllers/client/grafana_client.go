package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// ResolveNamespace returns the Grafana app-platform namespace the instance addresses
// resources in ("default" / "org-N" for on-prem, "stacks-N" for Grafana Cloud).
// It reads the value from the instance's /api/frontend/settings response using the
// same authentication path as the typed client, so it works for token, basic-auth,
// and in-cluster-SA-token configurations.
func ResolveNamespace(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (string, error) {
	cfg, err := restConfigFor(ctx, cl, cr)
	if err != nil {
		return "", fmt.Errorf("building rest config: %w", err)
	}

	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return "", fmt.Errorf("building http client: %w", err)
	}

	endpoint, err := url.JoinPath(cfg.Host, "api", "frontend", "settings")
	if err != nil {
		return "", fmt.Errorf("building frontend settings url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching frontend settings: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("frontend settings returned %d", resp.StatusCode)
	}

	var payload struct {
		Namespace string `json:"namespace"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding frontend settings: %w", err)
	}

	if payload.Namespace == "" {
		return "", fmt.Errorf("frontend settings response missing namespace")
	}

	return payload.Namespace, nil
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
