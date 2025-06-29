package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Grafana Status", Ordered, func() {
	// Prep
	ctx := context.Background()
	g := &Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "status-patch",
			Namespace: "default",
			Labels: map[string]string{
				"test": "status-patch",
			},
		},
		Spec: GrafanaSpec{},
	}

	BeforeAll(func() {
		By("creating Grafana cr and updating the status")
		Expect(k8sClient.Create(ctx, g)).Should(Succeed())

		g.Status.Stage = OperatorStageComplete
		g.Status.StageStatus = OperatorStageResultSuccess
		Expect(k8sClient.Status().Update(ctx, g)).Should(Succeed())
	})

	Context("NamespacedResourceList patching", func() {
		// Cases
		It("Can add new AlertRuleGroup entry when list is empty", func() {
			arg := &GrafanaAlertRuleGroup{ObjectMeta: metav1.ObjectMeta{
				Name:      "arg-on-status",
				Namespace: "default",
			}}
			Expect(g.AddNamespacedResource(ctx, k8sClient, arg, arg.NamespacedResource())).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.AlertRuleGroups).ToNot(BeEmpty())
			Expect(result.Status.AlertRuleGroups).To(HaveLen(1))
			idx := result.Status.AlertRuleGroups.IndexOf(arg.Namespace, arg.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.AlertRuleGroups[idx]).To(Equal(arg.NamespacedResource()))
		})

		It("Adds an additional ContactPoint entry when list is not empty", func() {
			c1 := &GrafanaContactPoint{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "contact-one-on-status",
					Namespace: "default",
				},
				Spec: GrafanaContactPointSpec{
					CustomUID: "contact-one",
				},
			}
			c2 := &GrafanaContactPoint{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "contact-two-on-status",
					Namespace: "default",
				},
				Spec: GrafanaContactPointSpec{
					CustomUID: "contact-two",
				},
			}
			Expect(g.AddNamespacedResource(ctx, k8sClient, c1, c1.NamespacedResource())).Should(Succeed())

			// Intermediate
			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)).To(Succeed())
			Expect(im.AddNamespacedResource(ctx, k8sClient, c2, c2.NamespacedResource())).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.ContactPoints).ToNot(BeEmpty())
			Expect(result.Status.ContactPoints).To(HaveLen(2))
			idx := result.Status.ContactPoints.IndexOf(c1.Namespace, c1.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.ContactPoints[idx]).To(Equal(c1.NamespacedResource()))

			idx = result.Status.ContactPoints.IndexOf(c2.Namespace, c2.Name)
			Expect(idx).To(Equal(1))
			Expect(result.Status.ContactPoints[idx]).To(Equal(c2.NamespacedResource()))
		})

		It("Skips patch when Dashboard entry already exists", func() {
			d := &GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dashboard-on-status",
					Namespace: "default",
				},
				Spec: GrafanaDashboardSpec{
					GrafanaContentSpec: GrafanaContentSpec{
						CustomUID: "dash-unique-identifier",
					},
				},
			}
			// Add resource twice
			Expect(g.AddNamespacedResource(ctx, k8sClient, d, d.NamespacedResource(d.Spec.CustomUID))).Should(Succeed())

			// Intermediate
			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)).To(Succeed())
			Expect(im.AddNamespacedResource(ctx, k8sClient, d, d.NamespacedResource(d.Spec.CustomUID))).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.Dashboards).ToNot(BeEmpty())
			Expect(result.Status.Dashboards).To(HaveLen(1))
			idx := result.Status.Dashboards.IndexOf(d.Namespace, d.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.Dashboards[idx]).To(Equal(d.NamespacedResource(d.Spec.CustomUID)))
		})

		It("Updates existing Datasource on uid changed", func() {
			ds1 := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "datasource-one-on-status",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					CustomUID: "ds-one-unique-identifier",
				},
			}
			ds2 := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "datasource-two-on-status",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					CustomUID: "ds-two-unique-identifier",
				},
			}
			ds3 := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "datasource-three-on-status",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					CustomUID: "ds-three-unique-identifier",
				},
			}

			Expect(g.AddNamespacedResource(ctx, k8sClient, ds1, ds1.NamespacedResource())).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())
			Expect(g.Status.Datasources).ToNot(BeEmpty())
			Expect(g.Status.Datasources).To(HaveLen(1))

			Expect(g.AddNamespacedResource(ctx, k8sClient, ds2, ds2.NamespacedResource())).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())
			Expect(g.Status.Datasources).ToNot(BeEmpty())
			Expect(g.Status.Datasources).To(HaveLen(2))

			Expect(g.AddNamespacedResource(ctx, k8sClient, ds3, ds3.NamespacedResource())).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())
			Expect(g.Status.Datasources).ToNot(BeEmpty())
			Expect(g.Status.Datasources).To(HaveLen(3))

			// Update existing Entry
			ds2.Spec.CustomUID = "ds-2-unique-identifier"
			Expect(g.AddNamespacedResource(ctx, k8sClient, ds2, ds2.NamespacedResource())).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.Datasources).ToNot(BeEmpty())
			Expect(result.Status.Datasources).To(HaveLen(3))
			idx := result.Status.Datasources.IndexOf(ds1.Namespace, ds1.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.Datasources[idx]).To(Equal(ds1.NamespacedResource()))
			idx = result.Status.Datasources.IndexOf(ds2.Namespace, ds2.Name)
			Expect(idx).To(Equal(1))
			Expect(result.Status.Datasources[idx]).To(Equal(ds2.NamespacedResource()))
			idx = result.Status.Datasources.IndexOf(ds3.Namespace, ds3.Name)
			Expect(idx).To(Equal(2))
			Expect(result.Status.Datasources[idx]).To(Equal(ds3.NamespacedResource()))
		})

		It("Removes Folder from undefined list", func() {
			f := &GrafanaFolder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "folder-on-status",
					Namespace: "default",
				},
				Spec: GrafanaFolderSpec{
					CustomUID: "f-unique-identifier",
				},
			}
			Expect(g.Status.Folders).To(BeEmpty())
			Expect(g.RemoveNamespacedResource(ctx, k8sClient, f)).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(g.Status.Folders).To(BeEmpty())
			idx := result.Status.Folders.IndexOf(f.Namespace, f.Name)
			Expect(idx).To(Equal(-1))
		})

		It("Remove LibraryPanel from list with multiple entries", func() {
			lp1 := &GrafanaLibraryPanel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "librarypanel-one-on-status",
					Namespace: "default",
				},
				Spec: GrafanaLibraryPanelSpec{
					GrafanaContentSpec: GrafanaContentSpec{
						CustomUID: "lp-one-unique-identifier",
					},
				},
			}
			lp2 := &GrafanaLibraryPanel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "librarypanel-two-on-status",
					Namespace: "default",
				},
				Spec: GrafanaLibraryPanelSpec{
					GrafanaContentSpec: GrafanaContentSpec{
						CustomUID: "lp-two-unique-identifier",
					},
				},
			}
			lp3 := &GrafanaLibraryPanel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "librarypanel-three-on-status",
					Namespace: "default",
				},
				Spec: GrafanaLibraryPanelSpec{
					GrafanaContentSpec: GrafanaContentSpec{
						CustomUID: "lp-three-unique-identifier",
					},
				},
			}
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp1, lp1.NamespacedResource(lp1.Spec.CustomUID))).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp2, lp2.NamespacedResource(lp2.Spec.CustomUID))).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp3, lp3.NamespacedResource(lp3.Spec.CustomUID))).Should(Succeed())

			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)).To(Succeed())

			// Verify state before removal
			Expect(g.Status.LibraryPanels).ToNot(BeEmpty())
			Expect(g.Status.LibraryPanels).To(HaveLen(3))
			idx := im.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			Expect(idx).To(Equal(0))
			Expect(im.Status.LibraryPanels[idx]).To(Equal(lp1.NamespacedResource(lp1.Spec.CustomUID)))
			idx = im.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			Expect(idx).To(Equal(1))
			Expect(im.Status.LibraryPanels[idx]).To(Equal(lp2.NamespacedResource(lp2.Spec.CustomUID)))
			idx = im.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(2))
			Expect(im.Status.LibraryPanels[idx]).To(Equal(lp3.NamespacedResource(lp3.Spec.CustomUID)))

			// Remove center item
			Expect(im.RemoveNamespacedResource(ctx, k8sClient, lp2)).To(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.LibraryPanels).ToNot(BeEmpty())
			Expect(result.Status.LibraryPanels).To(HaveLen(2))
			idx = result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp1.NamespacedResource(lp1.Spec.CustomUID)))
			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(1))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp3.NamespacedResource(lp3.Spec.CustomUID)))
			// Was removed and should not be found
			idx = result.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			Expect(idx).To(Equal(-1))
		})

		It("Remove Last MuteTiming entry from list", func() {
			mt := &GrafanaMuteTiming{ObjectMeta: metav1.ObjectMeta{
				Name:      "mutetiming-on-status",
				Namespace: "default",
			}}
			// Add resource twice
			Expect(g.AddNamespacedResource(ctx, k8sClient, mt, mt.NamespacedResource())).Should(Succeed())

			// Intermediate
			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)).To(Succeed())
			Expect(im.RemoveNamespacedResource(ctx, k8sClient, mt)).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.MuteTimings).To(BeEmpty())
			idx := result.Status.MuteTimings.IndexOf(mt.Namespace, mt.Name)
			Expect(idx).To(Equal(-1))
		})
	})
})
