/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the Licens_ at

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

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GrafanaContactPointSpec defines the desired state of GrafanaContactPoint
type GrafanaContactPointSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Receivers are grouped under the same ContactPoint using the Name
	// Defaults to the name of the CR
	// +optional
	// +kubebuilder:validation:type=string
	Name string `json:"name,omitempty"`

	// List of receivers that Grafana will fan out notifications to
	// +optional
	// +kubebuilder:validation:MaxItems=99
	Receivers []ContactPointReceiver `json:"receivers,omitempty"`

	// Deprecated: define the receiver under .spec.receivers[]
	// Manually specify the UID the Contact Point is created with. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	CustomUID string `json:"uid,omitempty"`

	// Deprecated: define the receiver under .spec.receivers[]
	// Will be removed in a later version
	// +optional
	DisableResolveMessage bool `json:"disableResolveMessage,omitempty"`

	// Deprecated: define the receiver under .spec.receivers[]
	// Will be removed in a later version
	// +optional
	Settings *apiextensions.JSON `json:"settings,omitempty"`

	// Deprecated: define the receiver under .spec.receivers[]
	// Will be removed in a later version
	// +kubebuilder:validation:MaxItems=99
	ValuesFrom []ValueFrom `json:"valuesFrom,omitempty"`

	// Deprecated: define the receiver under .spec.receivers[]
	// Will be removed in a later version
	// +optional
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type,omitempty"`
}

// Represents an integration to external services that receive Grafana notifications
type ContactPointReceiver struct {
	// Manually specify the UID the Contact Point is created with. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40
	// +optional
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	CustomUID string `json:"uid,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`

	// +optional
	DisableResolveMessage bool `json:"disableResolveMessage,omitempty"`

	Settings *apiextensions.JSON `json:"settings"`

	// +kubebuilder:validation:MaxItems=99
	ValuesFrom []ValueFrom `json:"valuesFrom,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaContactPoint is the Schema for the grafanacontactpoints API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaContactPoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaContactPointSpec `json:"spec"`
	Status GrafanaCommonStatus     `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaContactPoint)(nil)

//+kubebuilder:object:root=true

// GrafanaContactPointList contains a list of GrafanaContactPoint
type GrafanaContactPointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaContactPoint `json:"items"`
}

func (in *GrafanaContactPointList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

// Wrapper around Name or default metadata.name
func (in *GrafanaContactPoint) NameFromSpecOrMeta() string {
	if in.Spec.Name != "" {
		return in.Spec.Name
	}

	return in.Name
}

// Wrapper around receivers[].CustomUID or metadata.uid/idx
func (in *ContactPointReceiver) CustomUIDOrUID(metaUID types.UID, idx int) string {
	if in.CustomUID != "" {
		return in.CustomUID
	}

	// UID/idx is stable and allows overriding
	return fmt.Sprintf("%s_%d", string(metaUID), idx)
}

func (in *GrafanaContactPoint) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaContactPoint) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaContactPoint) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaContactPoint) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaContactPoint) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaContactPoint) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.Spec.Name)
}

func init() {
	SchemeBuilder.Register(&GrafanaContactPoint{}, &GrafanaContactPointList{})
}
