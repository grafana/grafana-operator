package grafanadatasource

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	DeleteDataSourceByNameUrl = "%v/api/datasources/name/%v"
)

type DataSourceDeleteResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
}

type GrafanaClient interface {
	DeleteDataSourceByName(dsName string) (DataSourceDeleteResponse, error)
}

type GrafanaClientImpl struct {
	url      string
	user     string
	password string
	client   *http.Client
}

func setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "grafana-operator")
}

func NewGrafanaClient(url, user, password string, timeoutSeconds time.Duration) GrafanaClient {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: &transport,
		Timeout:   time.Second * timeoutSeconds,
	}

	return &GrafanaClientImpl{
		url:      url,
		user:     user,
		password: password,
		client:   client,
	}
}

// Delete a datasource given by a Name
func (r *GrafanaClientImpl) DeleteDataSourceByName(dsName string) (DataSourceDeleteResponse, error) {
	rawUrl := fmt.Sprintf(DeleteDataSourceByNameUrl, r.url, dsName)
	response := DataSourceDeleteResponse{}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return response, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("DELETE", parsed.String(), nil)
	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	// Skip 404 not found because data source can be deleted via UI
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return response, errors.New(fmt.Sprintf(
			"error deleting datasource, expected status 200 or 404 but got %v",
			resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)

	return response, err
}
