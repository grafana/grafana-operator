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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID)))", message="Only one of folderUID or folderRef can be declared at the same time"
// +kubebuilder:validation:XValidation:rule="(has(self.folder) && !(has(self.folderRef) || has(self.folderUID))) || !(has(self.folder))", message="folder field cannot be set when folderUID or folderRef is already declared"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaDashboardSpec struct {
	GrafanaCommonSpec  `json:",inline"`
	GrafanaContentSpec `json:",inline"`

	// folder assignment for dashboard
	// +optional
	FolderTitle string `json:"folder,omitempty"`

	// UID of the target folder for this dashboard
	// +optional
	FolderUID string `json:"folderUID,omitempty"`

	// Name of a `GrafanaFolder` resource in the same namespace
	// +optional
	FolderRef string `json:"folderRef,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`
}

// GrafanaDashboardStatus defines the observed state of GrafanaDashboard
type GrafanaDashboardStatus struct {
	GrafanaCommonStatus  `json:",inline"`
	GrafanaContentStatus `json:",inline"`

	// The dashboard instanceSelector can't find matching grafana instances
	NoMatchingInstances bool `json:"NoMatchingInstances,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaDashboard is the Schema for the grafanadashboards API
// +kubebuilder:printcolumn:name="No matching instances",type="boolean",JSONPath=".status.NoMatchingInstances",description=""
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDashboardSpec   `json:"spec"`
	Status GrafanaDashboardStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaDashboard)(nil)

//+kubebuilder:object:root=true

// GrafanaDashboardList contains a list of GrafanaDashboard
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDashboard `json:"items"`
}

// FolderRef implements FolderReferencer.
func (in *GrafanaDashboard) FolderRef() string {
	return in.Spec.FolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaDashboard) FolderUID() string {
	return in.Spec.FolderUID
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaDashboard) FolderNamespace() string {
	return in.Namespace
}

// Conditions implements FolderReferencer.
func (in *GrafanaDashboard) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// CurrentGeneration implements FolderReferencer.
func (in *GrafanaDashboard) CurrentGeneration() int64 {
	return in.Generation
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaDashboard) GrafanaContentSpec() *GrafanaContentSpec {
	return &in.Spec.GrafanaContentSpec
}

// GrafanaContentSpec implements GrafanaContentResource
func (in *GrafanaDashboard) GrafanaContentStatus() *GrafanaContentStatus {
	return &in.Status.GrafanaContentStatus
}

var _ GrafanaContentResource = &GrafanaDashboard{}

func (in *GrafanaDashboardList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func (in *GrafanaDashboard) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaDashboard) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaDashboard) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaDashboard) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaDashboard) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

func (in *GrafanaDashboard) NamespacedResource(uid string) NamespacedResource {
	// Not enough context to call content.CustomUIDOrUID(uid).
	// Hence, use uid from args as the caller has more context
	return NewNamespacedResource(in.Namespace, in.Name, uid)
}

func (in *GrafanaDashboard) GetPluginConfigMapKey() string {
	return GetPluginConfigMapKey("dashboard", &in.ObjectMeta)
}

func (in *GrafanaDashboard) GetPluginConfigMapDeprecatedKey() string {
	return fmt.Sprintf("%v-dashboard", in.Name)
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}
