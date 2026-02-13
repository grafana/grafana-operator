package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListDashboard(t *testing.T) {
	t.Run("&Dashboard{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaDashboard{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "Dashboard does not have a case in Grafana.Status.StatusList")
	})
}

func newDashboard(name, uid string) *GrafanaDashboard {
	return &GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDashboard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDashboardSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			GrafanaContentSpec: GrafanaContentSpec{
				CustomUID: uid,
				JSON:      "",
			},
		},
	}
}

var _ = Describe("Dashboard type", func() {
	Context("Ensure Dashboard spec.uid is immutable", func() {
		t := GinkgoT()

		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newDashboard("missing-uid", "")

			By("Create new Dashboard without uid")

			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("Adding a uid")

			dash.Spec.CustomUID = "new-dash-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block removing uid field when set", func() {
			dash := newDashboard("existing-uid", "existing-uid")

			By("Creating Dashboard with existing UID")

			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("And setting UID to ''")

			dash.Spec.CustomUID = ""
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block changing value of uid", func() {
			dash := newDashboard("removing-uid", "existing-uid")

			By("Create new Dashboard with existing UID")

			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("Changing the existing UID")

			dash.Spec.CustomUID = "new-dash-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})
	})
})
