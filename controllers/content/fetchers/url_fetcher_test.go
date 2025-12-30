package fetchers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
)

const (
	basicAuthUsername = "root"
	basicAuthPassword = "secret"
)

func getCredentials(t *testing.T, secretName string) (*corev1.Secret, *v1beta1.GrafanaContentURLAuthorization) {
	t.Helper()

	const (
		usernameKey = "USERNAME"
		passwordKey = "PASSWORD"
	)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			usernameKey: []byte(basicAuthUsername),
			passwordKey: []byte(basicAuthPassword),
		},
	}

	urlAuthorization := &v1beta1.GrafanaContentURLAuthorization{
		BasicAuth: &v1beta1.GrafanaContentURLBasicAuth{
			Username: tk8s.GetSecretKeySelector(t, secretName, usernameKey),
			Password: tk8s.GetSecretKeySelector(t, secretName, passwordKey),
		},
	}

	return secret, urlAuthorization
}

func TestFetchFromURL(t *testing.T) {
	want := []byte(`{"dummyField": "dummyData"}`)
	wantCompressed, err := cache.Gzip(want)

	require.NoError(t, err)

	publicEndpoint := "/public"
	privateEndpoint := "/private"

	mux := http.NewServeMux()
	mux.HandleFunc(publicEndpoint, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(want))
	})
	mux.HandleFunc(privateEndpoint, func(w http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok || username != basicAuthUsername || password != basicAuthPassword {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(want))
	})

	ts := httptest.NewServer(mux)

	t.Cleanup(func() {
		ts.Close()
	})

	s := runtime.NewScheme()
	err = corev1.AddToScheme(s)
	require.NoError(t, err)

	cl := fake.NewClientBuilder().
		WithScheme(s).
		Build()

	t.Run("no authentication", func(t *testing.T) {
		url := ts.URL + publicEndpoint

		dashboard := &v1beta1.GrafanaDashboard{
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					URL: url,
				},
			},
			Status: v1beta1.GrafanaDashboardStatus{},
		}

		got, err := FetchFromURL(context.Background(), dashboard, cl, nil)
		require.NoError(t, err)

		assert.Equal(t, want, got)
		assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
		assert.Equal(t, url, dashboard.Status.ContentURL)
		assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
	})

	t.Run("using authentication", func(t *testing.T) {
		url := ts.URL + privateEndpoint

		credentialsSecret, urlAuthorization := getCredentials(t, "credentials")

		dashboard := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "url-basic-auth",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					URL:              url,
					URLAuthorization: urlAuthorization,
				},
			},
			Status: v1beta1.GrafanaDashboardStatus{},
		}

		err = cl.Create(context.Background(), credentialsSecret)
		require.NoError(t, err)

		got, err := FetchFromURL(context.Background(), dashboard, cl, nil)
		require.NoError(t, err)

		assert.Equal(t, want, got)
		assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
		assert.Equal(t, url, dashboard.Status.ContentURL)
		assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
	})
}
