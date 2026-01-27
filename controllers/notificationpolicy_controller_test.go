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
	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
)

func convertRoutesToClientObjects(routes []v1beta1.GrafanaNotificationPolicyRoute) []client.Object {
	objects := make([]client.Object, len(routes))

	for i := range routes {
		objects[i] = &routes[i]
	}

	return objects
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
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
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
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-A-receiver",
							},
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							Routes: []*v1beta1.Route{{
								PartialRoute: v1beta1.PartialRoute{
									Receiver: "team-A-receiver",
								},
								Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
							}},
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
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
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
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-A-receiver",
							},
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
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-A-receiver-other-namespace",
							},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							Routes: []*v1beta1.Route{
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-A-receiver",
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
			name: "Assembly with nested routes",
			notificationPolicy: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
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
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-A-receiver",
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "second"},
								},
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
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-B-receiver",
							},
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "B", IsEqual: true}},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							Routes: []*v1beta1.Route{
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-A-receiver",
										Routes: []*v1beta1.Route{
											{
												Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "B", IsEqual: true}},
												PartialRoute: v1beta1.PartialRoute{
													Receiver: "team-B-receiver",
												},
											},
										},
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
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							Routes: []*v1beta1.Route{
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-A-receiver",
										RouteSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{"tier": "second", "team": "A"},
										},
									},
								},
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "B", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-B-receiver",
										RouteSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{"tier": "second", "team": "B"},
										},
									},
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
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("project"), Value: "X", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "project-X-receiver",
							},
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
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("project"), Value: "Y", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "project-Y-receiver",
							},
						},
					},
				},
			},
			want: &v1beta1.GrafanaNotificationPolicy{
				Spec: v1beta1.GrafanaNotificationPolicySpec{
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							Routes: []*v1beta1.Route{
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-A-receiver",
										Routes: []*v1beta1.Route{
											{
												Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("project"), Value: "X", IsEqual: true}},
												PartialRoute: v1beta1.PartialRoute{
													Receiver: "project-X-receiver",
												},
											},
										},
									},
								},
								{
									Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "B", IsEqual: true}},
									PartialRoute: v1beta1.PartialRoute{
										Receiver: "team-B-receiver",
										Routes: []*v1beta1.Route{
											{
												Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("project"), Value: "Y", IsEqual: true}},
												PartialRoute: v1beta1.PartialRoute{
													Receiver: "project-Y-receiver",
												},
											},
										},
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
					Route: &v1beta1.TopLevelRoute{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "default-receiver",
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"tier": "first"},
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
						Labels:    map[string]string{"tier": "first"},
					},
					Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
						Route: v1beta1.Route{
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-A-receiver",
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "second"},
								},
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
							Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "B", IsEqual: true}},
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "team-B-receiver",
								RouteSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"tier": "first"},
								},
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
			testCtx := context.Background()
			s := runtime.NewScheme()

			err := v1beta1.AddToScheme(s)
			require.NoError(t, err, "adding scheme")

			initObjs := convertRoutesToClientObjects(tt.existingRoutes)

			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithObjects(initObjs...).
				Build()

			_, err = assembleNotificationPolicyRoutes(testCtx, cl, tt.notificationPolicy)
			if tt.wantLoopDetectedErr {
				require.ErrorIs(t, err, ErrLoopDetected)
			}

			if tt.wantErr {
				require.Error(t, err, "assembleNotificationPolicyRoutes() should return an error")
			} else {
				require.NoError(t, err, "assembleNotificationPolicyRoutes() should not return an error")
				assert.Equal(t, tt.want, tt.notificationPolicy, "assembleNotificationPolicyRoutes() returned unexpected policy")
			}
		})
	}
}

var _ = Describe("NotificationPolicy Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaNotificationPolicySpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaNotificationPolicySpec{
				GrafanaCommonSpec: commonSpecSuspended,
				Route: &v1beta1.TopLevelRoute{
					PartialRoute: v1beta1.PartialRoute{Receiver: "default-receiver"},
				},
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaNotificationPolicySpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
				Route: &v1beta1.TopLevelRoute{
					PartialRoute: v1beta1.PartialRoute{Receiver: "default-receiver"},
				},
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaNotificationPolicySpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
				Route: &v1beta1.TopLevelRoute{
					PartialRoute: v1beta1.PartialRoute{Receiver: "default-receiver"},
				},
			},
			want: metav1.Condition{
				Type:   conditionNotificationPolicySynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: LogMsgApplyErrors,
		},
		{
			name: "Mutually Exclusive fields routes/routeSelector",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaNotificationPolicySpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				Route: &v1beta1.TopLevelRoute{
					PartialRoute: v1beta1.PartialRoute{
						Receiver: "default-receiver",
						Routes: []*v1beta1.Route{{
							PartialRoute: v1beta1.PartialRoute{
								Receiver: "default-receiver",
							},
						}},
						RouteSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
					},
				},
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonFieldsMutuallyExclusive,
			},
			wantErr: "invalid route spec discovered: routeSelector is mutually exclusive with routes",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaNotificationPolicySpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Route: &v1beta1.TopLevelRoute{
					PartialRoute: v1beta1.PartialRoute{Receiver: "grafana-default-email"},
				},
			},
			want: metav1.Condition{
				Type:   conditionNotificationPolicySynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaNotificationPolicy{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaNotificationPolicyReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("NotificationPolicy Reconciler: Provoke LoopDetected Condition", func() {
	t := GinkgoT()

	cr := &v1beta1.GrafanaNotificationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "loopdetected-spec",
		},
		Spec: v1beta1.GrafanaNotificationPolicySpec{
			GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"loop-detected": "test"},
				},
			},
			Route: &v1beta1.TopLevelRoute{
				PartialRoute: v1beta1.PartialRoute{
					Receiver: "grafana-default-email",
					Routes: []*v1beta1.Route{{
						PartialRoute: v1beta1.PartialRoute{
							Receiver: "grafana-default-email",
							RouteSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"team-a": "child"},
							},
						},
						Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "a", IsEqual: true}},
					}},
				},
			},
		},
	}
	teamB := &v1beta1.GrafanaNotificationPolicyRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "team-b",
			Labels:    map[string]string{"team-a": "child"},
		},
		Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
			Route: v1beta1.Route{
				Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "b", IsEqual: true}},
				PartialRoute: v1beta1.PartialRoute{
					Receiver: "grafana-default-email",
					RouteSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"team-b": "child"},
					},
				},
			},
		},
	}
	teamC := &v1beta1.GrafanaNotificationPolicyRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "team-c",
			Labels:    map[string]string{"team-b": "child"},
		},
		Spec: v1beta1.GrafanaNotificationPolicyRouteSpec{
			Route: v1beta1.Route{
				Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "b", IsEqual: true}}, // Also matches team b
				PartialRoute: v1beta1.PartialRoute{
					Receiver: "grafana-default-email",
					RouteSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"team-b": "child"},
					},
				},
			},
		},
	}

	It("Provokes the NotificationPolicyLoopDetected Condition", func() {
		err := cl.Create(testCtx, teamB)
		require.NoError(t, err)

		err = cl.Create(testCtx, teamC)
		require.NoError(t, err)

		r := &GrafanaNotificationPolicyReconciler{Client: cl, Scheme: cl.Scheme()}

		want := metav1.Condition{
			Type:   conditionNotificationPolicyLoopDetected,
			Reason: conditionReasonLoopDetected,
		}

		wantErr := "failed to assemble notification policy routes"

		reconcileAndValidateCondition(r, cr, want, wantErr)
	})
})
