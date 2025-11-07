package v1beta1

import (
	"crypto/sha256"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WARN Run `make` on all file changes

type ValueFrom struct {
	TargetPath string          `json:"targetPath"`
	ValueFrom  ValueFromSource `json:"valueFrom"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.configMapKeyRef) && !has(self.secretKeyRef)) || (!has(self.configMapKeyRef) && has(self.secretKeyRef))", message="Either configMapKeyRef or secretKeyRef must be set"
type ValueFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// Common Options that all CRs should embed, excluding GrafanaSpec
// Ensure alignment on handling ResyncPeriod, InstanceSelector, and AllowCrossNamespaceImport
// +kubebuilder:validation:XValidation:rule="!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport)", message="disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state"
type GrafanaCommonSpec struct {
	// How often the resource is synced, defaults to 10m0s if not set
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	ResyncPeriod metav1.Duration `json:"resyncPeriod,omitempty"`

	// Selects Grafana instances for import
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.instanceSelector is immutable"
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector"`

	// Allow the Operator to match this resource with Grafanas outside the current namespace
	// +optional
	// +kubebuilder:default=false
	AllowCrossNamespaceImport bool `json:"allowCrossNamespaceImport,omitempty"`

	// Suspend pauses synchronizing attempts and tells the operator to ignore changes
	// +optional
	Suspend bool `json:"suspend,omitempty"`
}

// Common Functions that all CRs should implement, excluding Grafana
// +kubebuilder:object:generate=false
type CommonResource interface {
	client.Object
	MatchLabels() *metav1.LabelSelector
	MatchNamespace() string
	Metadata() metav1.ObjectMeta
	AllowCrossNamespace() bool
	CommonStatus() *GrafanaCommonStatus
}

// The most recent observed state of a Grafana resource
type GrafanaCommonStatus struct {
	// Results when synchonizing resource with Grafana instances
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Last time the resource was synchronized with Grafana instances
	LastResync metav1.Time `json:"lastResync,omitempty"`
}

func GetPluginConfigMapKey(prefix string, m metav1.Object) string {
	ns := m.GetNamespace() // Subject to 63 character limit
	name := m.GetName()    // Up to 253 characters, needs to be cut
	limit := 63

	if len(name) > limit {
		hash := sha256.New()
		hash.Write([]byte(name))

		name = fmt.Sprintf("%v-%x", name[:limit], hash.Sum(nil))
	}

	key := fmt.Sprintf("%v_%v_%v", prefix, ns, name)

	return key
}
