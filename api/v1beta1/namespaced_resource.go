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

func (in NamespacedResourceList) Add(namespace string, name string, uid string) NamespacedResourceList {
	resource := NamespacedResource(fmt.Sprintf("%v/%v/%v", namespace, name, uid))
	resources := NamespacedResourceList{resource}
	for _, r := range in {
		if r == resource {
			return in
		}
		resources = append(resources, r)
	}
	return resources
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
