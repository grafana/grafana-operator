package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const GrafanaVersionEndpoint = "/frontend/settings"

func NewHTTPClient(cr *v1beta1.Grafana, tlsConfig *tls.Config) *http.Client {
	var timeout time.Duration
	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		timeout = max(time.Duration(*cr.Spec.Client.TimeoutSeconds), 0)
	} else {
		timeout = 10
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
	}
}

func GetGrafanaVersion(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (string, error) {
	tlsConfig, err := buildTLSConfiguration(ctx, cl, cr)
	if err != nil {
		return "", fmt.Errorf("building tls config: %w", err)
	}

	httpClient := NewHTTPClient(cr, tlsConfig)

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return "", err
	}

	instanceURL := gURL.JoinPath(GrafanaVersionEndpoint).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, instanceURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("building request to fetch version: %w", err)
	}

	err = InjectAuthHeaders(ctx, cl, cr, req)
	if err != nil {
		return "", fmt.Errorf("fetching credentials for version detection: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	data := struct {
		BuildInfo struct {
			Version string `json:"version"`
		} `json:"buildInfo"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("parsing data from %s: %w", GrafanaVersionEndpoint, err)
	}

	if data.BuildInfo.Version == "" {
		return "", fmt.Errorf("empty version received from server")
	}

	return data.BuildInfo.Version, nil
}

func GetAuthenticationStatus(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (bool, error) {
	tlsConfig, err := buildTLSConfiguration(ctx, cl, cr)
	if err != nil {
		return false, fmt.Errorf("building tls config: %w", err)
	}

	httpClient := NewHTTPClient(cr, tlsConfig)

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return false, err
	}

	instanceURL := gURL.JoinPath("/login/ping").String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, instanceURL, http.NoBody)
	if err != nil {
		return false, fmt.Errorf("building request to fetch authentication status: %w", err)
	}

	err = InjectAuthHeaders(ctx, cl, cr, req)
	if err != nil {
		return false, fmt.Errorf("fetching credentials for authentication: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to authenticate with grafana, ensure connectivity and valid credentials")
	}

	data := struct {
		Message string `json:"message"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, fmt.Errorf("parsing health endpoint data: %w", err)
	}

	// The endpoint response has never changed since being introduced on 2015-04-07 in 22adf0d06e891a555d9ec40ec09f89d6679bafec (Grafana)
	if data.Message != "Logged in" {
		return false, fmt.Errorf("unexpected api response, expected: {\"message\": \"Logged in\"}, got: %v", data)
	}

	return true, nil
}
