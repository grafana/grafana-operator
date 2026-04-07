package tk8s

import (
	corev1 "k8s.io/api/core/v1"
)

func GetEnvFromSecretSource(t tHelper, secretName string) *corev1.SecretEnvSource {
	t.Helper()

	v := &corev1.SecretEnvSource{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: secretName,
		},
	}

	return v
}

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

func GetVolumeSecretSource(t tHelper, secretName string) corev1.VolumeSource {
	t.Helper()

	v := corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName: secretName,
		},
	}

	return v
}
