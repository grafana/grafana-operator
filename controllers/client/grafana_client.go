package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/grafana-operator/grafana-operator-experimental/controllers/metrics"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	grapi "github.com/grafana/grafana-api-golang-client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type instrumentedRoundTripper struct {
	instanceName string
	wrapped      http.RoundTripper
}

func newInstrumentedRoundTripper(instanceName string) http.RoundTripper {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &instrumentedRoundTripper{
		instanceName: instanceName,
		wrapped:      transport,
	}
}

func (in *instrumentedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := in.wrapped.RoundTrip(r)
	if resp != nil {
		metrics.GrafanaApiRequests.WithLabelValues(
			in.instanceName,
			r.URL.Path,
			r.Method,
			strconv.Itoa(resp.StatusCode)).
			Inc()
	}
	return resp, err
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
			Transport: newInstrumentedRoundTripper(grafana.Name),
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
