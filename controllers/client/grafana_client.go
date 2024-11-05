package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type grafanaAdminCredentials struct {
	username string
	password string
	apikey   string
}

func getAdminCredentials(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*grafanaAdminCredentials, error) {
	credentials := &grafanaAdminCredentials{}

	if grafana.IsExternal() {
		// prefer api key if present
		if grafana.Spec.External.ApiKey != nil {
			apikey, err := GetValueFromSecretKey(ctx, grafana.Spec.External.ApiKey, c, grafana.Namespace)
			if err != nil {
				return nil, err
			}
			credentials.apikey = string(apikey)
			return credentials, nil
		}

		// rely on username and password otherwise
		username, err := GetValueFromSecretKey(ctx, grafana.Spec.External.AdminUser, c, grafana.Namespace)
		if err != nil {
			return nil, err
		}

		password, err := GetValueFromSecretKey(ctx, grafana.Spec.External.AdminPassword, c, grafana.Namespace)
		if err != nil {
			return nil, err
		}

		credentials.username = string(username)
		credentials.password = string(password)
		return credentials, nil
	}

	deployment := model.GetGrafanaDeployment(grafana, nil)
	selector := client.ObjectKey{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}

	err := c.Get(ctx, selector, deployment)
	if err != nil {
		return nil, err
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == config.GrafanaAdminUserEnvVar {
				if env.Value != "" {
					credentials.username = env.Value
					continue
				}

				if env.ValueFrom != nil {
					if env.ValueFrom.SecretKeyRef != nil {
						usernameFromSecret, err := GetValueFromSecretKey(ctx, env.ValueFrom.SecretKeyRef, c, grafana.Namespace)
						if err != nil {
							return nil, err
						}
						credentials.username = string(usernameFromSecret)
					}
				}
			}
			if env.Name == config.GrafanaAdminPasswordEnvVar {
				if env.Value != "" {
					credentials.password = env.Value
					continue
				}

				if env.ValueFrom != nil {
					if env.ValueFrom.SecretKeyRef != nil {
						passwordFromSecret, err := GetValueFromSecretKey(ctx, env.ValueFrom.SecretKeyRef, c, grafana.Namespace)
						if err != nil {
							return nil, err
						}
						credentials.password = string(passwordFromSecret)
					}
				}
			}
		}
	}
	return credentials, nil
}

func InjectAuthHeaders(ctx context.Context, c client.Client, grafana *v1beta1.Grafana, req *http.Request) error {
	creds, err := getAdminCredentials(ctx, c, grafana)
	if err != nil {
		return fmt.Errorf("fetching admin credentials: %w", err)
	}
	if creds.apikey != "" {
		req.Header.Add("Authorization", "Bearer "+creds.apikey)
	} else {
		req.SetBasicAuth(creds.username, creds.password)
	}
	return nil
}

func NewGeneratedGrafanaClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*genapi.GrafanaHTTPAPI, error) {
	var timeout time.Duration
	if grafana.Spec.Client != nil && grafana.Spec.Client.TimeoutSeconds != nil {
		timeout = time.Duration(*grafana.Spec.Client.TimeoutSeconds)
		if timeout < 0 {
			timeout = 0
		}
	} else {
		timeout = 10
	}

	credentials, err := getAdminCredentials(ctx, c, grafana)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := buildTLSConfiguration(ctx, c, grafana)
	if err != nil {
		return nil, err
	}

	gURL, err := url.Parse(grafana.Status.AdminUrl)
	if err != nil {
		return nil, fmt.Errorf("parsing url for client: %w", err)
	}

	transport := NewInstrumentedRoundTripper(grafana.Name, metrics.GrafanaApiRequests, grafana.IsExternal(), tlsConfig)

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout * time.Second,
	}

	cfg := &genapi.TransportConfig{
		Schemes:  []string{gURL.Scheme},
		BasePath: "/api",
		Host:     gURL.Host,
		// APIKey is an optional API key or service account token.
		APIKey: credentials.apikey,
		// NumRetries contains the optional number of attempted retries
		NumRetries:  0,
		Client:      client,
		TLSConfig:   tlsConfig,
		HTTPHeaders: grafana.Spec.Client.Headers,
	}
	if credentials.username != "" {
		cfg.BasicAuth = url.UserPassword(credentials.username, credentials.password)
	}
	cl := genapi.NewHTTPClientWithConfig(nil, cfg)

	return cl, nil
}
