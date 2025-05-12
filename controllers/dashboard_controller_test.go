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
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetDashboardsToDelete(t *testing.T) {
	dashboardList := v1beta1.GrafanaDashboardList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []v1beta1.GrafanaDashboard{
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external",
					Namespace: "grafana-operator-system",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"dashboard": "external",
							},
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "internal",
					Namespace: "grafana-operator-system",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"dashboard": "external",
							},
						},
					},
				},
			},
		},
	}
	grafanaList := []v1beta1.Grafana{
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "external",
				Namespace: "grafana-operator-system",
				Labels: map[string]string{
					"dashboard": "external",
				},
			},
			Status: v1beta1.GrafanaStatus{
				Dashboards: v1beta1.NamespacedResourceList{
					"grafana-operator-system/external/cb1688d2-547a-465b-bc49-df3ccf3da883",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "internal-broken1",
				Namespace: "grafana-operator-system",
				Labels: map[string]string{
					"dashboard": "internal",
				},
			},
			Status: v1beta1.GrafanaStatus{
				Dashboards: v1beta1.NamespacedResourceList{
					"grafana-operator-system/broken1/cb1688d2-547a-465b-bc49-df3ccf3da883",
				},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "internal-broken2",
				Namespace: "grafana-operator-system",
				Labels: map[string]string{
					"dashboard": "internal",
				},
			},
			Status: v1beta1.GrafanaStatus{
				Dashboards: v1beta1.NamespacedResourceList{
					"grafana-operator-system/broken2/cb1688d2-547a-465b-bc49-df3ccf3da883",
				},
			},
		},
	}

	dashboardsToDelete := getDashboardsToDelete(&dashboardList, grafanaList)
	for dashboard := range dashboardsToDelete {
		if dashboard.Name == "internal-broken1" {
			assert.Equal(t, []v1beta1.NamespacedResource([]v1beta1.NamespacedResource{"grafana-operator-system/broken1/cb1688d2-547a-465b-bc49-df3ccf3da883"}), dashboardsToDelete[dashboard])
		}
		if dashboard.Name == "internal-broken2" {
			assert.Equal(t, []v1beta1.NamespacedResource([]v1beta1.NamespacedResource{"grafana-operator-system/broken2/cb1688d2-547a-465b-bc49-df3ccf3da883"}), dashboardsToDelete[dashboard])
		}
	}
}

var _ = Describe("Dashboard: Reconciler", func() {
	It("Results in NoMatchingInstances Condition", func() {
		// Create object
		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-match",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaCommonSpec:  instanceSelectorNoMatchingInstances,
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
			},
		}
		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		Expect(err).ToNot(HaveOccurred())

		// Reconciliation Request
		req := requestFromMeta(cr.ObjectMeta)

		// Reconcile
		r := GrafanaDashboardReconciler{Client: k8sClient}
		_, err = r.Reconcile(ctx, req)
		Expect(err).ShouldNot(HaveOccurred()) // NoMatchingInstances is a valid reconciliation result

		resultCr := &v1beta1.GrafanaDashboard{}
		Expect(r.Get(ctx, req.NamespacedName, resultCr)).Should(Succeed()) // NoMatchingInstances is a valid status

		// Verify NoMatchingInstances condition
		Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", conditionNoMatchingInstance)))
	})
})
