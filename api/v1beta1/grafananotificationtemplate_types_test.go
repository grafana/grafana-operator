package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListNotificationTemplate(t *testing.T) {
	t.Run("&NotificationTemplate{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaNotificationTemplate{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "NotificationTemplate does not have a case in Grafana.Status.StatusList")
	})
}

func newNotificationTemplate(name string, editable *bool) *GrafanaNotificationTemplate {
	return &GrafanaNotificationTemplate{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaNotificationTemplate",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaNotificationTemplateSpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "notificationtemplate",
					},
				},
			},
			Name:     name,
			Template: "mock template",
		},
	}
}

var _ = Describe("NotificationTemplate type", func() {
	Context("Ensure NotificationTemplate spec.editable is immutable", func() {
		ctx := context.Background()
		refTrue := true
		refFalse := false

		It("Should block adding editable field when missing", func() {
			notificationtemplate := newNotificationTemplate("missing-editable", nil)
			By("Create new NotificationTemplate without editable")
			Expect(k8sClient.Create(ctx, notificationtemplate)).To(Succeed())

			By("Adding a editable")
			notificationtemplate.Spec.Editable = &refTrue
			Expect(k8sClient.Update(ctx, notificationtemplate)).To(HaveOccurred())
		})

		It("Should block removing editable field when set", func() {
			notificationtemplate := newNotificationTemplate("existing-editable", &refTrue)
			By("Creating NotificationTemplate with existing editable")
			Expect(k8sClient.Create(ctx, notificationtemplate)).To(Succeed())

			By("And setting editable to ''")
			notificationtemplate.Spec.Editable = nil
			Expect(k8sClient.Update(ctx, notificationtemplate)).To(HaveOccurred())
		})

		It("Should block changing value of editable", func() {
			notificationtemplate := newNotificationTemplate("removing-editable", &refTrue)
			By("Create new NotificationTemplate with existing editable")
			Expect(k8sClient.Create(ctx, notificationtemplate)).To(Succeed())

			By("Changing the existing editable")
			notificationtemplate.Spec.Editable = &refFalse
			Expect(k8sClient.Update(ctx, notificationtemplate)).To(HaveOccurred())
		})
	})
})
