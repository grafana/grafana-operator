/*
Copyright 2022.

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

package v1beta1

import (
	"crypto/sha256"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaFolderSpec defines the desired state of GrafanaFolder
type GrafanaFolderSpec struct {
	Json string `json:"json,omitempty"`

	// selects Grafanas for import
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
}

// GrafanaFolderStatus defines the observed state of GrafanaFolder
type GrafanaFolderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hash string `json:"hash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaFolder is the Schema for the grafanafolders API
type GrafanaFolder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaFolderSpec   `json:"spec,omitempty"`
	Status GrafanaFolderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaFolderList contains a list of GrafanaFolder
type GrafanaFolderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaFolder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaFolder{}, &GrafanaFolderList{})
}

func (in *GrafanaFolderList) Find(namespace string, name string) *GrafanaFolder {
	for _, folder := range in.Items {
		if folder.Namespace == namespace && folder.Name == name {
			return &folder
		}
	}
	return nil
}

func (in *GrafanaFolder) Hash() string {
	hash := sha256.New()
	hash.Write([]byte(in.Spec.Json))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (in *GrafanaFolder) Unchanged() bool {
	return in.Hash() == in.Status.Hash
}
