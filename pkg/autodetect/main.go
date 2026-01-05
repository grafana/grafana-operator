// Package autodetect is for auto-detecting traits from the environment (platform, APIs, ...).
package autodetect

import (
	"slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

var _ AutoDetect = (*autoDetect)(nil)

// AutoDetect provides an assortment of routines that auto-detect traits based on the runtime.
type AutoDetect interface {
	IsOpenshift() (bool, error)
	HasGatewayAPI() (bool, error)
}

type autoDetect struct {
	dcl discovery.DiscoveryInterface
}

// New creates a new auto-detection worker, using the given client when talking to the current cluster.
func New(restConfig *rest.Config) (AutoDetect, error) {
	dcl, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		// it's pretty much impossible to get into this problem, as most of the
		// code branches from the previous call just won't fail at all,
		// but let's handle this error anyway...
		return nil, err
	}

	return &autoDetect{
		dcl: dcl,
	}, nil
}

func (a *autoDetect) hasAPIGroup(target string) (bool, error) {
	l, err := a.dcl.ServerGroups()
	if err != nil {
		return false, err
	}

	isFound := slices.ContainsFunc(l.Groups, func(g metav1.APIGroup) bool {
		return g.Name == target
	})

	return isFound, nil
}

func (a *autoDetect) hasKind(apiVersion, kind string) (bool, error) {
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
func (a *autoDetect) IsOpenshift() (bool, error) {
	return a.hasAPIGroup("route.openshift.io")
}

// Tests if the GatewayAPI CRDs are present
func (a *autoDetect) HasGatewayAPI() (bool, error) {
	return a.hasKind("gateway.networking.k8s.io/v1", "HTTPRoute")
}
