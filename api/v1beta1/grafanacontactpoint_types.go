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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaContactPointSpec defines the desired state of GrafanaContactPoint
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaContactPointSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Manually specify the UID the Contact Point is created with. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	CustomUID string `json:"uid,omitempty"`

	// +optional
	DisableResolveMessage bool `json:"disableResolveMessage,omitempty"`

	// +kubebuilder:validation:type=string
	Name string `json:"name"`

	Settings *apiextensions.JSON `json:"settings"`

	// +kubebuilder:validation:MaxItems=99
	ValuesFrom []ValueFrom `json:"valuesFrom,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`
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

// Wrapper around CustomUID or default metadata.uid
func (in *GrafanaContactPoint) CustomUIDOrUID() string {
	if in.Spec.CustomUID != "" {
		return in.Spec.CustomUID
	}

	return string(in.UID)
}

func (in *GrafanaContactPoint) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaContactPoint) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaContactPoint) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaContactPoint) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaContactPoint) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.CustomUIDOrUID())
}

func init() {
	SchemeBuilder.Register(&GrafanaContactPoint{}, &GrafanaContactPointList{})
}
