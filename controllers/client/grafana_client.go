package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/metrics"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	grapi "github.com/grafana/grafana-api-golang-client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type grafanaAdminCredentials struct {
	username string
	password string
}

func getAdminCredentials(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*grafanaAdminCredentials, error) {
	deployment := model.GetGrafanaDeployment(grafana, nil)
	selector := client.ObjectKey{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}

	err := c.Get(ctx, selector, deployment)
	if err != nil {
		return nil, err
	}

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
			return nil, errors.New(fmt.Sprintf("empty credential secret: %v/%v", grafana.Namespace, ref.Name))
		}

		if val, ok := secret.Data[ref.Key]; ok {
			return val, nil
		}

		return nil, errors.New(fmt.Sprintf("admin credentials not found: %v/%v", grafana.Namespace, ref.Name))
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

	userinfo := url.UserPassword(credentials.username, credentials.password)

	clientConfig := grapi.Config{
		APIKey:      "",
		BasicAuth:   userinfo,
		HTTPHeaders: nil,
		Client: &http.Client{
			Transport: NewInstrumentedRoundTripper(grafana.Name, metrics.GrafanaApiRequests),
			Timeout:   time.Second * timeout,
		},
		// TODO populate me
		OrgID: 0,
		// TODO populate me
		NumRetries: 0,
	}

	grafanaClient, err := grapi.New(grafana.Status.AdminUrl, clientConfig)
	if err != nil {
		return nil, err
	}

	return grafanaClient, nil
}
