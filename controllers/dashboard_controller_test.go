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
	"net/http/httptest"
	"time"

	gapi "github.com/grafana/grafana-api-golang-client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		dashboardName = "test-dashboard"
		dashboardUID  = "test-dashboard-uid"

		timeout  = time.Second
		duration = metav1.Duration{time.Second}
		interval = time.Millisecond * 250
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

		mockDashboardStorage map[string]gapi.Dashboard
		mockAPIRequests      map[string][]interface{}

		handlers       *http.ServeMux
		mockGrafanaAPI *httptest.Server

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
		mockAPIRequests = make(map[string][]interface{})

		handlers = http.NewServeMux()
		handlers.HandleFunc("/api/folders", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				var folder gapi.Folder
				json.NewDecoder(r.Body).Decode(&folder)
				mockAPIRequests["/api/folders"] = append(mockAPIRequests["/api/folders"], folder)
				json.NewEncoder(w).Encode(&gapi.Folder{
					ID: 13,
				})
			} else {
				json.NewEncoder(w).Encode(&[]gapi.Folder{{
					ID: 13,
				}})
			}
		})
		handlers.HandleFunc("/api/dashboards/db", func(w http.ResponseWriter, r *http.Request) {
			var dash gapi.Dashboard
			json.NewDecoder(r.Body).Decode(&dash)
			mockAPIRequests[r.URL.Path] = append(mockAPIRequests[r.URL.Path], dash)
			mockDashboardStorage[dashboardUID] = dash
			json.NewEncoder(w).Encode(&gapi.DashboardSaveResponse{
				Slug:    "fake-slug",
				ID:      42,
				UID:     dashboardUID,
				Status:  "ok",
				Version: 3,
			})
		})

		mockGrafanaAPI = httptest.NewServer(handlers)

		adminCredentials.SetResourceVersion("")
		grafana.SetResourceVersion("")
		dashboard.SetResourceVersion("")

		grafana.Spec.External.URL = mockGrafanaAPI.URL
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
		mockGrafanaAPI.Close()
	})

	Context("When creating GrafanaDashboard", func() {
		It("Should call the appropriate APIs on the Grafana instance", func() {

			handlers.HandleFunc("/api/dashboards/uid/some-uid", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				json.NewEncoder(w).Encode(mockDashboardStorage[dashboardUID])
			})

			By("By creating a dashboard resource")
			dashboard.Spec = v1beta1.GrafanaDashboardSpec{
				InstanceSelector: &metav1.LabelSelector{MatchLabels: grafana.ObjectMeta.Labels},
				Source: v1beta1.GrafanaDashboardSource{
					Inline: &v1beta1.GrafanaDashboardInlineSource{
						Json: &dashboardJsonString,
					},
				},
				Plugins: v1beta1.PluginList{piechartPlugin},
			}
			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
			}, timeout, interval).Should(Succeed())

			By("By ensuring the folder endpoint was called")
			Eventually(func() []interface{} {
				return mockAPIRequests["/api/folders"]
			}).Should(ContainElement(Equal(gapi.Folder{Title: grafana.Namespace})))
			By("By ensuring the dashboard endpoint was called")
			Eventually(func() []interface{} {
				return mockAPIRequests["/api/dashboards/db"]
			}).Should(ContainElement(HaveField("Message", MatchRegexp("Updated by Grafana Operator.*"))))

			By("By checking the UID in the dashboard status")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
				if err != nil {
					return "", err
				}
				return createdDashboard.Status.Instances[v1beta1.InstanceKeyFor(grafana)].UID, nil
			}).Should(Equal(dashboardUID))

			grafanaLookupKey := types.NamespacedName{Name: grafanaName, Namespace: grafanaNamespace}
			createdGrafana := &v1beta1.Grafana{}
			By("By checking the value of Grafana.Status.PluginList")
			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, grafanaLookupKey, createdGrafana)
				if err != nil {
					return false, err
				}
				return createdGrafana.Status.Plugins.HasSomeVersionOf(&piechartPlugin), nil
			}, timeout, interval).Should(BeTrue())
		})

		It("Should download dashboards from remote URLs", func() {
			handlers.HandleFunc("/other/dashboard.json", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				mockAPIRequests[r.URL.Path] = append(mockAPIRequests[r.URL.Path], "called")
				json.NewEncoder(w).Encode(mockDashboard)
			})

			By("By creating a dashboard resource")
			remoteDashboardURL := mockGrafanaAPI.URL + "/other/dashboard.json"
			dashboard.Spec = v1beta1.GrafanaDashboardSpec{
				InstanceSelector: &metav1.LabelSelector{MatchLabels: grafana.ObjectMeta.Labels},
				Source: v1beta1.GrafanaDashboardSource{
					Remote: &v1beta1.GrafanaDashboardRemoteSource{
						ContentCacheDuration: duration,
						Url:                  &remoteDashboardURL,
					},
				},
			}
			Expect(k8sClient.Create(ctx, dashboard)).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, dashboardLookupKey, createdDashboard)
			}, timeout, interval).Should(Succeed())

			By("By ensuring the dashboard content was downloaded")
			Eventually(func() []interface{} {
				return mockAPIRequests["/other/dashboard.json"]
			}).Should(ContainElement(Equal("called")))

			By("By ensuring the dashboard endpoint was called")
			Eventually(func() []interface{} {
				return mockAPIRequests["/api/dashboards/db"]
			}).Should(ContainElement(HaveField("Message", MatchRegexp("Updated by Grafana Operator.*"))))

			By("By ensuring the uploaded dashboard was the one provided")
			Eventually(func() interface{} {
				return mockDashboardStorage[dashboardUID].Model
			}).Should(BeEquivalentTo(mockDashboard))
		})
	})

})
