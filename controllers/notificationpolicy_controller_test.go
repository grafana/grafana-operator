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
	"errors"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func routesToRuntimeObjects(routes []v1beta1.GrafanaNotificationPolicyRoute) []runtime.Object {
	objects := make([]runtime.Object, len(routes))
	for i := range routes {
		objects[i] = &routes[i]
	}
	return objects
}

func stringP(s string) *string {
	return &s
}

func TestAssembleNotificationPolicyRoutes(t *testing.T) {
	tests := []struct {
		name                string
		notificationPolicy  *v1beta1.GrafanaNotificationPolicy
		existingRoutes      []v1beta1.GrafanaNotificationPolicyRoute
		want                *v1beta1.GrafanaNotificationPolicy
		wantErr             bool
		wantLoopDetectedErr bool
	}{
		{
			name: "Simple assembly with one level of routes",
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []v1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ignore routes from other namespace when cross-namespace import is not allowed",
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
						AllowCrossNamespaceImport: false,
					},
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []v1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-2",
						Namespace: "other-namespace",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-A-receiver-other-namespace",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Assembly with nested routes",
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []v1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
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
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-B-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
								Routes: []*v1beta1.Route{
									{
										Receiver: "team-B-receiver",
										Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
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
			name: "Assembly with nested routes and multiple RouteSelectors inside Routes",
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "second", "team": "A"},
								},
							},
							{
								Receiver: "team-B-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "second", "team": "B"},
								},
							},
						},
					},
				},
			},
			existingRoutes: []v1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "second", "team": "A"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "project-X-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("project"), Value: "X", IsEqual: true}},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-2",
						Namespace: "default",
						Labels:    map[string]string{"tier": "second", "team": "B"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "project-Y-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("project"), Value: "Y", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{
							{
								Receiver: "team-A-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
								Routes: []*v1beta1.Route{
									{
										Receiver: "project-X-receiver",
										Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("project"), Value: "X", IsEqual: true}},
									},
								},
							},
							{
								Receiver: "team-B-receiver",
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
								Routes: []*v1beta1.Route{
									{
										Receiver: "project-Y-receiver",
										Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("project"), Value: "Y", IsEqual: true}},
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
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.Route{
						Receiver: "default-receiver",
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"tier": "first"},
						},
					},
				},
			},
			existingRoutes: []v1beta1.GrafanaNotificationPolicyRoute{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "route-1",
						Namespace: "default",
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-A-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "A", IsEqual: true}},
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
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Receiver: "team-B-receiver",
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: stringP("team"), Value: "B", IsEqual: true}},
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
							},
						},
					},
				},
			},
			wantErr:             true,
			wantLoopDetectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := runtime.NewScheme()
			err := v1beta1.AddToScheme(s)
			assert.NoError(t, err, "adding scheme")
			client := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(routesToRuntimeObjects(tt.existingRoutes)...).Build()

			_, err = assembleNotificationPolicyRoutes(ctx, client, tt.notificationPolicy)
			if tt.wantLoopDetectedErr {
				assert.True(t, errors.Is(err, ErrLoopDetected))
			}
			if tt.wantErr {
				assert.Error(t, err, "assembleNotificationPolicyRoutes() should return an error")
			} else {
				assert.NoError(t, err, "assembleNotificationPolicyRoutes() should not return an error")
				assert.Equal(t, tt.want, tt.notificationPolicy, "assembleNotificationPolicyRoutes() returned unexpected policy")
			}
		})
	}
}

var _ = Describe("NotificationPolicy Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaNotificationPolicy
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					GrafanaCommonSpec: commonSpecSuspended,
					Route:             &v1beta1.Route{Receiver: "default-receiver"},
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					GrafanaCommonSpec: commonSpecNoMatchingInstances,
					Route:             &v1beta1.Route{Receiver: "default-receiver"},
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaNotificationPolicyReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaNotificationPolicy{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
