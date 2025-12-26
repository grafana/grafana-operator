package v1beta1

import (
	"fmt"
	"slices"
	"strings"
)

// +kubebuilder:object:generate=false
type NamespacedResourceChecker interface {
	Exists(namespace string, name string) bool
}

type NamespacedResource string

func NewNamespacedResource(namespace, name, identifier string) NamespacedResource {
	return NamespacedResource(fmt.Sprintf("%s/%s/%s", namespace, name, identifier))
}

func (in NamespacedResource) Split() (namespace, name, identifier string) {
	parts := strings.Split(string(in), "/")
	return parts[0], parts[1], parts[2]
}

type NamespacedResourceList []NamespacedResource

func (in NamespacedResourceList) Find(namespace, name string) (found bool, identifier *string) {
	i := in.IndexOf(namespace, name)

	if i == -1 {
		return false, nil
	}

	_, _, uid := in[i].Split()

	return true, &uid
}

func (in NamespacedResourceList) IndexOf(namespace, name string) int {
	p := fmt.Sprintf("%s/%s/", namespace, name)

	i := slices.IndexFunc(in, func(r NamespacedResource) bool {
		return strings.HasPrefix(string(r), p)
	})

	return i
}

func (in NamespacedResourceList) RemoveEntries(toRemove *NamespacedResourceList) NamespacedResourceList {
	resources := slices.DeleteFunc(in.DeepCopy(), func(r NamespacedResource) bool {
		return slices.Contains(*toRemove, r)
	})

	return resources
}
