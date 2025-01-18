package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDatasource(name string, uid string) *GrafanaDatasource {
	return &GrafanaDatasource{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDatasourceSpec{
			CustomUID: uid,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			Datasource: &GrafanaDatasourceInternal{
				Name:   "testdata",
				Type:   "grafana-testdata-datasource",
				Access: "proxy",
			},
		},
	}
}

var _ = Describe("Datasource type", func() {
	Context("Ensure Datasource spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			ds := newDatasource("missing-uid", "")
			By("Create new Datasource without uid")
			Expect(k8sClient.Create(ctx, ds)).To(Succeed())

			By("Adding a uid")
			ds.Spec.CustomUID = "new-ds-uid"
			Expect(k8sClient.Update(ctx, ds)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			ds := newDatasource("existing-uid", "existing-uid")
			By("Creating Datasource with existing UID")
			Expect(k8sClient.Create(ctx, ds)).To(Succeed())

			By("And setting UID to ''")
			ds.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, ds)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			ds := newDatasource("removing-uid", "existing-uid")
			By("Create new Datasource with existing UID")
			Expect(k8sClient.Create(ctx, ds)).To(Succeed())

			By("Changing the existing UID")
			ds.Spec.CustomUID = "new-ds-uid"
			Expect(k8sClient.Update(ctx, ds)).To(HaveOccurred())
		})
	})
})

var _ = Describe("Fail on field behavior changes", func() {
	emptyDatasource := &GrafanaDatasource{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-nil-datasource",
			Namespace: "default",
		},
		Spec: GrafanaDatasourceSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{},
			},
			Datasource: nil,
		},
	}

	ctx := context.Background()
	It("Fails creating GrafanaDatasource with undefined spec.datasource", func() {
		Expect(k8sClient.Create(ctx, emptyDatasource)).To(HaveOccurred())
	})
})
