package v1beta1

import (
	"context"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListNotificationPolicy(t *testing.T) {
	t.Run("&NotificationPolicy{} does not map to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaNotificationPolicy{}
		_, _, err := g.Status.StatusList(arg)
		assert.Error(t, err, "NotificationPolicy should not have a case in Grafana.Status.StatusList")
	})
}

func newNotificationPolicy(name string, editable *bool) *GrafanaNotificationPolicy {
	return &GrafanaNotificationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaNotificationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaNotificationPolicySpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "notificationpolicy",
					},
				},
			},
			Route: &TopLevelRoute{
				PartialRoute: PartialRoute{
					Receiver: "grafana-default-email",
					GroupBy:  []string{"group_name", "alert_name"},
					Routes:   []*Route{},
				},
			},
		},
	}
}

var _ = Describe("NotificationPolicy type", func() {
	t := GinkgoT()

	Context("Ensure NotificationPolicy spec.editable is immutable", func() {
		ctx := context.Background()

		It("Should block adding editable field when missing", func() {
			notificationpolicy := newNotificationPolicy("missing-editable", nil)

			By("Create new NotificationPolicy without editable")

			err := cl.Create(ctx, notificationpolicy)
			require.NoError(t, err)

			By("Adding a editable")

			notificationpolicy.Spec.Editable = new(true)
			err = cl.Update(ctx, notificationpolicy)
			require.Error(t, err)
		})

		It("Should block removing editable field when set", func() {
			notificationpolicy := newNotificationPolicy("existing-editable", new(true))

			By("Creating NotificationPolicy with existing editable")

			err := cl.Create(ctx, notificationpolicy)
			require.NoError(t, err)

			By("And setting editable to ''")

			notificationpolicy.Spec.Editable = nil
			err = cl.Update(ctx, notificationpolicy)
			require.Error(t, err)
		})

		It("Should block changing value of editable", func() {
			notificationpolicy := newNotificationPolicy("removing-editable", new(true))

			By("Create new NotificationPolicy with existing editable")

			err := cl.Create(ctx, notificationpolicy)
			require.NoError(t, err)

			By("Changing the existing editable")

			notificationpolicy.Spec.Editable = new(false)
			err = cl.Update(ctx, notificationpolicy)
			require.Error(t, err)
		})
	})

	Context("Invalidate root routes when using invalid fields", func() {
		ctx := context.Background()
		invalidFieldErr := "is invalid on the top level route node"

		It("Invalidate continue", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.Continue = true

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})

		It("Invalidate active_time_intervals", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.ActiveTimeIntervals = []string{"any-string"}

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})

		It("Invalidate mute_time_intervals", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.MuteTimeIntervals = []string{"any-string"}

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})

		It("Invalidate match_re", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.MatchRe = models.MatchRegexps{"match": "string"}

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})

		It("Invalidate matchers", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.Matchers = Matchers{&Matcher{}}
			// Matchers: v1beta1.Matchers{&v1beta1.Matcher{Name: ptr.To("team"), Value: "A", IsEqual: true}},

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})

		It("Invalidate matchers", func() {
			np := newNotificationPolicy("invalid-route-fields", nil)
			np.Spec.Route.ObjectMatchers = models.ObjectMatchers{[]string{"any"}}

			err := cl.Create(ctx, np)
			require.Error(t, err)
			assert.ErrorContains(t, err, invalidFieldErr)
		})
	})
})

func TestIsRouteSelectorMutuallyExclusive(t *testing.T) {
	tests := []struct {
		name     string
		route    *PartialRoute
		expected bool
	}{
		{
			name:     "Empty route",
			route:    &PartialRoute{},
			expected: true,
		},
		{
			name: "Route with only RouteSelector",
			route: &PartialRoute{
				RouteSelector: &metav1.LabelSelector{},
			},
			expected: true,
		},
		{
			name: "Route with only sub-routes",
			route: &PartialRoute{
				Routes: []*Route{
					{},
					{},
				},
			},
			expected: true,
		},
		{
			name: "Route with both RouteSelector and sub-routes",
			route: &PartialRoute{
				RouteSelector: &metav1.LabelSelector{},
				Routes: []*Route{
					{},
				},
			},
			expected: false,
		},
		{
			name: "Nested routes with mutual exclusivity",
			route: &PartialRoute{
				Routes: []*Route{
					{
						PartialRoute: PartialRoute{
							RouteSelector: &metav1.LabelSelector{},
						},
					},
					{
						PartialRoute: PartialRoute{
							Routes: []*Route{
								{},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Nested routes without mutual exclusivity",
			route: &PartialRoute{
				Routes: []*Route{
					{
						PartialRoute: PartialRoute{
							RouteSelector: &metav1.LabelSelector{},
							Routes: []*Route{
								{},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.route.IsRouteSelectorMutuallyExclusive()
			if result != tt.expected {
				t.Errorf("IsRouteSelectorMutuallyExclusive() = %v, want %v", result, tt.expected)
			}
		})
	}
}
