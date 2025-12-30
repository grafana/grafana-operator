package content

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetDashboardEnvs(t *testing.T) {
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	cl := fake.NewClientBuilder().
		WithScheme(s).
		Build()

	dashboard := v1beta1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dashboard",
			Namespace: "grafana-operator-system",
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaContentSpec: v1beta1.GrafanaContentSpec{
				Envs: []v1beta1.GrafanaContentEnv{
					{
						Name:  "TEST_ENV",
						Value: "test-env-value",
					},
				},
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	var contentResource v1beta1.GrafanaContentResource = &dashboard
	assert.NotNil(t, contentResource.GrafanaContentSpec(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")
	assert.NotNil(t, contentResource.GrafanaContentStatus(), "resource does not properly implement content spec or status fields; this indicates a bug in implementation")

	resolver := NewResolver(&dashboard, cl)

	envs, err := resolver.getContentEnvs(ctx)

	require.NoError(t, err)
	assert.NotNil(t, envs)
	assert.Len(t, envs, 1)
}
