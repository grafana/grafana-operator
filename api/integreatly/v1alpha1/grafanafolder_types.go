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
	"crypto/sha256"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GrafanaPermissionItem struct {
	PermissionTargetType string `json:"permissionTargetType"`
	PermissionTarget     string `json:"permissionTarget"`
	PermissionLevel      int    `json:"permissionLevel"`
}

type GrafanaFolderSpec struct {
	// FolderName is the display-name of the folder and must match CustomFolderName of any GrafanaDashboard you want to put in
	FolderName string `json:"title"`

	// FolderPermissions shall contain the _complete_ permissions for the folder.
	// Any permission not listed here, will be removed from the folder.
	FolderPermissions []GrafanaPermissionItem `json:"permissions,omitempty"`
}

// GrafanaFolder is the Schema for the grafana folders and folderpermissions API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type GrafanaFolder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GrafanaFolderSpec `json:"spec,omitempty"`
}

// GrafanaFolderList contains a list of GrafanaFolder
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type GrafanaFolderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaFolder `json:"items"`
}

// GrafanaFolderRef is used to keep a folder reference without having access to the folder-struct itself
type GrafanaFolderRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Hash      string `json:"hash"`
}

func init() {
	SchemeBuilder.Register(&GrafanaFolder{}, &GrafanaFolderList{})
}

func (f *GrafanaFolder) Hash() string {
	hash := sha256.New()

	io.WriteString(hash, f.Spec.FolderName) // nolint
	io.WriteString(hash, f.Namespace)       // nolint

	for _, p := range f.Spec.FolderPermissions {
		io.WriteString(hash, p.PermissionTarget)            // nolint
		io.WriteString(hash, p.PermissionTargetType)        // nolint
		io.WriteString(hash, fmt.Sprint(p.PermissionLevel)) // nolint
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (f *GrafanaFolder) GetPermissions() []*GrafanaPermissionItem {
	var permissions = make([]*GrafanaPermissionItem, 0, len(f.Spec.FolderPermissions))
	for _, p := range f.Spec.FolderPermissions {
		var p2 = p // ensure allocated memory for current item
		permissions = append(permissions, &p2)
	}

	return permissions
}
