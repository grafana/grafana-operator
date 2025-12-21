// Package autodetect is for auto-detecting traits from the environment (platform, APIs, ...).
package autodetect

import (
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
	apiList, err := a.dcl.ServerGroups()
	if err != nil {
		return false, err
	}

	apiGroups := apiList.Groups
	for _, group := range apiGroups {
		if group.Name == target {
			return true, nil
		}
	}

	return false, nil
}

// Tests for the presence of the `route.openshift.io` api group as an indicator of if we're running in OpenShift
func (a *autoDetect) IsOpenshift() (bool, error) {
	return a.hasAPIGroup("route.openshift.io")
}

// Tests if the GatewayAPI CRDs are present
func (a *autoDetect) HasGatewayAPI() (bool, error) {
	return a.hasAPIGroup("gateway.networking.k8s.io")
}
