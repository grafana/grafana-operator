package fetchers

import (
	"context"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Fetching dashboards from URL", func() {
	t := GinkgoT()

	want := []byte(`{"dummyField": "dummyData"}`)
	wantCompressed, err := cache.Gzip(want)

	require.NoError(t, err)

	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	When("using no authentication", func() {
		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, want),
			))
		})

		It("fetches the correct url", func() {
			dashboard := &v1beta1.GrafanaDashboard{
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						URL: server.URL(),
					},
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			}

			got, err := FetchFromURL(context.Background(), dashboard, k8sClient, nil)
			require.NoError(t, err)

			assert.Equal(t, want, got)
			assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
			assert.Equal(t, server.URL(), dashboard.Status.ContentURL)
			assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
		})
	})

	When("using authentication", func() {
		basicAuthUsername := "admin"
		basicAuthPassword := "admin"

		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth(basicAuthUsername, basicAuthPassword),
				ghttp.RespondWith(http.StatusOK, want),
			))
		})

		It("fetches the correct url", func() {
			dashboard := &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						URL: server.URL(),
						URLAuthorization: &v1beta1.GrafanaContentURLAuthorization{
							BasicAuth: &v1beta1.GrafanaContentURLBasicAuth{
								Username: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "credentials",
									},
									Key:      "USERNAME",
									Optional: nil,
								},
								Password: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "credentials",
									},
									Key:      "PASSWORD",
									Optional: nil,
								},
							},
						},
					},
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			}

			credentialsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "credentials",
					Namespace: "default",
				},
				StringData: map[string]string{
					"USERNAME": "admin",
					"PASSWORD": "admin",
				},
			}

			err = k8sClient.Create(context.Background(), credentialsSecret)
			require.NoError(t, err)

			got, err := FetchFromURL(context.Background(), dashboard, k8sClient, nil)
			require.NoError(t, err)

			assert.Equal(t, want, got)
			assert.Equal(t, wantCompressed, dashboard.Status.ContentCache)
			assert.Equal(t, server.URL(), dashboard.Status.ContentURL)
			assert.NotZero(t, dashboard.Status.ContentTimestamp.Time)
		})
	})
})
