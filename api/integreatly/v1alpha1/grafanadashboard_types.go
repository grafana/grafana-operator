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
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Json             string                       `json:"json"`
	Jsonnet          string                       `json:"jsonnet,omitempty"`
	Plugins          PluginList                   `json:"plugins,omitempty"`
	Url              string                       `json:"url,omitempty"`
	ConfigMapRef     *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`
	Datasources      []GrafanaDashboardDatasource `json:"datasources,omitempty"`
	CustomFolderName string                       `json:"customFolderName,omitempty"`
}
type GrafanaDashboardDatasource struct {
	InputName      string `json:"inputName"`
	DatasourceName string `json:"datasourceName"`
}

// Used to keep a dashboard reference without having access to the dashboard
// struct itself
type GrafanaDashboardRef struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	UID        string `json:"uid"`
	Hash       string `json:"hash"`
	FolderId   *int64 `json:"folderId"`
	FolderName string `json:"folderName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// GrafanaDashboard is the Schema for the grafanadashboards API
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GrafanaDashboardSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// GrafanaDashboardList contains a list of GrafanaDashboard
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
		io.WriteString(hash, input.DatasourceName)
		io.WriteString(hash, input.InputName)
	}

	io.WriteString(hash, d.Spec.Json)
	io.WriteString(hash, d.Spec.Url)
	io.WriteString(hash, d.Spec.Jsonnet)
	io.WriteString(hash, d.Namespace)
	io.WriteString(hash, d.Spec.CustomFolderName)

	if d.Spec.ConfigMapRef != nil {
		io.WriteString(hash, d.Spec.ConfigMapRef.Name)
		io.WriteString(hash, d.Spec.ConfigMapRef.Key)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (d *GrafanaDashboard) Parse(optional string) (map[string]interface{}, error) {
	var dashboardBytes = []byte(d.Spec.Json)
	if optional != "" {
		dashboardBytes = []byte(optional)
	}

	var parsed = make(map[string]interface{})
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
	return fmt.Sprintf("%x", sha1.Sum([]byte(d.Namespace+d.Name)))
}
