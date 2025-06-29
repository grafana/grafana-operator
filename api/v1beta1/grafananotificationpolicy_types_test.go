package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
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
			Route: &Route{
				Continue:          false,
				Receiver:          "grafana-default-email",
				GroupBy:           []string{"group_name", "alert_name"},
				MuteTimeIntervals: []string{},
				Routes:            []*Route{},
			},
		},
	}
}

var _ = Describe("NotificationPolicy type", func() {
	Context("Ensure NotificationPolicy spec.editable is immutable", func() {
		ctx := context.Background()
		refTrue := true
		refFalse := false

		It("Should block adding editable field when missing", func() {
			notificationpolicy := newNotificationPolicy("missing-editable", nil)
			By("Create new NotificationPolicy without editable")
			Expect(k8sClient.Create(ctx, notificationpolicy)).To(Succeed())

			By("Adding a editable")
			notificationpolicy.Spec.Editable = &refTrue
			Expect(k8sClient.Update(ctx, notificationpolicy)).To(HaveOccurred())
		})

		It("Should block removing editable field when set", func() {
			notificationpolicy := newNotificationPolicy("existing-editable", &refTrue)
			By("Creating NotificationPolicy with existing editable")
			Expect(k8sClient.Create(ctx, notificationpolicy)).To(Succeed())

			By("And setting editable to ''")
			notificationpolicy.Spec.Editable = nil
			Expect(k8sClient.Update(ctx, notificationpolicy)).To(HaveOccurred())
		})

		It("Should block changing value of editable", func() {
			notificationpolicy := newNotificationPolicy("removing-editable", &refTrue)
			By("Create new NotificationPolicy with existing editable")
			Expect(k8sClient.Create(ctx, notificationpolicy)).To(Succeed())

			By("Changing the existing editable")
			notificationpolicy.Spec.Editable = &refFalse
			Expect(k8sClient.Update(ctx, notificationpolicy)).To(HaveOccurred())
		})
	})
})

func TestIsRouteSelectorMutuallyExclusive(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected bool
	}{
		{
			name:     "Empty route",
			route:    &Route{},
			expected: true,
		},
		{
			name: "Route with only RouteSelector",
			route: &Route{
				RouteSelector: &metav1.LabelSelector{},
			},
			expected: true,
		},
		{
			name: "Route with only sub-routes",
			route: &Route{
				Routes: []*Route{
					{},
					{},
				},
			},
			expected: true,
		},
		{
			name: "Route with both RouteSelector and sub-routes",
			route: &Route{
				RouteSelector: &metav1.LabelSelector{},
				Routes: []*Route{
					{},
				},
			},
			expected: false,
		},
		{
			name: "Nested routes with mutual exclusivity",
			route: &Route{
				Routes: []*Route{
					{
						RouteSelector: &metav1.LabelSelector{},
					},
					{
						Routes: []*Route{
							{},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Nested routes without mutual exclusivity",
			route: &Route{
				Routes: []*Route{
					{
						RouteSelector: &metav1.LabelSelector{},
						Routes: []*Route{
							{},
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
