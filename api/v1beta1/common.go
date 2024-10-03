package v1beta1

import v1 "k8s.io/api/core/v1"

type ValueFrom struct {
	TargetPath string          `json:"targetPath"`
	ValueFrom  ValueFromSource `json:"valueFrom"`
}

type ValueFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
