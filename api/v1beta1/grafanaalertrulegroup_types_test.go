package v1beta1

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
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
					ExecErrState: "KeepLast",
					NoDataState:  &noDataState,
					For:          &metav1.Duration{Duration: 60 * time.Second},
					Data:         []*AlertQuery{},
				},
			},
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
	Context("Ensure AlertRuleGroup spec.folderRef and spec.folderUID are immutable", func() {
		ctx := context.Background()
		refTrue := true

		It("Should block changing value of folderRef", func() {
			arg := newAlertRuleGroup("changing-folder-ref", &refTrue)
			By("Creating new AlertRuleGroup with existing folderRef")
			Expect(k8sClient.Create(ctx, arg)).To(Succeed())

			By("Changing folderRef")
			arg.Spec.FolderRef = "newFolder"
			Expect(k8sClient.Update(ctx, arg)).To(HaveOccurred())
		})

		It("Should block changing value of folderUID", func() {
			arg := newAlertRuleGroup("changing-folder-uid", &refTrue)
			arg.Spec.FolderRef = ""

			arg.Spec.FolderUID = "originalUID"
			By("Creating new AlertRuleGroup with existing folderUID")
			Expect(k8sClient.Create(ctx, arg)).To(Succeed())

			By("Changing folderUID")
			arg.Spec.FolderUID = "newUID"
			Expect(k8sClient.Update(ctx, arg)).To(HaveOccurred())
		})

		It("At least one of spec.folderRef or spec.folderUID is defined", func() {
			arg := newAlertRuleGroup("missing-folder", &refTrue)
			arg.Spec.FolderRef = ""
			arg.Spec.FolderUID = ""
			By("Creating new AlertRuleGroup with neither folderUID")
			Expect(k8sClient.Create(ctx, arg)).To(HaveOccurred())
		})

		It("Only one of spec.folderRef or spec.folderUID is defined", func() {
			arg := newAlertRuleGroup("mutually-exclusive-folder-reference", &refTrue)
			arg.Spec.FolderUID = "DummyUID"
			By("Creating new AlertRuleGroup with neither folderUID")
			Expect(k8sClient.Create(ctx, arg)).To(HaveOccurred())
		})
	})
})
