package grafananotificationchannel

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
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

// Submit channel json to grafana
func (r *GrafanaClientImpl) CreateNotificationChannel(channel []byte) (GrafanaResponse, error) {
	log.Info(fmt.Sprintf("creating new channel"))
	return r.doRequest(opCreate, channel, "")
}

// Update existing channel
func (r *GrafanaClientImpl) UpdateNotificationChannel(channel []byte, UID string) (GrafanaResponse, error) {
	log.Info(fmt.Sprintf("updating channel UID: %v", UID))
	return r.doRequest(opUpdate, channel, UID)
}

// Get channel by UID
func (r *GrafanaClientImpl) GetNotificationChannel(UID string) (GrafanaResponse, error) {
	log.Info(fmt.Sprintf("checking channel UID: %v", UID))
	emptyChannel := make([]byte, 0)
	return r.doRequest(opRead, emptyChannel, UID)
}

// Delete a channel given by a UID
func (r *GrafanaClientImpl) DeleteNotificationChannelByUID(UID string) (GrafanaResponse, error) {
	emptyChannel := make([]byte, 0)
	log.Info(fmt.Sprintf("deleting channel UID: %v", UID))
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
		return response, errors.New(fmt.Sprintf("error unknown operation %v", op))
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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, errors.New(fmt.Sprintf(
			"error %v notificationChannel, expected status 200 but got %v",
			op, resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	log.Info(fmt.Sprintf("response %v", string(data)))
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
