package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

func TestRemoveMissingCRs(t *testing.T) {
	statusList := v1beta1.NamespacedResourceList{
		"default/present/uid",
		"default/missing/uid",
		"other/missing/uid",
	}

	dashboards := v1beta1.GrafanaDashboardList{
		Items: []v1beta1.GrafanaDashboard{
			{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "present"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "unrelated-dashboard"},
			},
		},
	}

	// Sanity checks before test
	assert.Len(t, statusList, 3)
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/present/uid"))
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/missing/uid"))
	assert.Contains(t, statusList, v1beta1.NamespacedResource("other/missing/uid"))

	updateStatus := false
	removeMissingCRs(&statusList, &dashboards, &updateStatus)

	assert.True(t, updateStatus, "Entries were removed but status change was not detected")

	assert.Len(t, statusList, 1)
	assert.Contains(t, statusList, v1beta1.NamespacedResource("default/present/uid"))
	assert.NotContains(t, statusList, v1beta1.NamespacedResource("default/missing/uid"))
	assert.NotContains(t, statusList, v1beta1.NamespacedResource("other/missing/uid"))

	found, _ := statusList.Find("default", "unrelated-dashboard")
	assert.False(t, found, "Dashboard is not in status and should not be")
}

func TestGrafanaIndexing(t *testing.T) {
	reconciler := &GrafanaReconciler{
		Client: cl,
	}

	t.Run("indexSecretSource returns secrets from container env secretKeyRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
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
													SecretKeyRef: tk8s.GetSecretKeySelector(t, "db-secret", "password"),
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

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/db-secret"}
		require.Equal(t, expected, result)
	})

	t.Run("indexSecretSource returns secrets from container envFrom secretRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
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
												LocalObjectReference: corev1.LocalObjectReference{Name: "bulk-secret"},
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

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/bulk-secret"}
		require.Equal(t, expected, result)
	})

	t.Run("indexSecretSource returns secrets from external spec", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{
					URL:           "https://grafana.example.com",
					AdminUser:     tk8s.GetSecretKeySelector(t, "ext-creds", "user"),
					AdminPassword: tk8s.GetSecretKeySelector(t, "ext-creds", "password"),
				},
			},
		}

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/ext-creds"}
		require.Equal(t, expected, result)
	})

	t.Run("indexSecretSource returns empty slice when no secret references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{ConfigMapRef: &corev1.ConfigMapEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{Name: "only-a-cm"},
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

		indexFunc := reconciler.indexSecretSource()
		result := indexFunc(cr)

		require.Empty(t, result)
	})

	t.Run("indexConfigMapSource returns configmaps from container env configMapKeyRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
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
													ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "app-config", "log_level"),
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

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/app-config"}
		require.Equal(t, expected, result)
	})

	t.Run("indexConfigMapSource returns configmaps from container envFrom configMapRef", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{ConfigMapRef: &corev1.ConfigMapEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{Name: "bulk-cm"},
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

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/bulk-cm"}
		require.Equal(t, expected, result)
	})

	t.Run("indexConfigMapSource returns configmaps from volume configMap reference", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Volumes: []corev1.Volume{
									{
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{Name: "vol-cm"},
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

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cr)

		expected := []string{"test-namespace/vol-cm"}
		require.Equal(t, expected, result)
	})

	t.Run("indexConfigMapSource returns empty slice when no configmap references", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
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
												LocalObjectReference: corev1.LocalObjectReference{Name: "only-a-secret"},
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

		indexFunc := reconciler.indexConfigMapSource()
		result := indexFunc(cr)

		require.Empty(t, result)
	})

	t.Run("both index functions handle multiple references across containers and initContainers", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				Deployment: &v1beta1.DeploymentV1{
					Spec: v1beta1.DeploymentV1Spec{
						Template: &v1beta1.DeploymentV1PodTemplateSpec{
							Spec: &v1beta1.DeploymentV1PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: tk8s.GetSecretKeySelector(t, "secret1", "key"),
											}},
											{ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: tk8s.GetConfigMapKeySelector(t, "cm1", "key"),
											}},
										},
									},
								},
								InitContainers: []corev1.Container{
									{
										EnvFrom: []corev1.EnvFromSource{
											{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "secret2"}}},
											{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "cm2"}}},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		secretIndexFunc := reconciler.indexSecretSource()
		secretResult := secretIndexFunc(cr)
		assert.Equal(t, []string{"test-namespace/secret1", "test-namespace/secret2"}, secretResult)

		cmIndexFunc := reconciler.indexConfigMapSource()
		cmResult := cmIndexFunc(cr)
		assert.Equal(t, []string{"test-namespace/cm1", "test-namespace/cm2"}, cmResult)
	})
}

var _ = Describe("Grafana Reconciler: Provoke Conditions", func() {
	t := GinkgoT()

	tests := []struct {
		name string
		meta metav1.ObjectMeta
		spec v1beta1.GrafanaSpec
		want metav1.Condition
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaSpec{
				Suspend: true,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonReconcileSuspended,
			},
		},
		// TODO When InvalidSpec is implemented for external instances admin secret referencing a non-existing secret
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.Grafana{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			err := cl.Create(testCtx, cr)
			require.NoError(t, err)

			r := GrafanaReconciler{Client: cl, Scheme: cl.Scheme()}
			req := tk8s.GetRequest(t, cr)

			_, err = r.Reconcile(testCtx, req)
			require.NoError(t, err)

			cr = &v1beta1.Grafana{}

			err = r.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			hasCondition := tk8s.HasCondition(t, cr, tt.want)
			assert.True(t, hasCondition)
		})
	}
})
