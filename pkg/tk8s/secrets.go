package tk8s

import (
	corev1 "k8s.io/api/core/v1"
)

func GetEnvVarSecretSource(t tHelper, secretName, key string) *corev1.EnvVarSource {
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

func GetSecretKeySelector(t tHelper, secretName, key string) *corev1.SecretKeySelector {
	t.Helper()

	v := &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: secretName,
		},
		Key: key,
	}

	return v
}
