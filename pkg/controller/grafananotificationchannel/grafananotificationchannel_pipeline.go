package grafananotificationchannel

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type NotificationChannelPipeline interface {
	ProcessNotificationChannel(knownHash string) ([]byte, error)
	NewHash() string
}

type NotificatiomChannelPipelineImpl struct {
	Client              client.Client
	NotificationChannel *v1alpha1.GrafanaNotificationChannel
	JSON                string
	Channel             map[string]interface{}
	Logger              logr.Logger
	Hash                string
}

func NewNotificationChannelPipeline(client client.Client, notificationChannel *v1alpha1.GrafanaNotificationChannel) NotificationChannelPipeline {
	return &NotificatiomChannelPipelineImpl{
		Client:              client,
		NotificationChannel: notificationChannel,
		JSON:                "",
		Logger:              logf.Log.WithName(fmt.Sprintf("notificationChannel-%v", notificationChannel.Name)),
	}
}

func (r *NotificatiomChannelPipelineImpl) ProcessNotificationChannel(knownHash string) ([]byte, error) {
	err := r.obtainJson()
	if err != nil {
		return nil, err
	}

	// NotificationChannel unchanged?
	hash := r.generateHash()
	if hash == knownHash {
		r.Hash = knownHash
		return nil, nil
	}
	r.Hash = hash

	// NotificationChannel valid?
	err = r.validateJson()
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(r.Channel)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(raw), nil
}

// Make sure the notificationchannel contains valid JSON
func (r *NotificatiomChannelPipelineImpl) validateJson() error {
	notificationchannelBytes := []byte(r.JSON)
	r.Channel = make(map[string]interface{})
	return json.Unmarshal(notificationchannelBytes, &r.Channel)
}

// Try to get the notificationchannel json definition raw json provided in the notificationchannel resource
func (r *NotificatiomChannelPipelineImpl) obtainJson() error {
	if r.NotificationChannel.Spec.Json != "" {
		r.JSON = r.NotificationChannel.Spec.Json
		return nil
	}

	return errors.New("notificationchannel does not contain json")
}

// Create a hash of the notificationchannel to detect if there are actually changes to the json
// If there are no changes we should avoid sending update requests as this will create
// a new notificationchannel version in Grafana
func (r *NotificatiomChannelPipelineImpl) generateHash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(
		r.NotificationChannel.Spec.Json)))
}

func (r *NotificatiomChannelPipelineImpl) NewHash() string {
	return r.Hash
}
