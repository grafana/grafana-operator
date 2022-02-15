package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GrafanaNotificationChannelKind = "GrafanaNotificationChannel"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaNotificationChannelSpec defines the desired state of GrafanaNotificationChannel
type GrafanaNotificationChannelSpec struct {
	Json string `json:"json"`
	Name string `json:"name"`
}

// GrafanaNotificationChannelStatus defines the observed state of GrafanaNotificationChannel
type GrafanaNotificationChannelStatus struct {
	Phase   StatusPhase `json:"phase"`
	UID     string      `json:"uid"`
	ID      uint        `json:"id"`
	Message string      `json:"message"`
	Hash    string      `json:"hash"`
}

// Used to keep a notification channel reference without having access to the notification channel
// struct itself
type GrafanaNotificationChannelRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
	Hash      string `json:"hash"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// GrafanaNotificationChannel is the Schema for the GrafanaNotificationChannels API
type GrafanaNotificationChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationChannelSpec   `json:"spec,omitempty"`
	Status GrafanaNotificationChannelStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// GrafanaNotificationChannelList contains a list of GrafanaNotificationChannel
type GrafanaNotificationChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationChannel `json:"items"`
}

type GrafanaNotificationChannelStatusMessage struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationChannel{}, &GrafanaNotificationChannelList{})
}
