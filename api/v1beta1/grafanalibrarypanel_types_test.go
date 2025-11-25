package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListLibraryPanel(t *testing.T) {
	t.Run("&LibraryPanel{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaLibraryPanel{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "LibraryPanel does not have a case in Grafana.Status.StatusList")
	})
}

func newLibraryPanel(name string, uid string) *GrafanaLibraryPanel {
	return &GrafanaLibraryPanel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaLibraryPanel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaLibraryPanelSpec{
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

var _ = Describe("LibraryPanel type", func() {
	t := GinkgoT()

	Context("Ensure LibraryPanel spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newLibraryPanel("missing-uid", "")
			By("Create new LibraryPanel without uid")
			err := k8sClient.Create(ctx, dash)
			require.NoError(t, err)

			By("Adding a uid")
			dash.Spec.CustomUID = "new-library-panel-uid"
			err = k8sClient.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block removing uid field when set", func() {
			dash := newLibraryPanel("existing-uid", "existing-uid")
			By("Creating LibraryPanel with existing UID")
			err := k8sClient.Create(ctx, dash)
			require.NoError(t, err)

			By("And setting UID to ''")
			dash.Spec.CustomUID = ""
			err = k8sClient.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block changing value of uid", func() {
			dash := newLibraryPanel("removing-uid", "existing-uid")
			By("Create new LibraryPanel with existing UID")
			err := k8sClient.Create(ctx, dash)
			require.NoError(t, err)

			By("Changing the existing UID")
			dash.Spec.CustomUID = "new-library-panel-uid"
			err = k8sClient.Update(ctx, dash)
			require.Error(t, err)
		})
	})
})
