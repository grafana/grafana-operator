package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type FolderReferencer interface {
	FolderRef() string
	FolderUID() string
	FolderNamespace() string
	ConditionsResource
}

type ConditionsResource interface {
	Conditions() *[]metav1.Condition
	CurrentGeneration() int64
}
