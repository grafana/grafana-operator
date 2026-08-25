package tk8s

import corev1 "k8s.io/api/core/v1"

func GetEnvFromConfigMapSource(t tHelper, configMapName string) *corev1.ConfigMapEnvSource {
	t.Helper()

	v := &corev1.ConfigMapEnvSource{
		Name: configMapName,
	}

	return v
}

func GetEnvVarConfigMapSource(t tHelper, configMapName, key string) *corev1.EnvVarSource {
	t.Helper()

	v := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			Name: configMapName,
			Key:  key,
		},
	}

	return v
}

func GetConfigMapKeySelector(t tHelper, configMapName, key string) *corev1.ConfigMapKeySelector {
	t.Helper()

	v := &corev1.ConfigMapKeySelector{
		Name: configMapName,
		Key:  key,
	}

	return v
}

func GetVolumeConfigMapSource(t tHelper, configMapName string) corev1.VolumeSource {
	t.Helper()

	v := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			Name: configMapName,
		},
	}

	return v
}
