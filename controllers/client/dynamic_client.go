package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// gvrCache caches discovery of endpoints for apiVersion/kind pairs.
// It is safe to share across instances as the schema should be the same across
// all installations due to the apis being scoped & versioned
var gvrCache sync.Map

type DynamicClient struct {
	dynamic.Interface
	discoveryClient          *discovery.DiscoveryClient
	defaultResourceNamespace string
}

func instanceNamespace(instance *v1beta1.Grafana) string {
	if instance.Spec.External != nil && instance.Spec.External.TenantNamespace != "" {
		return instance.Spec.External.TenantNamespace
	}

	return "default"
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
		Interface:                dynamicClient,
		discoveryClient:          dc,
		defaultResourceNamespace: instanceNamespace(cr),
	}, nil
}

func (c *DynamicClient) LookupGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	gvrKey := fmt.Sprintf("%s.%s", apiVersion, kind)

	cached, ok := gvrCache.Load(gvrKey)
	if ok {
		return cached.(schema.GroupVersionResource), nil //nolint:errcheck
	}

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

			gvrCache.Store(gvrKey, gvr)

			return gvr, nil
		}
	}

	return schema.GroupVersionResource{}, errors.New("group version not found")
}

func (c *DynamicClient) NamespaceFor(obj *unstructured.Unstructured) string {
	if ns := obj.GetNamespace(); ns != "" {
		return ns
	}

	return c.defaultResourceNamespace
}

func (c *DynamicClient) Apply(ctx context.Context, obj *unstructured.Unstructured) error {
	gvr, err := c.LookupGVR(obj.GetAPIVersion(), obj.GetKind())
	if err != nil {
		return fmt.Errorf("looking up api endpoints: %w", err)
	}

	rc := c.Resource(gvr).Namespace(c.NamespaceFor(obj))

	_, err = rc.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := rc.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating resource: %w", err)
		}

		return nil
	} else if err != nil {
		return fmt.Errorf("fetching existing resource: %w", err)
	}

	if _, err := rc.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating resource: %w", err)
	}

	return nil
}

func (c *DynamicClient) ApplyObject(ctx context.Context, obj runtime.Object) error {
	enc, _ := json.Marshal(obj) //nolint:errcheck // cannot fail as it's from the serialized kubernetes resource
	out := &unstructured.Unstructured{}
	_ = json.Unmarshal(enc, out) //nolint:errcheck // unmarshaling previously marshaled object with required fields

	return c.Apply(ctx, out)
}

func (c *DynamicClient) delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) error {
	err := c.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		if apierrors.IsForbidden(err) {
			logf.FromContext(ctx).Info("treating forbidden as delete success in case of invalid namespaces")
		} else {
			return fmt.Errorf("failed to delete resource: %w", err)
		}
	}

	return nil
}

func (c *DynamicClient) Delete(ctx context.Context, apiVersion, kind, name, namespace string) error {
	gvr, err := c.LookupGVR(apiVersion, kind)
	if err != nil {
		return fmt.Errorf("looking up api endpoints: %w", err)
	}

	return c.delete(ctx, gvr, name, namespace)
}

func (c *DynamicClient) DeleteObj(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Delete(ctx, obj.GetAPIVersion(), obj.GetKind(), obj.GetName(), obj.GetNamespace())
}

func (c *DynamicClient) DeleteInDefaultNamespace(ctx context.Context, gvr schema.GroupVersionResource, name string) error {
	return c.delete(ctx, gvr, name, c.defaultResourceNamespace)
}
