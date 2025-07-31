package model

import (
	"crypto/rand"
	"encoding/base64"
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}

func RandStringRunes(s int) string {
	b := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)
}

func MergeAnnotations(requested map[string]string, existing map[string]string) map[string]string {
	if existing == nil {
		return requested
	}

	maps.Copy(existing, requested)

	return existing
}

func BoolPtr(b bool) *bool { return &b }

func IntPtr(b int64) *int64 { return &b }

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
