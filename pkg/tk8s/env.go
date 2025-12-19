package tk8s

import (
	corev1 "k8s.io/api/core/v1"
)

type testHelper interface {
	Helper()
}

func GetConfigMapKeySelector(t testHelper, configMapName, key string) *corev1.ConfigMapKeySelector {
	t.Helper()

	v := &corev1.ConfigMapKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: configMapName,
		},
		Key: key,
	}

	return v
}

func GetEnvVarSecretSource(t testHelper, secretName, key string) *corev1.EnvVarSource {
	t.Helper()

	v := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
			Key: key,
		},
	}

	return v
}

func GetSecretKeySelector(t testHelper, secretName, key string) *corev1.SecretKeySelector {
	t.Helper()

	v := &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: secretName,
		},
		Key: key,
	}

	return v
}
