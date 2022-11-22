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

package v1alpha1

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1" //nolint
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	// Json is the dashboard's JSON
	Json string `json:"json,omitempty"`
	// GzipJson the dashboard's JSON compressed with Gzip. Base64-encoded when in YAML.
	GzipJson []byte     `json:"gzipJson,omitempty"`
	Jsonnet  string     `json:"jsonnet,omitempty"`
	Plugins  PluginList `json:"plugins,omitempty"`
	Url      string     `json:"url,omitempty"`
	// ConfigMapRef is a reference to a ConfigMap data field containing the dashboard's JSON
	ConfigMapRef *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`
	// GzipConfigMapRef is a reference to a ConfigMap binaryData field containing
	// the dashboard's JSON, compressed with Gzip.
	GzipConfigMapRef *corev1.ConfigMapKeySelector      `json:"gzipConfigMapRef,omitempty"`
	Datasources      []GrafanaDashboardDatasource      `json:"datasources,omitempty"`
	CustomFolderName string                            `json:"customFolderName,omitempty"`
	GrafanaCom       *GrafanaDashboardGrafanaComSource `json:"grafanaCom,omitempty"`

	// ContentCacheDuration sets how often the operator should resync with the external source when using
	// the `grafanaCom.id` or `url` field to specify the source of the dashboard. The default value is
	// decided by the `dashboardContentCacheDuration` field in the `Grafana` resource. The default is 0 which
	// is interpreted as never refetching.
	ContentCacheDuration *metav1.Duration `json:"contentCacheDuration,omitempty"`
}

type GrafanaDashboardDatasource struct {
	InputName      string `json:"inputName"`
	DatasourceName string `json:"datasourceName"`
}

type GrafanaDashboardGrafanaComSource struct {
	Id       int  `json:"id"`
	Revision *int `json:"revision,omitempty"`
}

// GrafanaDashboardRef is used to keep a dashboard reference without having access to the dashboard
// struct itself
type GrafanaDashboardRef struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	UID        string `json:"uid"`
	Hash       string `json:"hash"`
	FolderId   *int64 `json:"folderId"`
	FolderName string `json:"folderName"`
}

type GrafanaDashboardStatus struct {
	ContentCache     []byte                 `json:"contentCache,omitempty"`
	ContentTimestamp metav1.Time            `json:"contentTimestamp,omitempty"`
	ContentUrl       string                 `json:"contentUrl,omitempty"`
	Error            *GrafanaDashboardError `json:"error,omitempty"`
}

type GrafanaDashboardError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
	Retries int    `json:"retries,omitempty"`
}

// GrafanaDashboard is the Schema for the grafanadashboards API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDashboardSpec   `json:"spec,omitempty"`
	Status GrafanaDashboardStatus `json:"status,omitempty"`
}

// GrafanaDashboardList contains a list of GrafanaDashboard
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}

func (d *GrafanaDashboard) Hash() string {
	hash := sha256.New()

	for _, input := range d.Spec.Datasources {
		io.WriteString(hash, input.DatasourceName) // nolint
		io.WriteString(hash, input.InputName)      // nolint
	}

	io.WriteString(hash, d.Spec.Json) // nolint
	hash.Write(d.Spec.GzipJson)
	io.WriteString(hash, d.Spec.Url)              // nolint
	io.WriteString(hash, d.Spec.Jsonnet)          // nolint
	io.WriteString(hash, d.Namespace)             // nolint
	io.WriteString(hash, d.Spec.CustomFolderName) // nolint

	if d.Spec.ConfigMapRef != nil {
		io.WriteString(hash, d.Spec.ConfigMapRef.Name) // nolint
		io.WriteString(hash, d.Spec.ConfigMapRef.Key)  // nolint
	}
	if d.Spec.GzipConfigMapRef != nil {
		io.WriteString(hash, d.Spec.GzipConfigMapRef.Name) // nolint
		io.WriteString(hash, d.Spec.GzipConfigMapRef.Key)  // nolint
	}

	if d.Spec.GrafanaCom != nil {
		io.WriteString(hash, fmt.Sprint((d.Spec.GrafanaCom.Id))) // nolint
		if d.Spec.GrafanaCom.Revision != nil {
			io.WriteString(hash, fmt.Sprint(*d.Spec.GrafanaCom.Revision)) // nolint
		}
	}

	if d.Status.ContentCache != nil {
		hash.Write(d.Status.ContentCache)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (d *GrafanaDashboard) Parse(optional string) (map[string]interface{}, error) {
	var dashboardBytes []byte
	if d.Spec.GzipJson != nil {
		var err error
		dashboardBytes, err = Gunzip(d.Spec.GzipJson)
		if err != nil {
			return nil, err
		}
	} else {
		dashboardBytes = []byte(d.Spec.Json)
	}
	if optional != "" {
		dashboardBytes = []byte(optional)
	}

	parsed := make(map[string]interface{})
	err := json.Unmarshal(dashboardBytes, &parsed)
	return parsed, err
}

func (d *GrafanaDashboard) UID() string {
	content, err := d.Parse("")
	if err == nil {
		// Check if the user has defined an uid and if that's the
		// case, use that
		if content["uid"] != nil && content["uid"] != "" {
			return content["uid"].(string)
		}
	}

	// Use sha1 to keep the hash limit at 40 bytes which is what
	// Grafana allows for UIDs
	return fmt.Sprintf("%x", sha1.Sum([]byte(d.Namespace+d.Name))) // nolint
}

func (d *GrafanaDashboard) GetContentCache(url string) string {
	var cacheDuration time.Duration
	if d.Spec.ContentCacheDuration != nil {
		cacheDuration = d.Spec.ContentCacheDuration.Duration
	}

	return d.Status.getContentCache(url, cacheDuration)
}

// getContentCache returns content cache when the following conditions are met: url is the same, data is not expired, gzipped data is not corrupted
func (s *GrafanaDashboardStatus) getContentCache(url string, cacheDuration time.Duration) string {
	if s.ContentUrl != url {
		return ""
	}

	notExpired := cacheDuration <= 0 || s.ContentTimestamp.Add(cacheDuration).After(time.Now())
	if !notExpired {
		return ""
	}

	cache, err := Gunzip(s.ContentCache)
	if err != nil {
		return ""
	}

	return string(cache)
}

func Gunzip(compressed []byte) ([]byte, error) {
	decoder, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(decoder)
}

func Gzip(content string) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	writer := gzip.NewWriter(buf)
	_, err := writer.Write([]byte(content))
	if err != nil {
		return nil, err
	}
	writer.Close()
	return ioutil.ReadAll(buf)
}
