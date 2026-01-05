package autodetect

import (
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type AutoDetect struct {
	dcl discovery.DiscoveryInterface
}

func NewAutoDetect(restConfig *rest.Config) (*AutoDetect, error) {
	dcl, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &AutoDetect{
		dcl: dcl,
	}, nil
}

func (a *AutoDetect) HasAPIGroup(target string) (bool, error) {
	l, err := a.dcl.ServerGroups()
	if err != nil {
		return false, err
	}

	isFound := slices.ContainsFunc(l.Groups, func(g metav1.APIGroup) bool {
		return g.Name == target
	})

	return isFound, nil
}

func (a *AutoDetect) HasKind(apiVersion, kind string) (bool, error) {
	l, err := a.dcl.ServerResourcesForGroupVersion(apiVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	isFound := slices.ContainsFunc(l.APIResources, func(k metav1.APIResource) bool {
		return k.Kind == kind
	})

	return isFound, nil
}

// Tests for the presence of the `route.openshift.io` api group as an indicator of if we're running in OpenShift
func (a *AutoDetect) IsOpenshift() (bool, error) {
	return a.HasAPIGroup("route.openshift.io")
}

// Tests if the HTTPRoute CRD is present
func (a *AutoDetect) HasHTTPRouteCRD() (bool, error) {
	return a.HasKind("gateway.networking.k8s.io/v1", "HTTPRoute")
}
