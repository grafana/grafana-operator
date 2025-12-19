package tk8s

import corev1 "k8s.io/api/core/v1"

func GetConfigMapKeySelector(t tHelper, configMapName, key string) *corev1.ConfigMapKeySelector {
	t.Helper()

	v := &corev1.ConfigMapKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: configMapName,
		},
		Key: key,
	}

	return v
}
