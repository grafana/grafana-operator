package v1beta1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListContactPoint(t *testing.T) {
	t.Run("&ContactPoint{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaContactPoint{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "ContactPoint does not have a case in Grafana.Status.StatusList")
	})
}

func newContactPoint(name string) *GrafanaContactPoint {
	return &GrafanaContactPoint{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaContactPointSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{},
			},
		},
	}
}

var _ = Describe("ContactPoint type", func() {
	const webhookType = "webhook"

	Context("Ensure ContactPoint spec.name is immutable", func() {
		t := GinkgoT()

		It("Should block adding name field when missing", func() {
			contactpoint := newContactPoint("adding-name")
			contactpoint.Spec.Type = webhookType

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)

			contactpoint.Spec.Name = "update-name"
			err = k8sClient.Update(t.Context(), contactpoint)
			require.Error(t, err)
		})

		It("Should block removing name field when set", func() {
			contactpoint := newContactPoint("removing-name")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Name = "initial-name"

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)

			contactpoint.Spec.Name = ""
			err = k8sClient.Update(t.Context(), contactpoint)
			require.Error(t, err)
		})

		It("Should block changing value of name", func() {
			contactpoint := newContactPoint("updating-name")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Name = "initial-name"

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)

			contactpoint.Spec.Name = "new-name"
			err = k8sClient.Update(t.Context(), contactpoint)
			require.Error(t, err)
		})
	})

	Context("Ensure ContactPoint spec.editable is immutable", func() {
		t := GinkgoT()

		It("Should block enabling editable", func() {
			contactpoint := newContactPoint("updating-editable")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Editable = false

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)

			contactpoint.Spec.Editable = true
			err = k8sClient.Update(t.Context(), contactpoint)
			require.Error(t, err)
		})

		It("Should block disabling editable", func() {
			contactpoint := newContactPoint("removing-editable")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Editable = true

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)

			contactpoint.Spec.Editable = false
			err = k8sClient.Update(t.Context(), contactpoint)
			require.Error(t, err)
		})
	})

	Context("Ensure ContactPoint receivers is correctly handled", func() {
		settings := apiextensionsv1.JSON{Raw: []byte("{}")}

		It("Succeeds when no receiver is found", func() {
			t := GinkgoT()

			contactpoint := newContactPoint("missing-receivers")
			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)
		})

		It("Successfully created with top level receiver", func() {
			t := GinkgoT()

			contactpoint := newContactPoint("top-level-receiver")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Settings = &settings

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)
		})

		It("Successfully created with list of receivers", func() {
			t := GinkgoT()

			contactpoint := newContactPoint("list-of-receivers")
			contactpoint.Spec.Receivers = []ContactPointReceiver{{
				Type:     "webhook",
				Settings: &settings,
			}}
			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)
		})

		It("Successfully created with both top level and list of receivers", func() {
			t := GinkgoT()

			contactpoint := newContactPoint("both-top-level-and-list-of-receiver")
			contactpoint.Spec.Type = webhookType
			contactpoint.Spec.Settings = &settings
			contactpoint.Spec.Receivers = []ContactPointReceiver{{
				Type:     "webhook",
				Settings: &settings,
			}}

			err := k8sClient.Create(t.Context(), contactpoint)
			require.NoError(t, err)
		})
	})
})
