package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDashboard(name string, uid string) *GrafanaDashboard {
	return &GrafanaDashboard{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDashboard",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDashboardSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			GrafanaContentSpec: GrafanaContentSpec{
				CustomUID: uid,
				Json:      "",
			},
		},
	}
}

var _ = Describe("Dashboard type", func() {
	Context("Ensure Dashboard spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newDashboard("missing-uid", "")
			By("Create new Dashboard without uid")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Adding a uid")
			dash.Spec.CustomUID = "new-dash-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			dash := newDashboard("existing-uid", "existing-uid")
			By("Creating Dashboard with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("And setting UID to ''")
			dash.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			dash := newDashboard("removing-uid", "existing-uid")
			By("Create new Dashboard with existing UID")
			Expect(k8sClient.Create(ctx, dash)).To(Succeed())

			By("Changing the existing UID")
			dash.Spec.CustomUID = "new-dash-uid"
			Expect(k8sClient.Update(ctx, dash)).To(HaveOccurred())
		})
	})
})
