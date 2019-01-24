package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Json string `json:"json"`
	Name string `json:"name"`
}

// GrafanaDashboardStatus defines the observed state of GrafanaDashboard
type GrafanaDashboardStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Created bool `json: "created"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDashboard is the Schema for the grafanadashboards API
// +k8s:openapi-gen=true
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDashboardSpec   `json:"spec,omitempty"`
	Status GrafanaDashboardStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDashboardList contains a list of GrafanaDashboard
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}
