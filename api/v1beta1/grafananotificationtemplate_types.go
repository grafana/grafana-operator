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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaNotificationTemplateSpec defines the desired state of GrafanaNotificationTemplate
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable)))", message="spec.editable is immutable"
type GrafanaNotificationTemplateSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Template name
	Name string `json:"name"`

	// Template content
	Template string `json:"template,omitempty"`

	// Whether to enable or disable editing of the notification template in Grafana UI
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.editable is immutable"
	// +optional
	Editable *bool `json:"editable,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaNotificationTemplate is the Schema for the GrafanaNotificationTemplate API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaNotificationTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationTemplateSpec `json:"spec"`
	Status GrafanaCommonStatus             `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaNotificationTemplate)(nil)

func (in *GrafanaNotificationTemplate) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaNotificationTemplate) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaNotificationTemplate) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaNotificationTemplate) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaNotificationTemplate) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.Spec.Name)
}

func (in *GrafanaNotificationTemplate) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaNotificationTemplate) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

//+kubebuilder:object:root=true

// GrafanaNotificationTemplateList contains a list of GrafanaNotificationTemplate
type GrafanaNotificationTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationTemplate `json:"items"`
}

func (in *GrafanaNotificationTemplateList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationTemplate{}, &GrafanaNotificationTemplateList{})
}
