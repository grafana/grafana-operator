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
	"reflect"
	"testing"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

func Test_mergeNotificationPolicyRoutesWithRouteList(t *testing.T) {
	type args struct {
		notificationPolicy          *grafanav1beta1.GrafanaNotificationPolicy
		notificationPolicyRouteList *grafanav1beta1.GrafanaNotificationPolicyRouteList
	}
	tests := []struct {
		name string
		args args
		want *grafanav1beta1.GrafanaNotificationPolicy
	}{
		{
			name: "Merge with empty route list",
			args: args{
				notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
					Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
						Route: &grafanav1beta1.Route{
							Receiver: "default-receiver",
							Routes:   []*grafanav1beta1.Route{},
						},
					},
				},
				notificationPolicyRouteList: &grafanav1beta1.GrafanaNotificationPolicyRouteList{},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						Routes:   []*grafanav1beta1.Route{},
					},
				},
			},
		},
		{
			name: "Merge into nil route list",
			args: args{
				notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
					Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
						Route: &grafanav1beta1.Route{
							Receiver: "default-receiver",
							Routes:   nil,
						},
					},
				},
				notificationPolicyRouteList: &grafanav1beta1.GrafanaNotificationPolicyRouteList{},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						Routes:   nil,
					},
				},
			},
		},
		{
			name: "Merge with un-ordered non-empty route list",
			args: args{
				notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
					Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
						Route: &grafanav1beta1.Route{
							Receiver: "default-receiver",
							Routes:   []*grafanav1beta1.Route{},
						},
					},
				},
				notificationPolicyRouteList: &grafanav1beta1.GrafanaNotificationPolicyRouteList{
					Items: []grafanav1beta1.GrafanaNotificationPolicyRoute{
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "team-B-receiver",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
								},
								Priority: int8P(2),
							},
						},
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "team-A-receiver",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
								},
								Priority: int8P(1),
							},
						},
					},
				},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*grafanav1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							},
							{
								Receiver: "team-B-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
							},
						},
					},
				},
			},
		},
		{
			name: "Merge with existing routes in GrafanaNotificationPolicy, existing routes ordered first",
			args: args{
				notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
					Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
						Route: &grafanav1beta1.Route{
							Receiver: "default-receiver",
							Routes: []*grafanav1beta1.Route{
								{
									Receiver: "existing-receiver",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "critical", IsEqual: true}},
								},
							},
						},
					},
				},
				notificationPolicyRouteList: &grafanav1beta1.GrafanaNotificationPolicyRouteList{
					Items: []grafanav1beta1.GrafanaNotificationPolicyRoute{
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "new-receiver",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "C", IsEqual: true}},
								},
								Priority: int8P(1),
							},
						},
					},
				},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*grafanav1beta1.Route{
							{
								Receiver: "existing-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "critical", IsEqual: true}},
							},
							{
								Receiver: "new-receiver",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("team"), Value: "C", IsEqual: true}},
							},
						},
					},
				},
			},
		},
		{
			name: "Merge with multiple routes, nil priority has least priority",
			args: args{
				notificationPolicy: &grafanav1beta1.GrafanaNotificationPolicy{
					Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
						Route: &grafanav1beta1.Route{
							Receiver: "default-receiver",
							Routes:   []*grafanav1beta1.Route{},
						},
					},
				},
				notificationPolicyRouteList: &grafanav1beta1.GrafanaNotificationPolicyRouteList{
					Items: []grafanav1beta1.GrafanaNotificationPolicyRoute{
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "low-priority",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "info", IsEqual: true}},
								},
								Priority: nil,
							},
						},
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "high-priority",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "critical", IsEqual: true}},
								},
								Priority: int8P(1),
							},
						},
						{
							Spec: grafanav1beta1.GrafanaNotificationPolicyRouteSpec{
								Route: &grafanav1beta1.Route{
									Receiver: "medium-priority",
									Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "warning", IsEqual: true}},
								},
								Priority: int8P(2),
							},
						},
					},
				},
			},
			want: &grafanav1beta1.GrafanaNotificationPolicy{
				Spec: grafanav1beta1.GrafanaNotificationPolicySpec{
					Route: &grafanav1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*grafanav1beta1.Route{
							{
								Receiver: "high-priority",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "critical", IsEqual: true}},
							},
							{
								Receiver: "medium-priority",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "warning", IsEqual: true}},
							},
							{
								Receiver: "low-priority",
								Matchers: grafanav1beta1.Matchers{&grafanav1beta1.Matcher{Name: stringP("severity"), Value: "info", IsEqual: true}},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeNotificationPolicyRoutesWithRouteList(tt.args.notificationPolicy, tt.args.notificationPolicyRouteList)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeNotificationPolicyRoutesWithRouteList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringP(s string) *string {
	return &s
}

func int8P(i int) *int8 {
	i8 := int8(i)
	return &i8
}
