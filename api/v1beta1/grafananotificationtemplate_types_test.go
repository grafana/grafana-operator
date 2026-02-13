package v1beta1

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaNotificationTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaNotificationTemplateSpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
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
		t := GinkgoT()

		ctx := context.Background()
		refTrue := ptr.To(true)
		refFalse := ptr.To(false)

		It("Should block adding editable field when missing", func() {
			notificationtemplate := newNotificationTemplate("missing-editable", nil)

			By("Create new NotificationTemplate without editable")

			err := cl.Create(ctx, notificationtemplate)
			require.NoError(t, err)

			By("Adding a editable")

			notificationtemplate.Spec.Editable = refTrue
			err = cl.Update(ctx, notificationtemplate)
			require.Error(t, err)
		})

		It("Should block removing editable field when set", func() {
			notificationtemplate := newNotificationTemplate("existing-editable", refTrue)

			By("Creating NotificationTemplate with existing editable")

			err := cl.Create(ctx, notificationtemplate)
			require.NoError(t, err)

			By("And setting editable to ''")

			notificationtemplate.Spec.Editable = nil
			err = cl.Update(ctx, notificationtemplate)
			require.Error(t, err)
		})

		It("Should block changing value of editable", func() {
			notificationtemplate := newNotificationTemplate("removing-editable", refTrue)

			By("Create new NotificationTemplate with existing editable")

			err := cl.Create(ctx, notificationtemplate)
			require.NoError(t, err)

			By("Changing the existing editable")

			notificationtemplate.Spec.Editable = refFalse
			err = cl.Update(ctx, notificationtemplate)
			require.Error(t, err)
		})
	})
})
