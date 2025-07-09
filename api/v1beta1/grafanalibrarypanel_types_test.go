package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaLibraryPanel",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaLibraryPanelSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
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
	Context("Ensure LibraryPanel spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newLibraryPanel("missing-uid", "")
			By("Create new LibraryPanel without uid")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Adding a uid")
			dash.Spec.CustomUID = "new-library-panel-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			dash := newLibraryPanel("existing-uid", "existing-uid")
			By("Creating LibraryPanel with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("And setting UID to ''")
			dash.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			dash := newLibraryPanel("removing-uid", "existing-uid")
			By("Create new LibraryPanel with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Changing the existing UID")
			dash.Spec.CustomUID = "new-library-panel-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})
	})
})
