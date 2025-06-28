package v1beta1

import (
	"fmt"
	"strings"
)

type NamespacedResource string

type NamespacedResourceList []NamespacedResource

// +kubebuilder:object:generate=false
type NamespacedResourceImpl interface {
	Exists(namespace string, name string) bool
}

func NewNamespacedResource(namespace, name, identifier string) NamespacedResource {
	return NamespacedResource(fmt.Sprintf("%s/%s/%s", namespace, name, identifier))
}

func (in NamespacedResource) Split() (string, string, string) {
	parts := strings.Split(string(in), "/")
	return parts[0], parts[1], parts[2]
}

func (in NamespacedResourceList) Find(namespace string, name string) (bool, *string) {
	for _, r := range in {
		foundNamespace, foundName, foundUID := r.Split()
		if foundNamespace == namespace && foundName == name {
			return true, &foundUID
		}
	}
	return false, nil
}

func (in NamespacedResourceList) IndexOf(namespace string, name string) int {
	for i, r := range in {
		foundNamespace, foundName, _ := r.Split()
		if foundNamespace == namespace && foundName == name {
			return i
		}
	}
	return -1
}

func (in NamespacedResourceList) Remove(namespace string, name string) NamespacedResourceList {
	resources := NamespacedResourceList{}
	for _, r := range in {
		foundNamespace, foundName, _ := r.Split()
		if foundNamespace == namespace && foundName == name {
			continue
		}
		resources = append(resources, r)
	}
	return resources
}
