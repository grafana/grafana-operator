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

	// Manually specify the UID the Contact Point is created with
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	CustomUID string `json:"uid,omitempty"`

	// +optional
	DisableResolveMessage bool `json:"disableResolveMessage,omitempty"`

	// +kubebuilder:validation:type=string
	Name string `json:"name"`

	Settings *apiextensions.JSON `json:"settings"`

	// +kubebuilder:validation:MaxItems=99
	ValuesFrom []ValueFrom `json:"valuesFrom,omitempty"`

	// +kubebuilder:validation:Enum=alertmanager;prometheus-alertmanager;dingding;discord;email;googlechat;kafka;line;opsgenie;pagerduty;pushover;sensugo;sensu;slack;teams;telegram;threema;victorops;webex;webhook;wecom;hipchat;oncall
	Type string `json:"type,omitempty"`
}

// GrafanaContactPointStatus defines the observed state of GrafanaContactPoint
type GrafanaContactPointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaContactPoint is the Schema for the grafanacontactpoints API
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaContactPoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaContactPointSpec   `json:"spec,omitempty"`
	Status GrafanaContactPointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaContactPointList contains a list of GrafanaContactPoint
type GrafanaContactPointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaContactPoint `json:"items"`
}

// Wrapper around CustomUID or default metadata.uid
func (in *GrafanaContactPoint) CustomUIDOrUID() string {
	if in.Spec.CustomUID != "" {
		return in.Spec.CustomUID
	}
	return string(in.ObjectMeta.UID)
}

func init() {
	SchemeBuilder.Register(&GrafanaContactPoint{}, &GrafanaContactPointList{})
}
