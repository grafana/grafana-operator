package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DynamicClient struct {
	dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
}

func NewDynamicClient(ctx context.Context, cl client.Client, cr *v1beta1.Grafana) (*DynamicClient, error) {
	config, err := restConfigFor(ctx, cl, cr)
	if err != nil {
		return nil, fmt.Errorf("building rest config for client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("building k8s client: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("building discovery client: %w", err)
	}

	return &DynamicClient{
		Interface:       dynamicClient,
		discoveryClient: dc,
	}, nil
}

func (c *DynamicClient) LookupGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	_, resources, err := c.discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to discover groups and resources: %w", err)
	}

	for _, res := range resources {
		if res.GroupVersion != apiVersion {
			continue
		}

		gv, err := schema.ParseGroupVersion(res.GroupVersion)
		if err != nil {
			return schema.GroupVersionResource{}, fmt.Errorf("failed to parse groupversion returned by server: %w", err)
		}

		for _, api := range res.APIResources {
			// skip subresources and unrelated kinds
			if strings.Contains(api.Name, "/") || api.Kind != kind {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: api.Name,
			}

			return gvr, nil
		}
	}

	return schema.GroupVersionResource{}, errors.New("group version not found")
}
