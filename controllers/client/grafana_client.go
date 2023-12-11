package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	v1 "k8s.io/api/core/v1"

	grapi "github.com/grafana/grafana-api-golang-client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
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
	getValueFromSecret := func(ref *v1.SecretKeySelector) ([]byte, error) {
		secret := &v1.Secret{}
		selector := client.ObjectKey{
			Name:      ref.Name,
			Namespace: grafana.Namespace,
		}
		err := c.Get(ctx, selector, secret)
		if err != nil {
			return nil, err
		}

		if secret.Data == nil {
			return nil, fmt.Errorf("empty credential secret: %v/%v", grafana.Namespace, ref.Name)
		}

		if val, ok := secret.Data[ref.Key]; ok {
			return val, nil
		}

		return nil, fmt.Errorf("admin credentials not found: %v/%v", grafana.Namespace, ref.Name)
	}

	if grafana.IsExternal() {
		// prefer api key if present
		if grafana.Spec.External.ApiKey != nil {
			apikey, err := getValueFromSecret(grafana.Spec.External.ApiKey)
			if err != nil {
				return nil, err
			}
			credentials.apikey = string(apikey)
			return credentials, nil
		}

		// rely on username and password otherwise
		username, err := getValueFromSecret(grafana.Spec.External.AdminUser)
		if err != nil {
			return nil, err
		}

		password, err := getValueFromSecret(grafana.Spec.External.AdminPassword)
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
						usernameFromSecret, err := getValueFromSecret(env.ValueFrom.SecretKeyRef)
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
						passwordFromSecret, err := getValueFromSecret(env.ValueFrom.SecretKeyRef)
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

func NewGrafanaClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*grapi.Client, error) {
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

	clientConfig := grapi.Config{
		HTTPHeaders: nil,
		Client: &http.Client{
			Transport: NewInstrumentedRoundTripper(grafana.Name, metrics.GrafanaApiRequests, grafana.IsExternal()),
			Timeout:   time.Second * timeout,
		},
		// TODO populate me
		OrgID: 0,
		// TODO populate me
		NumRetries: 0,
	}

	if credentials.apikey != "" {
		clientConfig.APIKey = credentials.apikey
	}

	if credentials.username != "" && credentials.password != "" {
		clientConfig.BasicAuth = url.UserPassword(credentials.username, credentials.password)
	}

	grafanaClient, err := grapi.New(grafana.Status.AdminUrl, clientConfig)
	if err != nil {
		return nil, err
	}

	return grafanaClient, nil
}
