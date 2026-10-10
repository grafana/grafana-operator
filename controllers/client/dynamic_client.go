package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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

	existing, err := rc.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := rc.Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating resource: %w", err)
		}

		return nil
	} else if err != nil {
		return fmt.Errorf("fetching existing resource: %w", err)
	}

	// An update that changes nothing is not free: Grafana handles it like any other and
	// announces a dashboard as saved to every browser that has it open, which reloads it there.
	// Without this check that happens at every resync.
	if matchesExisting(obj, existing) {
		return nil
	}

	// Carry over the resource version to ensure the object will be accepted if other updates have been made
	obj.SetResourceVersion(existing.GetResourceVersion())

	if _, err := rc.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating resource: %w", err)
	}

	return nil
}

// matchesExisting reports whether existing already holds everything an update with obj would
// set. Grafana adds metadata of its own (labels and annotations under grafana.app/), so a label
// or an annotation of obj only has to be present in existing, and status is the server's. Every
// other top-level field has to be equal both ways, as an update replaces it.
func matchesExisting(obj, existing *unstructured.Unstructured) bool {
	keys := make(map[string]struct{}, len(obj.Object)+len(existing.Object))
	for key := range obj.Object {
		keys[key] = struct{}{}
	}

	for key := range existing.Object {
		keys[key] = struct{}{}
	}

	for key := range keys {
		switch key {
		case "apiVersion", "kind", "metadata", "status":
			continue
		}

		if !equalIgnoringNull(obj.Object[key], existing.Object[key]) {
			return false
		}
	}

	return isSubset(obj.GetLabels(), existing.GetLabels()) &&
		isSubset(obj.GetAnnotations(), existing.GetAnnotations())
}

func isSubset(subset, set map[string]string) bool {
	for key, value := range subset {
		if actual, ok := set[key]; !ok || actual != value {
			return false
		}
	}

	return true
}

// equalIgnoringNull compares two values by their JSON, so that a number decoded as int64 equals
// the same number produced by a patch as an int, while every digit of a large number still
// counts. A null member of an object counts as absent: Grafana stores some nullable fields
// explicitly, such as the value of a dashboard's base threshold step, that a template leaves out.
func equalIgnoringNull(a, b any) bool {
	na, errA := withoutNull(a)
	nb, errB := withoutNull(b)

	if errA != nil || errB != nil {
		return false
	}

	return reflect.DeepEqual(na, nb)
}

func withoutNull(v any) (any, error) {
	enc, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(enc))
	dec.UseNumber()

	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}

	return dropNull(out), nil
}

func dropNull(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		for key, value := range typed {
			if value == nil {
				delete(typed, key)

				continue
			}

			typed[key] = dropNull(value)
		}
	case []any:
		for i, value := range typed {
			typed[i] = dropNull(value)
		}
	}

	return v
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

	// This should be safe as there are no cluster scoped resources (yet). If we
	// ever need to support them, check if the resoure is namespaced or cluster
	// scoped first
	if namespace == "" {
		namespace = c.defaultResourceNamespace
	}

	return c.delete(ctx, gvr, name, namespace)
}

func (c *DynamicClient) DeleteObj(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Delete(ctx, obj.GetAPIVersion(), obj.GetKind(), obj.GetName(), obj.GetNamespace())
}

func (c *DynamicClient) DeleteInDefaultNamespace(ctx context.Context, gvr schema.GroupVersionResource, name string) error {
	return c.delete(ctx, gvr, name, c.defaultResourceNamespace)
}
