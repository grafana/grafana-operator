package fetchers

import (
	"context"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetching dashboards from URL", func() {
	dashboardJSON := []byte(`{"dummyField": "dummyData"}`)
	compressedJSON, err := v1beta1.Gzip(dashboardJSON)
	Expect(err).NotTo(HaveOccurred())

	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	When("using no authentication", func() {
		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, dashboardJSON),
			))
		})

		It("fetches the correct url", func() {
			dashboard := &v1beta1.GrafanaDashboard{
				Spec: v1beta1.GrafanaDashboardSpec{
					Url: server.URL(),
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			}

			fetchedDashboard, err := FetchDashboardFromUrl(context.Background(), dashboard, k8sClient, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedDashboard).To(Equal(fetchedDashboard))
			Expect(dashboard.Status.ContentTimestamp.Time.IsZero()).To(BeFalse())
			Expect(dashboard.Status.ContentCache).To(Equal(compressedJSON))
			Expect(dashboard.Status.ContentUrl).To(Equal(server.URL()))
		})
	})
	When("using authentication", func() {
		basicAuthUsername := "admin"
		basicAuthPassword := "admin"
		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth(basicAuthUsername, basicAuthPassword),
				ghttp.RespondWith(http.StatusOK, dashboardJSON),
			))
		})

		It("fetches the correct url", func() {
			dashboard := &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					Url: server.URL(),
					UrlAuthorization: &v1beta1.GrafanaDashboardUrlAuthorization{
						BasicAuth: &v1beta1.GrafanaDashboardUrlBasicAuth{
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
			Expect(err).NotTo(HaveOccurred())
			fetchedDashboard, err := FetchDashboardFromUrl(context.Background(), dashboard, k8sClient, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedDashboard).To(Equal(fetchedDashboard))
			Expect(dashboard.Status.ContentTimestamp.Time.IsZero()).To(BeFalse())
			Expect(dashboard.Status.ContentCache).To(Equal(compressedJSON))
			Expect(dashboard.Status.ContentUrl).To(Equal(server.URL()))
		})
	})
})
