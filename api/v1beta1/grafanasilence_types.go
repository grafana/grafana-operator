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

// SilenceIDAnnotation stores the Grafana-assigned silence IDs as a JSON map of
// "<instance namespace>/<instance name>" to silence ID. The operator writes the IDs back
// after creating a silence; users can pre-populate an entry to adopt (import) an existing
// silence instead of creating a new one.
const SilenceIDAnnotation = "grafana.integreatly.org/silence-id"

// GrafanaSilenceSpec defines the desired state of GrafanaSilence
type GrafanaSilenceSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Matchers used to select the alerts that should be silenced
	// +kubebuilder:validation:MinItems=1
	Matchers []*SilenceMatcher `json:"matchers"`

	// StartsAt is the time the silence starts taking effect (in UTC, RFC3339)
	StartsAt metav1.Time `json:"startsAt"`

	// EndsAt is the time the silence expires (in UTC, RFC3339)
	EndsAt metav1.Time `json:"endsAt"`

	// Comment describing the reason for the silence
	Comment string `json:"comment"`

	// CreatedBy is the author attributed to the silence
	// +optional
	// +kubebuilder:default="grafana-operator"
	CreatedBy string `json:"createdBy,omitempty"`
}

type SilenceMatcher struct {
	// The name of the label to match against
	Name string `json:"name"`

	// The value to match against
	Value string `json:"value"`

	// Whether to interpret the value as a regular expression
	// +optional
	IsRegex bool `json:"isRegex,omitempty"`

	// Whether the matcher is an equality matcher (true) or a negative matcher (false)
	// +optional
	// +kubebuilder:default=true
	IsEqual bool `json:"isEqual"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaSilence is the Schema for the GrafanaSilence API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaSilence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaSilenceSpec  `json:"spec"`
	Status GrafanaCommonStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaSilence)(nil)

func (in *GrafanaSilence) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaSilence) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaSilence) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaSilence) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaSilence) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.Name)
}

func (in *GrafanaSilence) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaSilence) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

//+kubebuilder:object:root=true

// GrafanaSilenceList contains a list of GrafanaSilence
type GrafanaSilenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaSilence `json:"items"`
}

func (in *GrafanaSilenceList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}
