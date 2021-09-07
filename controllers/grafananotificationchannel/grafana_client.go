/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grafananotificationchannel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	CreateNotificationChannelUrl      = "%v/api/alert-notifications/"
	ReadNotificationChannelByUIDUrl   = "%v/api/alert-notifications/uid/%v"
	UpdateNotificationChannelByUIDUrl = "%v/api/alert-notifications/uid/%v"
	DeleteNotificationChannelByUIDUrl = "%v/api/alert-notifications/uid/%v"
)

const (
	opCreate = "creating"
	opRead   = "reading"
	opUpdate = "updating"
	opDelete = "deleting"
)

type GrafanaResponse struct {
	ID                    *uint   `json:"id"`
	UID                   *string `json:"uid"`
	Name                  *string `json:"name"`
	Type                  *string `json:"type"`
	IsDefault             *bool   `json:"isDefault"`
	SendReminder          *bool   `json:"sendReminder"`
	DisableResolveMessage *bool   `json:"disableResolveMessage"`
	Created               *string `json:"created"`
	Updated               *string `json:"updated"`
	Message               *string `json:"message"`
}

type GrafanaClient interface {
	CreateNotificationChannel(channel []byte) (GrafanaResponse, error)
	DeleteNotificationChannelByUID(UID string) (GrafanaResponse, error)
	UpdateNotificationChannel(channel []byte, UID string) (GrafanaResponse, error)
	GetNotificationChannel(UID string) (GrafanaResponse, error)
}

type GrafanaClientImpl struct {
	url      string
	user     string
	password string
	client   *http.Client
	logger   logr.Logger
}

func setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "grafana-operator")
}

func NewGrafanaClient(url, user, password string, transport *http.Transport, timeoutSeconds time.Duration) GrafanaClient {
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * timeoutSeconds,
	}

	return &GrafanaClientImpl{
		url:      url,
		user:     user,
		password: password,
		client:   client,
	}
}

// CreateNotificationChannel Submits channel json to grafana
func (r *GrafanaClientImpl) CreateNotificationChannel(channel []byte) (GrafanaResponse, error) {
	return r.doRequest(opCreate, channel, "")
}

// UpdateNotificationChannel Updates existing channel
func (r *GrafanaClientImpl) UpdateNotificationChannel(channel []byte, UID string) (GrafanaResponse, error) {
	return r.doRequest(opUpdate, channel, UID)
}

// GetNotificationChannel Gets channel by UID
func (r *GrafanaClientImpl) GetNotificationChannel(UID string) (GrafanaResponse, error) {
	emptyChannel := make([]byte, 0)
	return r.doRequest(opRead, emptyChannel, UID)
}

// DeleteNotificationChannelByUID Deletes a channel given by a UID
func (r *GrafanaClientImpl) DeleteNotificationChannelByUID(UID string) (GrafanaResponse, error) {
	emptyChannel := make([]byte, 0)
	return r.doRequest(opDelete, emptyChannel, UID)
}

func (r *GrafanaClientImpl) doRequest(op string, channel []byte, UID string) (GrafanaResponse, error) {
	response := newResponse()

	var method, rawUrl string
	var body io.Reader

	switch op {
	case opCreate:
		method, body = "POST", bytes.NewBuffer(channel)
		rawUrl = fmt.Sprintf(CreateNotificationChannelUrl, r.url)
	case opRead:
		method, body = "GET", nil
		rawUrl = fmt.Sprintf(ReadNotificationChannelByUIDUrl, r.url, UID)
	case opUpdate:
		method, body = "PUT", bytes.NewBuffer(channel)
		rawUrl = fmt.Sprintf(UpdateNotificationChannelByUIDUrl, r.url, UID)
	case opDelete:
		method, body = "DELETE", nil
		rawUrl = fmt.Sprintf(DeleteNotificationChannelByUIDUrl, r.url, UID)
	default:
		return response, fmt.Errorf("error unknown operation %v", op)
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return response, err
	}

	parsed.User = url.UserPassword(r.user, r.password)

	req, err := http.NewRequest(method, parsed.String(), body)
	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			r.logger.Error(err, "failed to close body")
			return
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return response, fmt.Errorf(
			"error %v notificationChannel, expected status 200 but got %v",
			op, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)
	return response, err
}

func newResponse() GrafanaResponse {
	var id uint = 0
	var uid string
	var name = "(empty)"
	var nType string
	var isDefault = false
	var sendReminder = false
	var disableResolveMessage = false
	var created string
	var updated string
	var message = "(empty)"

	return GrafanaResponse{
		ID:                    &id,
		UID:                   &uid,
		Name:                  &name,
		Type:                  &nType,
		IsDefault:             &isDefault,
		SendReminder:          &sendReminder,
		DisableResolveMessage: &disableResolveMessage,
		Created:               &created,
		Updated:               &updated,
		Message:               &message,
	}
}
