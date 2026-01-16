/*
Copyright 2026.

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
	"encoding/json"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaManifestSpec defines the desired state of a GrafanaManifest
type GrafanaManifestSpec struct {
	GrafanaCommonSpec `json:",inline"`
	Template          GrafanaManifestTemplate `json:"template"`
}

type GrafanaManifestTemplate struct {
	RequiredTypeMeta `json:",inline"`
	Metadata         RequiredObjectMeta `json:"metadata"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

func (t *GrafanaManifestTemplate) ToUnstructured() *unstructured.Unstructured {
	enc, _ := json.Marshal(t) //nolint:errcheck // cannot fail as it's from the serialized kubernetes resource
	out := &unstructured.Unstructured{}
	_ = json.Unmarshal(enc, out) //nolint:errcheck // unmarshaling previously marshaled object with required fields

	return out
}

// GrafanaManifestStatus defines the observed state of GrafanaManifest
type GrafanaManifestStatus struct {
	GrafanaCommonStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaManifest is the Schema for the grafana manifests
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.template.kind",description=""
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaManifestSpec   `json:"spec"`
	Status GrafanaManifestStatus `json:"status,omitempty"`
}

// Conditions implements [CommonResource].
func (in *GrafanaManifest) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

var _ CommonResource = (*GrafanaManifest)(nil)

// GetGrafanaUID selects a UID to be used for Grafana API requests (preference: spec.CustomUID -> metadata.uid)
func (in *GrafanaManifest) GetGrafanaUID() string {
	return string(in.UID)
}

//+kubebuilder:object:root=true

// GrafanaManifestList contains a list of GrafanaManifest
type GrafanaManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaManifest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaManifest{}, &GrafanaManifestList{})
}

func (in *GrafanaManifestList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func (in *GrafanaManifest) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaManifest) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaManifest) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaManifest) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaManifest) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

func (in *GrafanaManifest) NamespacedResource(uid string) NamespacedResource {
	// .GetGrafanaUID() can be wrong when the fallback to search is used.
	// Hence, use uid from args as the caller has more context
	// TODO Remove uid arg along with the search fallback
	return NewNamespacedResource(in.Namespace, in.Name, uid)
}
