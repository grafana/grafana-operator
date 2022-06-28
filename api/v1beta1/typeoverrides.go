package v1beta1

import (
	"encoding/json"
	v12 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	v13 "k8s.io/api/apps/v1"
	v14 "k8s.io/api/core/v1"
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

type DeploymentV1 struct {
	ObjectMeta ObjectMeta       `json:"metadata,omitempty"`
	Spec       DeploymentV1Spec `json:"spec,omitempty"`
}

type DeploymentV1Spec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	Template *v14.PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`

	// +optional
	// +patchStrategy=retainKeys
	Strategy *v13.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,5,opt,name=minReadySeconds"`

	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// +optional
	Paused bool `json:"paused,omitempty" protobuf:"varint,7,opt,name=paused"`

	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`
}

// +kubebuilder:object:generate=true

type IngressNetworkingV1 struct {
	ObjectMeta ObjectMeta      `json:"metadata,omitempty"`
	Spec       *v1.IngressSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true

type RouteOpenshiftV1 struct {
	ObjectMeta ObjectMeta            `json:"metadata,omitempty"`
	Spec       *RouteOpenShiftV1Spec `json:"spec,omitempty"`
}

type RouteOpenShiftV1Spec struct {
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	To *v12.RouteTargetReference `json:"to,omitempty" protobuf:"bytes,3,opt,name=to"`

	AlternateBackends []v12.RouteTargetReference `json:"alternateBackends,omitempty" protobuf:"bytes,4,rep,name=alternateBackends"`

	Port *v12.RoutePort `json:"port,omitempty" protobuf:"bytes,5,opt,name=port"`

	TLS *v12.TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`

	WildcardPolicy v12.WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,7,opt,name=wildcardPolicy"`
}

type ServiceV1 struct {
	ObjectMeta ObjectMeta       `json:"metadata,omitempty"`
	Spec       *v14.ServiceSpec `json:"spec,omitempty"`
}

type PersistentVolumeClaimV1 struct {
	ObjectMeta ObjectMeta                   `json:"metadata,omitempty"`
	Spec       *PersistentVolumeClaimV1Spec `json:"spec,omitempty"`
}

type PersistentVolumeClaimV1Spec struct {
	// +optional
	AccessModes []v14.PersistentVolumeAccessMode `json:"accessModes,omitempty" protobuf:"bytes,1,rep,name=accessModes,casttype=PersistentVolumeAccessMode"`
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,4,opt,name=selector"`
	// +optional
	Resources *v14.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
	// VolumeName is the binding reference to the PersistentVolume backing this claim.
	// +optional
	VolumeName string `json:"volumeName,omitempty" protobuf:"bytes,3,opt,name=volumeName"`
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty" protobuf:"bytes,5,opt,name=storageClassName"`
	// +optional
	VolumeMode *v14.PersistentVolumeMode `json:"volumeMode,omitempty" protobuf:"bytes,6,opt,name=volumeMode,casttype=PersistentVolumeMode"`
	// +optional
	DataSource *v14.TypedLocalObjectReference `json:"dataSource,omitempty" protobuf:"bytes,7,opt,name=dataSource"`
	// +optional
	DataSourceRef *v14.TypedLocalObjectReference `json:"dataSourceRef,omitempty" protobuf:"bytes,8,opt,name=dataSourceRef"`
}

type ServiceAccountV1 struct {
	ObjectMeta ObjectMeta            `json:"metadata,omitempty"`
	Secrets    []v14.ObjectReference `json:"secrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=secrets"`
	// +optional
	ImagePullSecrets []v14.LocalObjectReference `json:"imagePullSecrets,omitempty" protobuf:"bytes,3,rep,name=imagePullSecrets"`
	// +optional
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty" protobuf:"varint,4,opt,name=automountServiceAccountToken"`
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
