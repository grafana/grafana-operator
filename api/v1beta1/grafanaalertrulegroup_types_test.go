package v1beta1

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListAlertRuleGroup(t *testing.T) {
	t.Run("&AlertRuleGroup{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaAlertRuleGroup{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "AlertRuleGroup does not have a case in Grafana.Status.StatusList")
	})
}

func newAlertRuleGroup(name string, editable *bool) *GrafanaAlertRuleGroup {
	noDataState := "NoData"

	forDurationString := "60s"

	return &GrafanaAlertRuleGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaAlertRuleGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaAlertRuleGroupSpec{
			Name:      name,
			Editable:  editable,
			FolderRef: "DummyFolderRef",
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "alertrulegroup",
					},
				},
			},
			Rules: []AlertRule{
				{
					Title:        "TestRule",
					UID:          "akdj-wonvo",
					ExecErrState: "KeepLast",
					NoDataState:  &noDataState,
					For:          &forDurationString,
					Data:         []*AlertQuery{},
				},
			},
		},
	}
}

var _ = Describe("AlertRuleGroup type", func() {
	t := GinkgoT()
	Context("Ensure AlertRuleGroup spec.editable is immutable", func() {
		ctx := context.Background()
		refTrue := ptr.To(true)
		refFalse := ptr.To(false)

		It("Should block adding editable field when missing", func() {
			arg := newAlertRuleGroup("missing-editable", nil)
			By("Create new AlertRuleGroup without editable")
			err := cl.Create(ctx, arg)
			require.NoError(t, err)

			By("Adding a editable")
			arg.Spec.Editable = refTrue
			err = cl.Update(ctx, arg)
			require.Error(t, err)
		})

		It("Should block removing editable field when set", func() {
			arg := newAlertRuleGroup("existing-editable", refTrue)
			By("Creating AlertRuleGroup with existing editable")
			err := cl.Create(ctx, arg)
			require.NoError(t, err)

			By("And setting editable to ''")
			arg.Spec.Editable = nil
			err = cl.Update(ctx, arg)
			require.Error(t, err)
		})

		It("Should block changing value of editable", func() {
			arg := newAlertRuleGroup("removing-editable", refTrue)
			By("Create new AlertRuleGroup with existing editable")
			err := cl.Create(ctx, arg)
			require.NoError(t, err)

			By("Changing the existing editable")
			arg.Spec.Editable = refFalse
			err = cl.Update(ctx, arg)
			require.Error(t, err)
		})
	})
	Context("Ensure AlertRuleGroup spec.folderRef and spec.folderUID are immutable", func() {
		ctx := context.Background()
		refTrue := ptr.To(true)

		It("Should block changing value of folderRef", func() {
			arg := newAlertRuleGroup("changing-folder-ref", refTrue)
			By("Creating new AlertRuleGroup with existing folderRef")
			err := cl.Create(ctx, arg)
			require.NoError(t, err)

			By("Changing folderRef")
			arg.Spec.FolderRef = "newFolder"
			err = cl.Update(ctx, arg)
			require.Error(t, err)
		})

		It("Should block changing value of folderUID", func() {
			arg := newAlertRuleGroup("changing-folder-uid", refTrue)
			arg.Spec.FolderRef = ""

			arg.Spec.FolderUID = "originalUID"
			By("Creating new AlertRuleGroup with existing folderUID")
			err := cl.Create(ctx, arg)
			require.NoError(t, err)

			By("Changing folderUID")
			arg.Spec.FolderUID = "newUID"
			err = cl.Update(ctx, arg)
			require.Error(t, err)
		})

		It("At least one of spec.folderRef or spec.folderUID is defined", func() {
			arg := newAlertRuleGroup("missing-folder", refTrue)
			arg.Spec.FolderRef = ""
			arg.Spec.FolderUID = ""
			By("Creating new AlertRuleGroup with neither folderUID")
			err := cl.Create(ctx, arg)
			require.Error(t, err)
		})

		It("Only one of spec.folderRef or spec.folderUID is defined", func() {
			arg := newAlertRuleGroup("mutually-exclusive-folder-reference", refTrue)
			arg.Spec.FolderUID = "DummyUID"
			By("Creating new AlertRuleGroup with neither folderUID")
			err := cl.Create(ctx, arg)
			require.Error(t, err)
		})
	})
})

func TestAlertRuleGroupDuration(t *testing.T) {
	t.Run("Should accept day duration in For field", func(t *testing.T) {
		dayDuration := "1d"
		arg := newAlertRuleGroup("day-duration-test", nil)
		arg.Spec.Rules[0].For = &dayDuration

		assert.Equal(t, "1d", *arg.Spec.Rules[0].For)
	})

	t.Run("Should accept week duration in For field", func(t *testing.T) {
		weekDuration := "1w"
		arg := newAlertRuleGroup("week-duration-test", nil)
		arg.Spec.Rules[0].For = &weekDuration

		assert.Equal(t, "1w", *arg.Spec.Rules[0].For)
	})

	t.Run("Should accept fractional day duration in For field", func(t *testing.T) {
		halfDayDuration := "12h"
		arg := newAlertRuleGroup("half-day-duration-test", nil)
		arg.Spec.Rules[0].For = &halfDayDuration

		assert.Equal(t, "12h", *arg.Spec.Rules[0].For)
	})

	t.Run("Should accept fractional week minutes in For field", func(t *testing.T) {
		halfWeekMinutes := "30m"
		arg := newAlertRuleGroup("half-week-minutes-test", nil)
		arg.Spec.Rules[0].For = &halfWeekMinutes

		assert.Equal(t, "30m", *arg.Spec.Rules[0].For)
	})

	t.Run("Should accept fractional week seconds in For field", func(t *testing.T) {
		halfWeekSeconds := "30s"
		arg := newAlertRuleGroup("half-week-seconds-test", nil)
		arg.Spec.Rules[0].For = &halfWeekSeconds

		assert.Equal(t, "30s", *arg.Spec.Rules[0].For)
	})
}
