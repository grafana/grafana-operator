package v1beta1

import (
	v1 "k8s.io/api/networking/v1"
)

// +kubebuilder:object:generate=true

// ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta).
type ObjectMeta struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// +kubebuilder:object:generate=true

// IngressNetworkingV1 is a subset of [Ingress in k8s.io/api/networking/v1beta1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingress-v1-networking-k8s-io).
type IngressNetworkingV1 struct {
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`
	// Kubernetes [Ingress Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingressclassspec-v1-networking-k8s-io)
	Spec v1.IngressSpec `json:"spec,omitempty"`
}
