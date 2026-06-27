package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Silences are part of the Grafana-managed Alertmanager API, which is not covered by the
// generated grafana-openapi-client-go. They are therefore managed through raw HTTP calls,
// following the same pattern as GetGrafanaVersion/GetAuthenticationStatus.
//
// The endpoint is relative to the "/api" base path already added by ParseAdminURL, so the
// effective paths are:
//
//	POST   <admin>/api/alertmanager/grafana/api/v2/silences
//	GET    <admin>/api/alertmanager/grafana/api/v2/silence/{id}
//	DELETE <admin>/api/alertmanager/grafana/api/v2/silence/{id}
const (
	silencesEndpoint = "/alertmanager/grafana/api/v2/silences"
	silenceEndpoint  = "/alertmanager/grafana/api/v2/silence"
)

// SilenceMatcher mirrors the Alertmanager matcher payload.
type SilenceMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

// PostableSilence is the request body for creating or updating a silence. When ID is set
// the existing silence is updated, otherwise a new one is created.
type PostableSilence struct {
	ID        string           `json:"id,omitempty"`
	Matchers  []SilenceMatcher `json:"matchers"`
	StartsAt  string           `json:"startsAt"`
	EndsAt    string           `json:"endsAt"`
	CreatedBy string           `json:"createdBy"`
	Comment   string           `json:"comment"`
}

// GettableSilence is the relevant subset of a silence returned by Grafana.
type GettableSilence struct {
	ID     string `json:"id"`
	Status struct {
		// State is one of "active", "pending" or "expired".
		State string `json:"state"`
	} `json:"status"`
}

func silenceHTTPClient(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (*http.Client, error) {
	tlsConfig, err := buildTLSConfiguration(ctx, cl, cr)
	if err != nil {
		return nil, fmt.Errorf("building tls config: %w", err)
	}

	return NewHTTPClient(cr, tlsConfig), nil
}

// CreateOrUpdateSilence posts a silence to Grafana and returns the silence ID assigned by
// the server. If payload.ID is set, the matching silence is updated in place.
func CreateOrUpdateSilence(ctx context.Context, cl client.Client, cr *v1beta1.Grafana, payload PostableSilence) (string, error) {
	httpClient, err := silenceHTTPClient(ctx, cl, cr)
	if err != nil {
		return "", err
	}

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encoding silence payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gURL.JoinPath(silencesEndpoint).String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request to create silence: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if err := InjectAuthHeaders(ctx, cl, cr, req); err != nil {
		return "", fmt.Errorf("fetching credentials for silence: %w", err)
	}

	resp, err := httpClient.Do(req) //#nosec G704
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status creating silence: %s: %s", resp.Status, readErrorBody(resp.Body))
	}

	data := struct {
		SilenceID string `json:"silenceID"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("parsing silence creation response: %w", err)
	}

	if data.SilenceID == "" {
		return "", fmt.Errorf("empty silence ID received from server")
	}

	return data.SilenceID, nil
}

// GetSilence fetches a silence by ID. It returns (nil, nil) when the silence does not exist.
func GetSilence(ctx context.Context, cl client.Client, cr *v1beta1.Grafana, id string) (*GettableSilence, error) {
	httpClient, err := silenceHTTPClient(ctx, cl, cr)
	if err != nil {
		return nil, err
	}

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gURL.JoinPath(silenceEndpoint, id).String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("building request to get silence: %w", err)
	}

	if err := InjectAuthHeaders(ctx, cl, cr, req); err != nil {
		return nil, fmt.Errorf("fetching credentials for silence: %w", err)
	}

	resp, err := httpClient.Do(req) //#nosec G704
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status getting silence: %s: %s", resp.Status, readErrorBody(resp.Body))
	}

	silence := &GettableSilence{}
	if err := json.NewDecoder(resp.Body).Decode(silence); err != nil {
		return nil, fmt.Errorf("parsing silence response: %w", err)
	}

	return silence, nil
}

// DeleteSilence expires a silence by ID. A missing silence is treated as already deleted.
func DeleteSilence(ctx context.Context, cl client.Client, cr *v1beta1.Grafana, id string) error {
	httpClient, err := silenceHTTPClient(ctx, cl, cr)
	if err != nil {
		return err
	}

	gURL, err := ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, gURL.JoinPath(silenceEndpoint, id).String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("building request to delete silence: %w", err)
	}

	if err := InjectAuthHeaders(ctx, cl, cr, req); err != nil {
		return fmt.Errorf("fetching credentials for silence: %w", err)
	}

	resp, err := httpClient.Do(req) //#nosec G704
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil
	}

	return fmt.Errorf("unexpected status deleting silence: %s: %s", resp.Status, readErrorBody(resp.Body))
}

func readErrorBody(body io.Reader) string {
	b, err := io.ReadAll(io.LimitReader(body, 1<<16))
	if err != nil {
		return ""
	}

	return string(bytes.TrimSpace(b))
}
