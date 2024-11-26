package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newAlertRuleGroup(name string, editable *bool) *GrafanaAlertRuleGroup {
	return &GrafanaAlertRuleGroup{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaAlertRuleGroup",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaAlertRuleGroupSpec{
			Name:      name,
			Editable:  editable,
			FolderRef: "DummyFolderRef",
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "alertrulegroup",
					},
				},
			},
			Rules: []AlertRule{},
		},
	}
}

var _ = Describe("AlertRuleGroup type", func() {
	Context("Ensure AlertRuleGroup spec.editable is immutable", func() {
		ctx := context.Background()
		refTrue := true
		refFalse := false

		It("Should block adding editable field when missing", func() {
			arg := newAlertRuleGroup("missing-editable", nil)
			By("Create new AlertRuleGroup without editable")
			Expect(k8sClient.Create(ctx, arg)).To(Succeed())

			By("Adding a editable")
			arg.Spec.Editable = &refTrue
			Expect(k8sClient.Update(ctx, arg)).To(HaveOccurred())
		})

		It("Should block removing editable field when set", func() {
			arg := newAlertRuleGroup("existing-editable", &refTrue)
			By("Creating AlertRuleGroup with existing editable")
			Expect(k8sClient.Create(ctx, arg)).To(Succeed())

			By("And setting editable to ''")
			arg.Spec.Editable = nil
			Expect(k8sClient.Update(ctx, arg)).To(HaveOccurred())
		})

		It("Should block changing value of editable", func() {
			arg := newAlertRuleGroup("removing-editable", &refTrue)
			By("Create new AlertRuleGroup with existing editable")
			Expect(k8sClient.Create(ctx, arg)).To(Succeed())

			By("Changing the existing editable")
			arg.Spec.Editable = &refFalse
			Expect(k8sClient.Update(ctx, arg)).To(HaveOccurred())
		})
	})
})
