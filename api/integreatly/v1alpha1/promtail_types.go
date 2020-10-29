package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PromtailSpec defines the desired state of Promtail
type PromtailSpec struct {
	HostPaths                      []string              `json:"hostPaths,omitempty"`
	PodAggregatorLabelSelector     *metav1.LabelSelector `json:"podAggregatorLabelSelector,omitempty"`
	PodAggregatorNamespaceSelector *metav1.LabelSelector `json:"podAggregatorNamespaceSelector,omitempty"`
	InstanceSelector               *metav1.LabelSelector `json:"instanceSelector,omitempty"`
}

// PromtailStatus defines the observed state of Promtail
type PromtailStatus struct {
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Promtail is the Schema for the promtails API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=promtails,scope=Namespaced
type Promtail struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PromtailSpec   `json:"spec,omitempty"`
	Status PromtailStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PromtailList contains a list of Promtail
type PromtailList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Promtail `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Promtail{}, &PromtailList{})
}
