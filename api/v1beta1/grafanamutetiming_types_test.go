package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaMuteTiming",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaMuteTimingSpec{
			Editable: editable,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
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
	t := GinkgoT()

	Context("Ensure MuteTiming spec.editable is immutable", func() {
		ctx := context.Background()

		It("Should block changing value of editable", func() {
			mutetiming := newMuteTiming("removing-editable", true)

			By("Create new MuteTiming with existing editable")

			err := cl.Create(ctx, mutetiming)
			require.NoError(t, err)

			By("Changing the existing editable")

			mutetiming.Spec.Editable = false
			err = cl.Update(ctx, mutetiming)
			require.Error(t, err)
		})
	})
})
