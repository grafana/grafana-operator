package grafana

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestGetGrafanaImage(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "not specified(default version)",
			version: "",
			want:    fmt.Sprintf("%s:%s", config.GrafanaImage, config.GrafanaVersion),
		},
		{
			name:    "custom tag",
			version: "10.4.0",
			want:    fmt.Sprintf("%s:10.4.0", config.GrafanaImage),
		},
		{
			name:    "fully-qualified image",
			version: "docker.io/grafana/grafana@sha256:b7fcb534f7b3512801bb3f4e658238846435804deb479d105b5cdc680847c272",
			want:    "docker.io/grafana/grafana@sha256:b7fcb534f7b3512801bb3f4e658238846435804deb479d105b5cdc680847c272",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Version: tt.version,
				},
			}

			got := getGrafanaImage(cr)

			assert.Equal(t, tt.want, got)
		})
	}
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

var _ = Describe("Deployment reconciler secrets hash", func() {
	t := GinkgoT()

	It("sets SecretsHash and checksum/secrets annotation when referenced secrets exist", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-test-secret",
			},
			StringData: map[string]string{"password": "s3cr3t"},
		}

		ctx := context.Background()

		err := cl.Create(ctx, secret)
		require.NoError(t, err)

		containers := []corev1.Container{
			{
				Name: "grafana",
				Env: []corev1.EnvVar{
					{
						Name:      "PASSWORD",
						ValueFrom: tk8s.GetEnvVarSecretSource(t, secret.Name, "password"),
					},
				},
			},
		}

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-test-grafana",
			},
		}
		cr.Spec.SetContainers(containers)

		err = cl.Create(ctx, cr)
		require.NoError(t, err)

		r := NewDeploymentReconciler(cl, false)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.NotEmpty(t, vars.SecretsHash)
	})

	It("sets empty SecretsHash when no secrets are referenced", func() {
		ctx := context.Background()

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-no-refs",
			},
			Spec: v1beta1.GrafanaSpec{},
		}

		err := cl.Create(ctx, cr)
		require.NoError(t, err)

		r := NewDeploymentReconciler(cl, false)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.Empty(t, vars.SecretsHash)
	})

	It("produces a different hash after a referenced secret is updated", func() {
		ctx := context.Background()

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-rotation-secret",
			},
			StringData: map[string]string{"password": "initial"},
		}

		err := cl.Create(ctx, secret)
		require.NoError(t, err)

		containers := []corev1.Container{
			{
				Name: "grafana",
				EnvFrom: []corev1.EnvFromSource{
					{
						SecretRef: tk8s.GetEnvFromSecretSource(t, secret.Name),
					},
				},
			},
		}

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-rotation-grafana",
			},
		}

		cr.Spec.SetContainers(containers)

		err = cl.Create(ctx, cr)
		require.NoError(t, err)

		r := NewDeploymentReconciler(cl, false)

		vars1 := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(context.Background(), cr, vars1, scheme.Scheme)
		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.NotEmpty(t, vars1.SecretsHash)

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
		ctx := context.Background()

		containers := []corev1.Container{
			{
				Name: "grafana",
				Env: []corev1.EnvVar{
					{
						Name:      "MISSING_KEY",
						ValueFrom: tk8s.GetEnvVarSecretSource(t, "does-not-exist", "key"),
					},
				},
			},
		}

		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "deploy-secrets-hash-missing-secret",
			},
		}

		cr.Spec.SetContainers(containers)

		err := cl.Create(ctx, cr)
		require.NoError(t, err)

		r := NewDeploymentReconciler(cl, false)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
	})
})
