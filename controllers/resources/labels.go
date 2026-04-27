package resources

import (
	"maps"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ignoredLabelsPrefix = "applyset.kubernetes.io/"

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

	for k, v := range inheritedLabels {
		// https://github.com/grafana/grafana-operator/issues/2642
		if strings.HasPrefix(k, ignoredLabelsPrefix) {
			continue
		}

		labels[k] = v
	}

	maps.Copy(labels, GetCommonLabels())

	meta.SetLabels(labels)
}
