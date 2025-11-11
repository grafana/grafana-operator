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

func (a *autoDetect) HasGatewayAPI() (bool, error) {
	apiGroupList, err := a.dcl.ServerGroups()
	if err != nil {
		return false, err
	}
	apiGroups := apiGroupList.Groups
	for _, apiGroup := range apiGroups {
		if apiGroup.Name == "gateway.networking.k8s.io" {
			return true, nil
		}
	}
	return false, nil
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

// Platform returns the detected platform this operator is running on. Possible values: Kubernetes, OpenShift.
func (a *autoDetect) IsOpenshift() (bool, error) {
	apiList, err := a.dcl.ServerGroups()
	if err != nil {
		return false, err
	}

	apiGroups := apiList.Groups
	for i := range apiGroups {
		if apiGroups[i].Name == "route.openshift.io" {
			return true, nil
		}
	}

	return false, nil
}
