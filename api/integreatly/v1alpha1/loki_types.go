package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LokiSpec defines the desired state of Loki
type LokiSpec struct {
	// When set, refer to unmamnaged Loki instance and do not create a managed one
	External *LokiExternal `json:"external,omitempty"`
}

type LokiExternal struct {
	Url string `json:"url,omitempty"`
}

// LokiStatus defines the observed state of Loki
type LokiStatus struct {
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
	Url     string      `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Loki is the Schema for the lokis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=lokis,scope=Namespaced
type Loki struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LokiSpec   `json:"spec,omitempty"`
	Status LokiStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LokiList contains a list of Loki
type LokiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Loki `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Loki{}, &LokiList{})
}
