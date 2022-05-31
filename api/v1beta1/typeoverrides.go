package v1beta1

import (
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"reflect"
)

// +kubebuilder:object:generate=true

// ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta).
type ObjectMeta struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// Merge merges it's receivers values into the incoming ObjectMeta by overwriting values for existing keys and adding new ones.
func (override *ObjectMeta) Merge(meta metav1.ObjectMeta) metav1.ObjectMeta {
	if override == nil {
		return meta
	}
	if len(override.Annotations) > 0 {
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
		for key, val := range override.Annotations {
			meta.Annotations[key] = val
		}
	}
	if len(override.Labels) > 0 {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		for key, val := range override.Labels {
			meta.Labels[key] = val
		}
	}
	return meta
}

// +kubebuilder:object:generate=true

// IngressNetworkingV1 is a subset of [Ingress in k8s.io/api/networking/v1beta1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingress-v1-networking-k8s-io).
type IngressNetworkingV1 struct {
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`
	// Kubernetes [Ingress Specification](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#ingressclassspec-v1-networking-k8s-io)
	Spec v1.IngressSpec `json:"spec,omitempty"`
}

// Merge merges `overrides` into `base` using the SMP (structural merge patch) approach.
// - It intentionally does not remove fields present in base but missing from overrides
// - It merges slices only if the `patchStrategy:"merge"` tag is present and the `patchMergeKey` identifies the unique field
func Merge(base, overrides interface{}) error {
	if overrides == nil {
		return nil
	}

	baseBytes, err := json.Marshal(base)
	if err != nil {
		return errors.Wrap(err, "failed to convert current object to byte sequence")
	}

	overrideBytes, err := json.Marshal(overrides)
	if err != nil {
		return errors.Wrap(err, "failed to convert current object to byte sequence")
	}

	patchMeta, err := strategicpatch.NewPatchMetaFromStruct(base)
	if err != nil {
		return errors.Wrap(err, "failed to produce patch meta from struct")
	}
	patch, err := strategicpatch.CreateThreeWayMergePatch(overrideBytes, overrideBytes, baseBytes, patchMeta, true)
	if err != nil {
		return errors.Wrap(err, "failed to create three way merge patch")
	}

	merged, err := strategicpatch.StrategicMergePatchUsingLookupPatchMeta(baseBytes, patch, patchMeta)
	if err != nil {
		return errors.Wrap(err, "failed to apply patch")
	}

	valueOfBase := reflect.Indirect(reflect.ValueOf(base))
	into := reflect.New(valueOfBase.Type())
	if err := json.Unmarshal(merged, into.Interface()); err != nil {
		return err
	}
	if !valueOfBase.CanSet() {
		return errors.New("unable to set unmarshalled value into base object")
	}
	valueOfBase.Set(reflect.Indirect(into))
	return nil
}
