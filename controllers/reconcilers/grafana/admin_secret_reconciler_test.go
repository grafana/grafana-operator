package grafana

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Reconcile AdminSecret", func() {
	It("runs successfully with disabled default admin secret", func() {
		r := NewAdminSecretReconciler(k8sClient)
		cr := &v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				DisableDefaultAdminSecret: true,
			},
		}

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))
	})
})

func TestGetAdminUser(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]map[string]string
		secret *corev1.Secret
		want   []byte
	}{
		{
			name: "config section is preferred",
			config: map[string]map[string]string{
				"security": {
					"admin_user": "user_from_config",
				},
			},
			secret: &corev1.Secret{
				Data: map[string][]byte{
					config.GrafanaAdminUserEnvVar: []byte("user_from_secret"),
				},
			},
			want: []byte("user_from_config"),
		},
		{
			name:   "value from secret when config is not set",
			config: map[string]map[string]string{},
			secret: &corev1.Secret{
				Data: map[string][]byte{
					config.GrafanaAdminUserEnvVar: []byte("user_from_secret"),
				},
			},
			want: []byte("user_from_secret"),
		},
		{
			name:   "default user when config is not set and secret is empty",
			config: map[string]map[string]string{},
			secret: &corev1.Secret{
				Data: map[string][]byte{},
			},
			want: []byte(config.DefaultAdminUser),
		},
		{
			name:   "default user when config is not set and secret data is nil",
			config: map[string]map[string]string{},
			secret: &corev1.Secret{
				Data: nil,
			},
			want: []byte(config.DefaultAdminUser),
		},
		{
			name:   "default user when config is not set and secret is nil",
			config: map[string]map[string]string{},
			secret: nil,
			want:   []byte(config.DefaultAdminUser),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Config: tt.config,
				},
			}

			got := getAdminUser(cr, tt.secret)

			assert.Equal(t, tt.want, got)
		})
	}
}
