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
	"encoding/json"
	"testing"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"dashboard": "external",
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
					InstanceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"dashboard": "internal",
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

func TestGrafanaDashboardIsUpdatedUID(t *testing.T) {
	crUID := "crUID"
	tests := []struct {
		name         string
		crUID        string
		statusUID    string
		dashboardUID string
		want         bool
	}{
		{
			name:         "Status.UID and dashboard UID are empty",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "newUID",
			want:         false,
		},
		{
			name:         "Status.UID is empty, dashboard UID is not",
			crUID:        crUID,
			statusUID:    "",
			dashboardUID: "newUID",
			want:         false,
		},
		{
			name:         "Status.UID is not empty (same as CR uid), new UID is empty",
			crUID:        crUID,
			statusUID:    crUID,
			dashboardUID: "",
			want:         false,
		},
		{
			name:         "Status.UID is not empty (different from CR uid), new UID is empty",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "",
			want:         true,
		},
		{
			name:         "Status.UID is not empty, new UID is different",
			crUID:        crUID,
			statusUID:    "oldUID",
			dashboardUID: "newUID",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := getDashboardCR(t, tt.crUID, tt.statusUID)
			dashboardJson := getDashboardJSONWithUID(t, tt.dashboardUID)
			got := cr.IsUpdatedUID(dashboardJson)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getDashboardCR(t *testing.T, crUID string, statusUID string) v1beta1.GrafanaDashboard {
	t.Helper()

	cr := v1beta1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: "grafana-operator-system",
			UID:       types.UID(crUID),
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			InstanceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"dashboard": "grafana",
				},
			},
		},
		Status: v1beta1.GrafanaDashboardStatus{
			UID: statusUID,
		},
	}

	return cr
}

func getDashboardJSONWithUID(t *testing.T, uid string) []byte {
	t.Helper()

	type DashboardUID struct {
		UID string `json:"uid,omitempty"`
	}

	dashboardUID := DashboardUID{
		UID: uid,
	}

	dashboardJson, err := json.Marshal(dashboardUID)
	if err != nil {
		t.Errorf("failed to generate dashboard JSON with uid: %s", uid)
	}

	return dashboardJson
}
