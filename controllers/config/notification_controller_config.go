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

package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
)

type NotificationControllerConfig struct {
	*sync.Mutex
	Values               map[string]interface{}
	NotificationChannels map[string][]*v1alpha1.GrafanaNotificationChannelRef
}

var notificationConfig *NotificationControllerConfig
var ncInitOnce sync.Once

func GetNotificationControllerConfig() *NotificationControllerConfig {
	ncInitOnce.Do(func() {
		notificationConfig = &NotificationControllerConfig{
			Mutex:                &sync.Mutex{},
			Values:               map[string]interface{}{},
			NotificationChannels: map[string][]*v1alpha1.GrafanaNotificationChannelRef{},
		}
	})
	return notificationConfig
}

func (c *NotificationControllerConfig) GetNotificationChannelId(namespace, name string) string {
	return fmt.Sprintf("%v/%v", namespace, name)
}

func (c *NotificationControllerConfig) AddNotificationChannel(notificationchannel *v1alpha1.GrafanaNotificationChannel) {
	ns := notificationchannel.Namespace
	if i, exists := c.HasNotificationChannel(ns, notificationchannel.Name); !exists {
		c.Lock()
		defer c.Unlock()
		c.NotificationChannels[ns] = append(c.NotificationChannels[ns], &v1alpha1.GrafanaNotificationChannelRef{
			Name:      notificationchannel.Name,
			Namespace: ns,
			UID:       notificationchannel.Status.UID,
			Hash:      notificationchannel.Status.Hash,
		})
	} else {
		c.Lock()
		defer c.Unlock()
		c.NotificationChannels[ns][i].Namespace = ns
		c.NotificationChannels[ns][i].UID = notificationchannel.Status.UID
		c.NotificationChannels[ns][i].Hash = notificationchannel.Status.Hash
	}
}

func (c *NotificationControllerConfig) InvalidateNotificationChannels() {
	c.Lock()
	defer c.Unlock()
	for _, v := range c.NotificationChannels {
		for _, d := range v {
			d.Hash = ""
		}
	}
}

func (c *NotificationControllerConfig) SetNotificationChannels(notificationchannels map[string][]*v1alpha1.GrafanaNotificationChannelRef) {
	c.Lock()
	defer c.Unlock()
	c.NotificationChannels = notificationchannels
}

func (c *NotificationControllerConfig) RemoveNotificationChannel(namespace, name string) {
	if i, exists := c.HasNotificationChannel(namespace, name); exists {
		c.Lock()
		defer c.Unlock()
		list := c.NotificationChannels[namespace]
		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
		c.NotificationChannels[namespace] = list
	}
}

func (c *NotificationControllerConfig) GetNotificationChannels(namespace string) []*v1alpha1.GrafanaNotificationChannelRef {
	c.Lock()
	defer c.Unlock()
	// Cluster level?
	if namespace == "" {
		var notificationchannels []*v1alpha1.GrafanaNotificationChannelRef
		for _, v := range c.NotificationChannels {
			notificationchannels = append(notificationchannels, v...)
		}
		return notificationchannels
	}

	if notificationchannels, ok := c.NotificationChannels[namespace]; ok {
		return notificationchannels
	}
	return []*v1alpha1.GrafanaNotificationChannelRef{}
}

func (c *NotificationControllerConfig) AddConfigItem(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if key != "" && value != nil && value != "" {
		c.Values[key] = value
	}
}

func (c *NotificationControllerConfig) RemoveConfigItem(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.Values, key)
}

func (c *NotificationControllerConfig) GetConfigItem(key string, defaultValue interface{}) interface{} {
	if c.HasConfigItem(key) {
		return c.Values[key]
	}
	return defaultValue
}

func (c *NotificationControllerConfig) GetConfigString(key, defaultValue string) string {
	if c.HasConfigItem(key) {
		return c.Values[key].(string)
	}
	return defaultValue
}

func (c *NotificationControllerConfig) GetConfigBool(key string, defaultValue bool) bool {
	if c.HasConfigItem(key) {
		return c.Values[key].(bool)
	}
	return defaultValue
}

func (c *NotificationControllerConfig) GetConfigTimestamp(key string, defaultValue time.Time) time.Time {
	if c.HasConfigItem(key) {
		return c.Values[key].(time.Time)
	}
	return defaultValue
}

func (c *NotificationControllerConfig) HasConfigItem(key string) bool {
	c.Lock()
	defer c.Unlock()
	_, ok := c.Values[key]
	return ok
}

func (c *NotificationControllerConfig) HasNotificationChannel(namespace, name string) (int, bool) {
	if notificationchannels, ok := c.NotificationChannels[namespace]; ok {
		for i, notificationchannel := range notificationchannels {
			if notificationchannel.Name == name {
				return i, true
			}
		}
	}
	return -1, false
}
