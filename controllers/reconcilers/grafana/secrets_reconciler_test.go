package grafana

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestReferencedSecretsAndConfigMaps(t *testing.T) {
	t.Run("container env secretKeyRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret"},
														Key:                  "password",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"db-secret"}, secrets)
		assert.Empty(t, configMaps)
	})

	t.Run("container envFrom secretRef and configMapRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "bulk-secret"}}},
											{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "bulk-cm"}}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"bulk-secret"}, secrets)
		assert.Equal(t, []string{"bulk-cm"}, configMaps)
	})

	t.Run("initContainer env references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								InitContainers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: "init-secret"},
														Key:                  "token",
													},
												},
											},
										},
										EnvFrom: []corev1.EnvFromSource{
											{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "init-cm"}}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"init-secret"}, secrets)
		assert.Equal(t, []string{"init-cm"}, configMaps)
	})

	t.Run("volume secret and configmap references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Volumes: []corev1.Volume{
									{
										VolumeSource: corev1.VolumeSource{
											Secret: &corev1.SecretVolumeSource{SecretName: "tls-secret"},
										},
									},
									{
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{Name: "config-vol"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"tls-secret"}, secrets)
		assert.Equal(t, []string{"config-vol"}, configMaps)
	})

	t.Run("external grafana secret references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{
					URL: "https://grafana.example.com",
					AdminUser: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "ext-creds"},
						Key:                  "user",
					},
					AdminPassword: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "ext-creds"},
						Key:                  "password",
					},
					APIKey: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "ext-apikey"},
						Key:                  "key",
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"ext-apikey", "ext-creds"}, secrets)
		assert.Empty(t, configMaps)
	})

	t.Run("client tls secret reference", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Client: &v1beta1.GrafanaClient{
					TLS: &v1beta1.TLSConfig{
						CertSecretRef: &corev1.SecretReference{Name: "client-tls"},
					},
				},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"client-tls"}, secrets)
		assert.Empty(t, configMaps)
	})

	t.Run("deduplicates repeated secret references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: "shared-secret"},
														Key:                  "user",
													},
												},
											},
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: "shared-secret"},
														Key:                  "password",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, _ := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"shared-secret"}, secrets)
	})

	t.Run("returns sorted results", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "zebra-secret"}}},
											{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "alpha-secret"}}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secrets, _ := ReferencedSecretsAndConfigMaps(cr)

		assert.Equal(t, []string{"alpha-secret", "zebra-secret"}, secrets)
	})

	t.Run("returns empty slices when no references exist", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec:       v1beta1.GrafanaSpec{},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Empty(t, secrets)
		assert.Empty(t, configMaps)
	})

	t.Run("ignores nil deployment spec fields", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{},
			},
		}

		secrets, configMaps := ReferencedSecretsAndConfigMaps(cr)

		assert.Empty(t, secrets)
		assert.Empty(t, configMaps)
	})
}

func TestHashResourceVersions(t *testing.T) {
	t.Run("empty list returns empty string", func(t *testing.T) {
		result := hashResourceVersions(nil)
		assert.Empty(t, result)

		result = hashResourceVersions([]string{})
		assert.Empty(t, result)
	})

	t.Run("same inputs produce same hash", func(t *testing.T) {
		versions := []string{"secret/db=100", "configmap/cfg=200"}

		hash1 := hashResourceVersions(versions)
		hash2 := hashResourceVersions(versions)

		assert.Equal(t, hash1, hash2)
		assert.NotEmpty(t, hash1)
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		hash1 := hashResourceVersions([]string{"secret/db=100"})
		hash2 := hashResourceVersions([]string{"secret/db=101"})

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("order-independent: same entries in different order produce same hash", func(t *testing.T) {
		hash1 := hashResourceVersions([]string{"secret/a=1", "configmap/b=2"})
		hash2 := hashResourceVersions([]string{"configmap/b=2", "secret/a=1"})

		assert.Equal(t, hash1, hash2)
	})

	t.Run("returns a valid hex string", func(t *testing.T) {
		result := hashResourceVersions([]string{"secret/x=42"})
		assert.Regexp(t, `^[0-9a-f]{64}$`, result)
	})
}

var _ = Describe("Reconcile Secrets", func() {
	t := GinkgoT()

	It("sets SecretsHash when referenced secrets exist", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-test-secret",
			},
			StringData: map[string]string{"password": "s3cr3t"},
		}

		err := cl.Create(context.Background(), secret)
		require.NoError(t, err)

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
														Key:                  "password",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		r := NewSecretsReconciler(cl)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.NotEmpty(t, vars.SecretsHash)
	})

	It("sets empty SecretsHash when no secrets are referenced", func() {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-no-refs",
			},
			Spec: v1beta1.GrafanaSpec{},
		}

		r := NewSecretsReconciler(cl)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.Empty(t, vars.SecretsHash)
	})

	It("produces a different hash after a referenced secret is updated", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-rotation-secret",
			},
			StringData: map[string]string{"password": "initial"},
		}

		err := cl.Create(context.Background(), secret)
		require.NoError(t, err)

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-rotation-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{SecretRef: &corev1.SecretEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
											}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		r := NewSecretsReconciler(cl)

		vars1 := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(context.Background(), cr, vars1, scheme.Scheme)
		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.NotEmpty(t, vars1.SecretsHash)

		// Simulate a secret rotation by updating the secret data
		secret.StringData = map[string]string{"password": "rotated"}
		err = cl.Update(context.Background(), secret)
		require.NoError(t, err)

		vars2 := &v1beta1.OperatorReconcileVars{}
		status, err = r.Reconcile(context.Background(), cr, vars2, scheme.Scheme)
		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.NotEmpty(t, vars2.SecretsHash)

		assert.NotEqual(t, vars1.SecretsHash, vars2.SecretsHash, "hash should change after secret rotation")
	})

	It("skips missing secrets without failing", func() {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secrets-reconciler-missing-secret",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{Name: "does-not-exist"},
														Key:                  "key",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		r := NewSecretsReconciler(cl)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
	})
})
