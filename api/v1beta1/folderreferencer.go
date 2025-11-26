package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:generate=false
type FolderReferencer interface {
	Conditions() *[]metav1.Condition
	FolderNamespace() string
	FolderRef() string
	FolderUID() string
	GetGeneration() int64
}
