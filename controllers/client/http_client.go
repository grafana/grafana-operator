package client

import (
	"context"
	"encoding/json"
	"fmt"
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

func GetGrafanaVersion(ctx context.Context, c client.Client, cr *v1beta1.Grafana) (string, error) {
	httpClient, err := NewHTTPClient(ctx, c, cr)
	if err != nil {
		return "", fmt.Errorf("setup of the http client: %w", err)
	}

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return "", err
	}

	instanceURL := gURL.JoinPath("/frontend/settings").String()

	req, err := http.NewRequest(http.MethodGet, instanceURL, nil)
	if err != nil {
		return "", fmt.Errorf("building request to fetch version: %w", err)
	}

	err = InjectAuthHeaders(ctx, c, cr, req)
	if err != nil {
		return "", fmt.Errorf("fetching credentials for version detection: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	data := struct {
		BuildInfo struct {
			Version string `json:"version"`
		} `json:"buildInfo"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("parsing health endpoint data: %w", err)
	}

	if data.BuildInfo.Version == "" {
		return "", fmt.Errorf("empty version received from server")
	}

	return data.BuildInfo.Version, nil
}
