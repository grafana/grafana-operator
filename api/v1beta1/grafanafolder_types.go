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
	"time"

	operatorapi "github.com/grafana/grafana-operator/v5/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaFolderSpec defines the desired state of GrafanaFolder
// +kubebuilder:validation:XValidation:rule="(has(self.parentFolderUID) && !(has(self.parentFolderRef))) || (has(self.parentFolderRef) && !(has(self.parentFolderUID))) || !(has(self.parentFolderRef) && (has(self.parentFolderUID)))", message="Only one of parentFolderUID or parentFolderRef can be set"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaFolderSpec struct {
	// Manually specify the UID the Folder is created with
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	CustomUID string `json:"uid,omitempty"`

	// Display name of the folder in Grafana
	// +optional
	Title string `json:"title,omitempty"`

	// Raw json with folder permissions, potentially exported from Grafana
	// +optional
	Permissions string `json:"permissions,omitempty"`

	// Selects Grafanas for import
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector"`

	// Enable matching Grafana instances outside the current namespace
	// +optional
	AllowCrossNamespaceImport *bool `json:"allowCrossNamespaceImport,omitempty"`

	// UID of the folder in which the current folder should be created
	// +optional
	ParentFolderUID string `json:"parentFolderUID,omitempty"`

	// Reference to an existing GrafanaFolder CR in the same namespace
	// +optional
	ParentFolderRef string `json:"parentFolderRef,omitempty"`

	// How often the folder is synced, defaults to 5m if not set
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:default="5m"
	ResyncPeriod string `json:"resyncPeriod,omitempty"`
}

// GrafanaFolderStatus defines the observed state of GrafanaFolder
type GrafanaFolderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hash string `json:"hash,omitempty"`
	// The folder instanceSelector can't find matching grafana instances
	NoMatchingInstances bool `json:"NoMatchingInstances,omitempty"`
	// Last time the folder was resynced
	LastResync metav1.Time        `json:"lastResync,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaFolder is the Schema for the grafanafolders API
// +kubebuilder:printcolumn:name="No matching instances",type="boolean",JSONPath=".status.NoMatchingInstances",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaFolder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaFolderSpec   `json:"spec,omitempty"`
	Status GrafanaFolderStatus `json:"status,omitempty"`
}

// Conditions implements FolderReferencer.
func (in *GrafanaFolder) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// CurrentGeneration implements FolderReferencer.
func (in *GrafanaFolder) CurrentGeneration() int64 {
	return in.Generation
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaFolder) FolderNamespace() string {
	return in.Namespace
}

// FolderRef implements FolderReferencer.
func (in *GrafanaFolder) FolderRef() string {
	return in.Spec.ParentFolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaFolder) FolderUID() string {
	return in.Spec.ParentFolderUID
}

// Wrapper around CustomUID or default metadata.uid
func (in *GrafanaFolder) CustomUIDOrUID() string {
	if in.Spec.CustomUID != "" {
		return in.Spec.CustomUID
	}
	return string(in.ObjectMeta.UID)
}

var _ operatorapi.FolderReferencer = (*GrafanaFolder)(nil)

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
	hash.Write([]byte(in.Spec.Title))
	hash.Write([]byte(in.Spec.Permissions))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (in *GrafanaFolder) Unchanged() bool {
	return in.Hash() == in.Status.Hash
}

func (in *GrafanaFolder) IsAllowCrossNamespaceImport() bool {
	if in.Spec.AllowCrossNamespaceImport != nil {
		return *in.Spec.AllowCrossNamespaceImport
	}
	return false
}

func (in *GrafanaFolder) GetTitle() string {
	if in.Spec.Title != "" {
		return in.Spec.Title
	}

	return in.Name
}

func (in *GrafanaFolder) GetResyncPeriod() time.Duration {
	if in.Spec.ResyncPeriod == "" {
		in.Spec.ResyncPeriod = DefaultResyncPeriod
		return in.GetResyncPeriod()
	}

	duration, err := time.ParseDuration(in.Spec.ResyncPeriod)
	if err != nil {
		in.Spec.ResyncPeriod = DefaultResyncPeriod
		return in.GetResyncPeriod()
	}

	return duration
}

func (in *GrafanaFolder) ResyncPeriodHasElapsed() bool {
	deadline := in.Status.LastResync.Add(in.GetResyncPeriod())
	return time.Now().After(deadline)
}
