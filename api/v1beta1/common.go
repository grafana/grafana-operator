package v1beta1

import v1 "k8s.io/api/core/v1"

type ValueFrom struct {
	TargetPath string          `json:"targetPath"`
	ValueFrom  ValueFromSource `json:"valueFrom"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.configMapKeyRef) && !has(self.secretKeyRef)) || (!has(self.configMapKeyRef) && has(self.secretKeyRef))", message="Either configMapKeyRef or secretKeyRef must be set"
type ValueFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
