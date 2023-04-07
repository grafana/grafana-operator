/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	gapi "github.com/grafana/grafana-api-golang-client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
)

var _ = Describe("GrafanaDashboard controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	var (
		grafanaName          = "dashboard-test-grafana"
		grafanaNamespace     = "default"
		adminCredentialsName = fmt.Sprintf("%s-admin-credentials", grafanaName)
		dashboardJsonString  = `{"totally": "a dashboard"}`

		folderTitle   = "test-folder"
		dashboardName = "test-dashboard"
		dashboardUID  = "test-dashboard-uid"

		expectedFolderOutput = gapi.Folder{ID: 13, UID: "folder-uid", Title: folderTitle}

		oneSecondDuration     = time.Second
		oneSecondMetaDuration = metav1.Duration{Duration: oneSecondDuration}
		interval              = time.Millisecond * 250
	)

	var (
		piechartPlugin = v1beta1.GrafanaPlugin{
			Name:    "grafana-piechart-panel",
			Version: "1.6.1",
		}

		mockDashboard = map[string]interface{}{
			"not": "really",
			"a":   "dashboard",
		}

		mockFolders          []gapi.Folder
		mockDashboardStorage map[string]gapi.Dashboard

		server *ghttp.Server

		grafana = &v1beta1.Grafana{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "grafana.integreatly.org/v1beta1",
				Kind:       "Grafana",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      grafanaName,
				Namespace: grafanaNamespace,
				Labels:    map[string]string{"test": grafanaName},
			},
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{
					URL: "replaced BeforeEach",
					AdminUser: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: adminCredentialsName,
						},
						Key: config.GrafanaAdminUserEnvVar,
					},
					AdminPassword: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: adminCredentialsName,
						},
						Key: config.GrafanaAdminPasswordEnvVar,
					},
				},
			},
		}

		adminCredentials = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      adminCredentialsName,
				Namespace: grafanaNamespace,
			},
			Data: map[string][]byte{
				config.GrafanaAdminUserEnvVar:     []byte("admin"),
				config.GrafanaAdminPasswordEnvVar: []byte("password"),
			},
		}

		dashboard = &v1beta1.GrafanaDashboard{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "grafana.integreatly.org/v1beta1",
				Kind:       "GrafanaDashboard",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      dashboardName,
				Namespace: grafanaNamespace,
			},
		}
		dashboardLookupKey = types.NamespacedName{Name: dashboardName, Namespace: grafanaNamespace}
		createdDashboard   = &v1beta1.GrafanaDashboard{}

		ctx = context.Background()
	)

	BeforeEach(func() {
		mockDashboardStorage = make(map[string]gapi.Dashboard)
		server = ghttp.NewServer()

		server.RouteToHandler("GET", "/", ghttp.RespondWith(200, nil))
		server.RouteToHandler("GET", "/api/folders", ghttp.RespondWithJSONEncoded(http.StatusOK, &mockFolders))
		server.RouteToHandler("POST", "/api/folders", ghttp.CombineHandlers(
			ghttp.VerifyBasicAuth("admin", "password"),
			func(w http.ResponseWriter, r *http.Request) {
				mockFolders = append(mockFolders, expectedFolderOutput)
			},
			ghttp.RespondWithJSONEncoded(http.StatusOK, expectedFolderOutput),
		))

		server.RouteToHandler("GET", regexp.MustCompile("/api/dashboards/.*"), ghttp.CombineHandlers(
			ghttp.VerifyBasicAuth("admin", "password"),
			func(w http.ResponseWriter, r *http.Request) {
				if content, ok := mockDashboardStorage[dashboardUID]; ok {
					w.WriteHeader(200)
					json.NewEncoder(w).Encode(content)
				} else {
					w.WriteHeader(404)
				}
			},
		))
		server.RouteToHandler("DELETE", regexp.MustCompile("/api/dashboards/.*"), ghttp.CombineHandlers(
			ghttp.VerifyBasicAuth("admin", "password"),
			func(w http.ResponseWriter, r *http.Request) {
				delete(mockDashboardStorage, dashboardUID)
			},
			ghttp.RespondWith(200, nil),
		))
		server.RouteToHandler("POST", "/api/dashboards/db",
			ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("admin", "password"),
				func(w http.ResponseWriter, r *http.Request) {
					var dash gapi.Dashboard
					json.NewDecoder(r.Body).Decode(&dash)
					dash.Model["version"] = 3
					dash.Model["uid"] = dashboardUID
					mockDashboardStorage[dashboardUID] = dash
				},
				ghttp.RespondWithJSONEncoded(http.StatusOK, gapi.DashboardSaveResponse{
					Slug:    "fake-slug",
					ID:      42,
					UID:     dashboardUID,
					Status:  "success",
					Version: 3,
				}),
			),
		)

		adminCredentials.SetResourceVersion("")
		grafana.SetResourceVersion("")
		dashboard.SetResourceVersion("")

		grafana.Spec.External.URL = server.URL()
		Expect(k8sClient.Create(ctx, grafana)).Should(Succeed())
		Expect(k8sClient.Create(ctx, adminCredentials)).Should(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, dashboard)).Should(Succeed())
		Eventually(func() error {
			return k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
		}).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, grafana)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, adminCredentials)).Should(Succeed())
		server.Close()
	})

	Context("When creating GrafanaDashboard", func() {
		It("Should call the appropriate APIs on the Grafana instance", func() {
			By("By creating a dashboard resource")
			dashboard.Spec = v1beta1.GrafanaDashboardSpec{
				InstanceSelector: &metav1.LabelSelector{MatchLabels: grafana.ObjectMeta.Labels},
				Source: v1beta1.GrafanaDashboardSource{
					Inline: &v1beta1.GrafanaDashboardInlineSource{
						Json: &dashboardJsonString,
					},
				},
				FolderTitle: folderTitle,
				Plugins:     v1beta1.PluginList{piechartPlugin},
			}
			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
			}, oneSecondDuration, interval).Should(Succeed())

			By("By ensuring the folder endpoint was called")
			Eventually(func() []gapi.Folder {
				return mockFolders
			}).Should(ContainElement(HaveField("Title", folderTitle)))

			By("By ensuring the dashboard endpoint was called")
			Eventually(func() gapi.Dashboard {
				return mockDashboardStorage[dashboardUID]
			}).Should(HaveField("Message", MatchRegexp("Updated by Grafana Operator.*")))

			By("By checking the UID in the dashboard status")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				if err != nil {
					return "", err
				}
				return createdDashboard.Status.Instances[v1beta1.InstanceKeyFor(grafana)].UID, nil
			}).Should(Equal(dashboardUID))

			// TODO: test with incluster grafana
			// grafanaLookupKey := types.NamespacedName{Name: grafanaName, Namespace: grafanaNamespace}
			// createdGrafana := &v1beta1.Grafana{}
			// By("By checking the value of Grafana.Status.PluginList")
			// Eventually(func() (bool, error) {
			// 	err := k8sClient.Get(ctx, grafanaLookupKey, createdGrafana)
			// 	if err != nil {
			// 		return false, err
			// 	}
			// 	return createdGrafana.Status.Plugins.HasSomeVersionOf(&piechartPlugin), nil
			// }, oneSecondDuration, interval).Should(BeTrue())
		})

		It("Should download dashboards from remote URLs", func() {
			server.RouteToHandler("GET", "/other/dashboard.json", ghttp.RespondWithJSONEncoded(200, mockDashboard))

			By("By creating a dashboard resource")
			remoteDashboardURL := server.URL() + "/other/dashboard.json"
			dashboard.Spec = v1beta1.GrafanaDashboardSpec{
				InstanceSelector: &metav1.LabelSelector{MatchLabels: grafana.ObjectMeta.Labels},
				Source: v1beta1.GrafanaDashboardSource{
					Remote: &v1beta1.GrafanaDashboardRemoteSource{
						ContentCacheDuration: oneSecondMetaDuration,
						Url:                  &remoteDashboardURL,
					},
				},
			}
			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
			}, oneSecondDuration, interval).Should(Succeed())

			By("By ensuring the dashboard content was downloaded")
			Eventually(func() []*http.Request {
				return server.ReceivedRequests()
			}).Should(ContainElement(HaveField("URL", HaveField("Path", "/other/dashboard.json"))))

			By("By ensuring the dashboard endpoint was called")
			Eventually(func() gapi.Dashboard {
				return mockDashboardStorage[dashboardUID]
			}).Should(HaveField("Message", MatchRegexp("Updated by Grafana Operator.*")))

			By("By ensuring the uploaded dashboard was the one provided")
			Eventually(func() interface{} {
				return mockDashboardStorage[dashboardUID].Model
			}).Should(And(
				HaveKeyWithValue("not", "really"),
				HaveKeyWithValue("a", "dashboard"),
			))
		})
	})
})
