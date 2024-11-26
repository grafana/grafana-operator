package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newNotificationPolicy(name string, editable *bool) *GrafanaNotificationPolicy {
	return &GrafanaNotificationPolicy{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaNotificationPolicy",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaNotificationPolicySpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "notificationpolicy",
					},
				},
			},
			Route: &Route{
				Continue:          false,
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
