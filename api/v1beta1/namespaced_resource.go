package v1beta1

import (
	"fmt"
	"slices"
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
	i := in.IndexOf(namespace, name)

	if i == -1 {
		return false, nil
	}

	_, _, uid := in[i].Split()
	return true, &uid
}

func (in NamespacedResourceList) IndexOf(namespace string, name string) int {
	i := slices.IndexFunc(in, func(r NamespacedResource) bool {
		p := fmt.Sprintf("%s/%s/", namespace, name)
		return strings.HasPrefix(string(r), p)
	})

	return i
}

func (in NamespacedResourceList) Remove(namespace string, name string) NamespacedResourceList {
	i := in.IndexOf(namespace, name)

	if i == -1 {
		return in
	}

	// Swapback delete and return slice
	// Does not preserve order
	in[i] = in[len(in)-1]
	return in[:len(in)-1]
}
