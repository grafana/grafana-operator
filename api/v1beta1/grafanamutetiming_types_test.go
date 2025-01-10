package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMuteTiming(name string, editable *bool) *GrafanaMuteTiming {
	return &GrafanaMuteTiming{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaMuteTiming",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaMuteTimingSpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "mutetiming",
					},
				},
			},
			Name: name,
			TimeIntervals: []*TimeInterval{
				{
					DaysOfMonth: []string{"1"},
					Location:    "Asia/Shanghai",
					Months:      []string{"1"},
					Times: []*TimeRange{
						{
							StartTime: "00:00",
							EndTime:   "02:00",
						},
					},
					Weekdays: []string{"1"},
					Years:    []string{"2025"},
				},
			},
		},
	}
}

var _ = Describe("MuteTiming type", func() {
	Context("Ensure MuteTiming spec.editable is immutable", func() {
		ctx := context.Background()
		refTrue := true
		refFalse := false

		It("Should block adding editable field when missing", func() {
			mutetiming := newMuteTiming("missing-editable", nil)
			By("Create new MuteTiming without editable")
			Expect(k8sClient.Create(ctx, mutetiming)).To(Succeed())

			By("Adding a editable")
			mutetiming.Spec.Editable = &refTrue
			Expect(k8sClient.Update(ctx, mutetiming)).To(HaveOccurred())
		})

		It("Should block removing editable field when set", func() {
			mutetiming := newMuteTiming("existing-editable", &refTrue)
			By("Creating MuteTiming with existing editable")
			Expect(k8sClient.Create(ctx, mutetiming)).To(Succeed())

			By("And setting editable to ''")
			mutetiming.Spec.Editable = nil
			Expect(k8sClient.Update(ctx, mutetiming)).To(HaveOccurred())
		})

		It("Should block changing value of editable", func() {
			mutetiming := newMuteTiming("removing-editable", &refTrue)
			By("Create new MuteTiming with existing editable")
			Expect(k8sClient.Create(ctx, mutetiming)).To(Succeed())

			By("Changing the existing editable")
			mutetiming.Spec.Editable = &refFalse
			Expect(k8sClient.Update(ctx, mutetiming)).To(HaveOccurred())
		})
	})
})
