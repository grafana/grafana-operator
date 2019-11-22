package grafanadashboard

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/grafana-tools/sdk"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	DeleteDashboardByUIDUrl = "%v/api/dashboards/uid/%v"
)

type GrafanaClient interface {
	CreateOrUpdateDashboard(dashboard sdk.Board) (sdk.StatusMessage, error)
	DeleteDashboardByUID(UID string) (sdk.StatusMessage, error)
}

type GrafanaClientImpl struct {
	url           string
	user          string
	password      string
	client        *http.Client
	grafanaClient *sdk.Client
}

func NewGrafanaClient(url, user, password string) GrafanaClient {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: &transport,
	}

	return &GrafanaClientImpl{
		url:           url,
		user:          user,
		password:      password,
		client:        client,
		grafanaClient: sdk.NewClient(url, fmt.Sprintf("%v:%v", user, password), client),
	}
}

func (r *GrafanaClientImpl) CreateOrUpdateDashboard(dashboard sdk.Board) (sdk.StatusMessage, error) {
	return r.grafanaClient.SetDashboard(dashboard, true)
}

func (r *GrafanaClientImpl) DeleteDashboardByUID(UID string) (sdk.StatusMessage, error) {
	rawUrl := fmt.Sprintf(DeleteDashboardByUIDUrl, r.url, UID)
	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return sdk.StatusMessage{}, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("DELETE", parsed.String(), nil)
	if err != nil {
		return sdk.StatusMessage{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "autograf")
	resp, err := r.client.Do(req)
	if err != nil {
		return sdk.StatusMessage{}, err
	}

	reply := sdk.StatusMessage{}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return sdk.StatusMessage{}, err
	}

	err = json.Unmarshal(data, &reply)
	return reply, err
}
