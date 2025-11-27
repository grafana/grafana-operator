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

func SetInheritedLabels(obj metav1.ObjectMetaAccessor, inheritedLabels map[string]string) {
	meta := obj.GetObjectMeta()

	labels := meta.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	maps.Copy(labels, inheritedLabels)
	maps.Copy(labels, GetCommonLabels())

	meta.SetLabels(labels)
}
