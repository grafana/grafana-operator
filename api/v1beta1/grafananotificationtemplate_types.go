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

// GrafanaNotificationTemplateSpec defines the desired state of GrafanaNotificationTemplate
type GrafanaNotificationTemplateSpec struct {
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:default="10m"
	ResyncPeriod metav1.Duration `json:"resyncPeriod,omitempty"`

	// selects Grafanas for import
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector"`

	Name string `json:"name"`

	Template string `json:"template,omitempty"`

	// Whether to enable or disable editing of the notification template in Grafana UI
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	Editable *bool `json:"editable,omitempty"`

	// +optional
	AllowCrossNamespaceImport *bool `json:"allowCrossNamespaceImport,omitempty"`
}

// GrafanaNotificationTemplateStatus defines the observed state of GrafanaNotificationTemplate
type GrafanaNotificationTemplateStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaNotificationTemplate is the Schema for the GrafanaNotificationTemplate API
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaNotificationTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationTemplateSpec   `json:"spec,omitempty"`
	Status GrafanaNotificationTemplateStatus `json:"status,omitempty"`
}

func (np *GrafanaNotificationTemplate) NamespacedResource() string {
	return fmt.Sprintf("%v/%v/%v", np.ObjectMeta.Namespace, np.ObjectMeta.Name, np.ObjectMeta.UID)
}

//+kubebuilder:object:root=true

// GrafanaNotificationTemplateList contains a list of GrafanaNotificationTemplate
type GrafanaNotificationTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationTemplate{}, &GrafanaNotificationTemplateList{})
}
