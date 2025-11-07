package fetchers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"

	. "github.com/onsi/ginkgo/v2"
)

const (
	basicAuthUsername = "root"
	basicAuthPassword = "secret"
)

func getCredentials(secretName string) (*corev1.Secret, *v1beta1.GrafanaContentURLAuthorization) {
	GinkgoHelper()

	const (
		usernameKey = "USERNAME"
		passwordKey = "PASSWORD"
	)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		StringData: map[string]string{
			usernameKey: basicAuthUsername,
			passwordKey: basicAuthPassword,
		},
	}

	urlAuthorization := &v1beta1.GrafanaContentURLAuthorization{
		BasicAuth: &v1beta1.GrafanaContentURLBasicAuth{
			Username: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key:      usernameKey,
				Optional: nil,
			},
			Password: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key:      passwordKey,
				Optional: nil,
			},
		},
	}

	return secret, urlAuthorization
}

var _ = Describe("URL fetcher", Ordered, func() {
	t := GinkgoT()

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
	AfterAll(func() {
		ts.Close()
	})

	When("no authentication", func() {
		url := ts.URL + publicEndpoint

		It("successfully fetches the dashboard", func() {
			dashboard := &v1beta1.GrafanaDashboard{
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						URL: url,
					},
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			}

			got, err := FetchFromURL(context.Background(), dashboard, k8sClient, nil)
			require.NoError(t, err)

			assert.Equal(t, want, got)
			assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
			assert.Equal(t, url, dashboard.Status.ContentURL)
			assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
		})
	})

	When("using authentication", func() {
		url := ts.URL + privateEndpoint

		It("successfully fetches the dashboard", func() {
			credentialsSecret, urlAuthorization := getCredentials("credentials")

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

			err = k8sClient.Create(context.Background(), credentialsSecret)
			require.NoError(t, err)

			got, err := FetchFromURL(context.Background(), dashboard, k8sClient, nil)
			require.NoError(t, err)

			assert.Equal(t, want, got)
			assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
			assert.Equal(t, url, dashboard.Status.ContentURL)
			assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
		})
	})
})
