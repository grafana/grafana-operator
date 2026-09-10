package tk8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func getFakeClientBuilder(t *testing.T, initObjs ...client.Object) *fake.ClientBuilder {
	t.Helper()

	s := runtime.NewScheme()

	err := corev1.AddToScheme(s)
	require.NoError(t, err)

	err = appsv1.AddToScheme(s)
	require.NoError(t, err)

	cb := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(initObjs...)

	return cb
}

// GetFakeClient returns a fake k8s client with preconfigured runtime scheme and, optionally, initObjects
func GetFakeClient(t *testing.T, initObjs ...client.Object) client.WithWatch {
	t.Helper()

	cl := getFakeClientBuilder(t, initObjs...).
		Build()

	return cl
}

// GetFakeInterceptingClient returns a fake k8s client with preconfigured runtime scheme, and, optionally, initObjects
// But allows overriding client functions Apply, Update, etc...
func GetFakeInterceptingClient(t *testing.T, intercepterFns *interceptor.Funcs, initObjs ...client.Object) client.WithWatch {
	t.Helper()

	cb := getFakeClientBuilder(t, initObjs...)

	if intercepterFns != nil {
		cb.WithInterceptorFuncs(*intercepterFns)
	}

	cl := cb.Build()

	return cl
}
