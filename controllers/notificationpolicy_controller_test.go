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
	"reflect"
	"testing"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func routesToRuntimeObjects(routes []grafanav1beta1.GrafanaNotificationPolicyRoute) []runtime.Object {
	objects := make([]runtime.Object, len(routes))
	for i := range routes {
		objects[i] = &routes[i]
	}
	return objects
}

func stringP(s string) *string {
	return &s
}

func int8P(i int) *int8 {
	i8 := int8(i)
	return &i8
}

func TestAssembleNotificationPolicyRoutes(t *testing.T) {
	tests := []struct {
		name               string
		notificationPolicy *grafanav1beta1.GrafanaNotificationPolicy
		existingRoutes     []grafanav1beta1.GrafanaNotificationPolicyRoute
		want               *grafanav1beta1.GrafanaNotificationPolicy
		wantErr            bool
	}{
		{
			name: "Simple assembly with one level of routes",
			notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []grafanav1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: &grafanav1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
						},
					},
				},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
						Routes: []*grafanav1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Assembly with nested routes",
			notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []grafanav1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: &grafanav1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "second"},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-2",
						Namespace: "default",
						Labels:    map[string]string{"tier": "second"},
					},
					Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: &grafanav1beta1.Route{
							Receiver: "team-B-receiver",
							Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
						},
					},
				},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
						Routes: []*grafanav1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "second"},
								},
								Routes: []*grafanav1beta1.Route{
									{
										Receiver: "team-B-receiver",
										Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Detect loop in routes",
			notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []grafanav1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: &grafanav1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "second"},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-2",
						Namespace: "default",
						Labels:    map[string]string{"tier": "second"},
					},
					Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: &grafanav1beta1.Route{
							Receiver: "team-B-receiver",
							Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client := fake.NewClientBuilder().WithRuntimeObjects(routesToRuntimeObjects(tt.existingRoutes)...).Build()

			gotPolicy, _, err := assembleNotificationPolicyRoutes(ctx, client, nil, tt.notificationPolicy)
			if (err != nil) != tt.wantErr {
				t.Errorf("assembleNotificationPolicyRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPolicy, tt.want) {
				t.Errorf("assembleNotificationPolicyRoutes() = %v, want %v", gotPolicy, tt.want)
			}
		})
	}
}
