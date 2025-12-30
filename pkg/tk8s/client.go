package tk8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func GetFakeClient(t *testing.T, initObjs ...client.Object) client.WithWatch {
	t.Helper()

	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err)

	err = appsv1.AddToScheme(s)
	require.NoError(t, err)

	cl := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(initObjs...).
		Build()

	return cl
}
