package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	grapi "github.com/grafana/grafana-api-golang-client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGrafanaClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*grapi.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	var timeoutSeconds time.Duration
	if grafana.Spec.Client != nil && grafana.Spec.Client.TimeoutSeconds != nil {
		timeoutSeconds = time.Duration(*grafana.Spec.Client.TimeoutSeconds)
		if timeoutSeconds < 0 {
			timeoutSeconds = 0
		}
	} else {
		timeoutSeconds = 10
	}

	credentialSecret := model.GetGrafanaAdminSecret(grafana, nil)
	selector := client.ObjectKey{
		Namespace: credentialSecret.Namespace,
		Name:      credentialSecret.Name,
	}

	err := c.Get(ctx, selector, credentialSecret)
	if err != nil {
		return nil, err
	}

	username := ""
	password := ""
	if val, ok := credentialSecret.Data[config.GrafanaAdminUserEnvVar]; ok {
		username = string(val)
	} else {
		return nil, errors.New("grafana admin secret does not contain username")
	}

	if val, ok := credentialSecret.Data[config.GrafanaAdminPasswordEnvVar]; ok {
		password = string(val)
	} else {
		return nil, errors.New("grafana admin secret does not contain password")
	}

	userinfo := url.UserPassword(username, password)

	clientConfig := grapi.Config{
		APIKey:      "",
		BasicAuth:   userinfo,
		HTTPHeaders: nil,
		Client: &http.Client{
			Transport: transport,
			Timeout:   time.Second * timeoutSeconds,
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
