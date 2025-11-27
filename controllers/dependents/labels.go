package dependents

import (
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "grafana-operator",
	}
}

func SetInheritedLabels(obj metav1.ObjectMetaAccessor, extraLabels map[string]string) {
	meta := obj.GetObjectMeta()

	labels := meta.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	// Inherit labels from the parent grafana instance if any
	maps.Copy(labels, extraLabels)
	// Ensure default CommonLabels for child resources
	maps.Copy(labels, GetCommonLabels())
	meta.SetLabels(labels)
}
