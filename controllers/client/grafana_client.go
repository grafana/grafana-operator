package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/config"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type GrafanaRequest struct {
	Dashboard  json.RawMessage `json:"dashboard"`
	FolderId   int64           `json:"folderId"`
	FolderName string          `json:"folderName"`
	Overwrite  bool            `json:"overwrite"`
}

type GrafanaResponse struct {
	ID         *uint   `json:"id"`
	OrgID      *uint   `json:"orgId"`
	Message    *string `json:"message"`
	Slug       *string `json:"slug"`
	Version    *int    `json:"version"`
	Status     *string `json:"resp"`
	UID        *string `json:"uid"`
	URL        *string `json:"url"`
	FolderId   *int64  `json:"folderId"`
	FolderName string  `json:"folderName"`
}

type GrafanaClient interface {
	CreateOrUpdateDashboard(dashboard *v1beta1.GrafanaDashboard) error
}

type GrafanaClientImpl struct {
	kubeClient client.Client
	httpClient *http.Client
	username   string
	password   string
	url        string
	ctx        context.Context
}

func NewGrafanaClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (GrafanaClient, error) {
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

	return &GrafanaClientImpl{
		url:        grafana.Status.AdminUrl,
		username:   username,
		password:   password,
		kubeClient: c,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   time.Second * timeoutSeconds,
		},
	}, nil
}

func (r *GrafanaClientImpl) CreateOrUpdateDashboard(dashboard *v1beta1.GrafanaDashboard) error {
	return nil
}
