package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListMuteTiming(t *testing.T) {
	t.Run("&MuteTiming{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaMuteTiming{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "MuteTiming does not have a case in Grafana.Status.StatusList")
	})
}

func newMuteTiming(name string, editable bool) *GrafanaMuteTiming {
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

		It("Should block changing value of editable", func() {
			mutetiming := newMuteTiming("removing-editable", true)
			By("Create new MuteTiming with existing editable")
			Expect(k8sClient.Create(ctx, mutetiming)).To(Succeed())

			By("Changing the existing editable")
			mutetiming.Spec.Editable = false
			Expect(k8sClient.Update(ctx, mutetiming)).To(HaveOccurred())
		})
	})
})
