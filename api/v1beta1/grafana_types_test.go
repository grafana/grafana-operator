package v1beta1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Grafana status NamespacedResourceList all CRs works", func() {
	ctx := context.Background()
	Context("Update entry in NamespacedResourceList", Ordered, func() {
		meta := func() metav1.ObjectMeta {
			return metav1.ObjectMeta{
				Name:      "status-list-item",
				Namespace: "default",
			}
		}

		alertRuleGroup := &GrafanaAlertRuleGroup{ObjectMeta: meta()}
		contactPoint := &GrafanaContactPoint{
			ObjectMeta: meta(),
			Spec: GrafanaContactPointSpec{
				CustomUID: "contact-one",
			},
		}
		dashboard := &GrafanaDashboard{
			ObjectMeta: meta(),
			Spec: GrafanaDashboardSpec{
				GrafanaContentSpec: GrafanaContentSpec{
					CustomUID: "db-unique-identifier",
				},
			},
		}
		datasource := &GrafanaDatasource{
			ObjectMeta: meta(),
			Spec: GrafanaDatasourceSpec{
				CustomUID: "ds-one-unique-identifier",
			},
		}
		folder := &GrafanaFolder{
			ObjectMeta: meta(),
			Spec: GrafanaFolderSpec{
				CustomUID: "folder-unique-identifier",
			},
		}
		libraryPanel := &GrafanaLibraryPanel{
			ObjectMeta: meta(),
			Spec: GrafanaLibraryPanelSpec{
				GrafanaContentSpec: GrafanaContentSpec{
					CustomUID: "lp-one-unique-identifier",
				},
			},
		}
		muteTiming := &GrafanaMuteTiming{ObjectMeta: meta()}
		notificationTemplate := &GrafanaNotificationTemplate{ObjectMeta: meta()}

		crList := []client.Object{alertRuleGroup, contactPoint, dashboard, datasource, folder, libraryPanel, muteTiming, notificationTemplate}

		crGrafana := &Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "status-patch-all-crs",
				Namespace: "default",
				Labels: map[string]string{
					"test": "status-patch",
				},
			},
			Spec: GrafanaSpec{},
		}

		BeforeAll(func() {
			Expect(k8sClient.Create(ctx, crGrafana)).To(Succeed())
			crGrafana.Status.Stage = OperatorStageComplete
			crGrafana.Status.StageStatus = OperatorStageResultSuccess
			Expect(k8sClient.Status().Update(ctx, crGrafana)).To(Succeed())
		})

		It("Adds item to status of Grafana", func() {
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, alertRuleGroup, alertRuleGroup.NamespacedResource())).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, contactPoint, contactPoint.NamespacedResource())).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, dashboard, dashboard.NamespacedResource(dashboard.Spec.CustomUID))).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, datasource, datasource.NamespacedResource())).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, folder, folder.NamespacedResource(folder.Spec.CustomUID))).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, libraryPanel, libraryPanel.NamespacedResource(libraryPanel.Spec.CustomUID))).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, muteTiming, muteTiming.NamespacedResource())).Should(Succeed())
			Expect(crGrafana.AddNamespacedResource(ctx, k8sClient, notificationTemplate, notificationTemplate.NamespacedResource())).Should(Succeed())

			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: crGrafana.Namespace,
				Name:      crGrafana.Name,
			}, im)).To(Succeed())

			for _, cr := range crList {
				list, _, err := im.Status.StatusList(cr)
				Expect(err).ToNot(HaveOccurred())
				Expect(list).ToNot(BeNil())
				Expect(*list).ToNot(BeEmpty())
				Expect(*list).To(HaveLen(1))

				idx := im.Status.Datasources.IndexOf(cr.GetNamespace(), cr.GetName())
				Expect(idx).To(Equal(0))
			}

			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, alertRuleGroup)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, contactPoint)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, dashboard)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, datasource)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, folder)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, libraryPanel)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, muteTiming)).Should(Succeed())
			Expect(crGrafana.RemoveNamespacedResource(ctx, k8sClient, notificationTemplate)).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: crGrafana.Namespace,
				Name:      crGrafana.Name,
			}, result)).To(Succeed())

			for _, cr := range crList {
				list, _, err := result.Status.StatusList(cr)
				Expect(err).ToNot(HaveOccurred())
				Expect(*list).To(BeEmpty())

				idx := result.Status.Datasources.IndexOf(cr.GetNamespace(), cr.GetName())
				Expect(idx).To(Equal(-1))
			}
		})
	})
})

var _ = Describe("Grafana Status NamespacedResourceList CRUD", Ordered, func() {
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
		By("creating Grafana cr and updating the status before testing")
		Expect(k8sClient.Create(ctx, g)).Should(Succeed())

		g.Status.Stage = OperatorStageComplete
		g.Status.StageStatus = OperatorStageResultSuccess
		Expect(k8sClient.Status().Update(ctx, g)).Should(Succeed())
	})
	// Fetch latest status before each Spec
	BeforeEach(func() {
		By("fetching latest Grafana manifest")
		tmpGrafana := &Grafana{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Namespace: g.Namespace,
			Name:      g.Name,
		}, tmpGrafana)).To(Succeed())
		g = tmpGrafana
	})

	Context("Create, Update, Delete entries in a NamespacedResourceList", Ordered, func() {
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

		It("Can add new LibraryPanel entry when list is empty", func() {
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp1, lp1.NamespacedResource(lp1.Spec.CustomUID))).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.LibraryPanels).ToNot(BeEmpty())
			Expect(result.Status.LibraryPanels).To(HaveLen(1))
			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp1.NamespacedResource(lp1.Spec.CustomUID)))
		})

		It("Adds an additional LibraryPanel entries when list is not empty", func() {
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp2, lp2.NamespacedResource(lp2.Spec.CustomUID))).Should(Succeed())
			Expect(g.AddNamespacedResource(ctx, k8sClient, lp3, lp3.NamespacedResource(lp3.Spec.CustomUID))).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.LibraryPanels).ToNot(BeEmpty())
			Expect(result.Status.LibraryPanels).To(HaveLen(3))

			idx := result.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			Expect(idx).To(Equal(1))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp2.NamespacedResource(lp2.Spec.CustomUID)))

			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(2))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp3.NamespacedResource(lp3.Spec.CustomUID)))
		})

		It("Removes LibraryPanel from the middle of a list with multiple entries", func() {
			// Verify state before removal
			idx := g.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(2))
			Expect(g.Status.LibraryPanels[idx]).To(Equal(lp3.NamespacedResource(lp3.Spec.CustomUID)))

			// Remove middle entry
			Expect(g.RemoveNamespacedResource(ctx, k8sClient, lp2)).To(Succeed())

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

			// Was removed and should not be found
			idx = result.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			Expect(idx).To(Equal(-1))

			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(1))
			Expect(result.Status.LibraryPanels[idx]).To(Equal(lp3.NamespacedResource(lp3.Spec.CustomUID)))
		})

		It("Removes LibraryPanels from list", func() {
			// Only lp1 and lp3 remains in the Status at this time
			Expect(g.Status.LibraryPanels).ToNot(BeEmpty())
			Expect(g.Status.LibraryPanels).To(HaveLen(2))

			Expect(g.RemoveNamespacedResource(ctx, k8sClient, lp1)).Should(Succeed())
			Expect(g.RemoveNamespacedResource(ctx, k8sClient, lp3)).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			Expect(idx).To(Equal(-1))
			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			Expect(idx).To(Equal(-1))
			Expect(g.Status.LibraryPanels).To(BeEmpty())
		})

		It("Removes LibraryPanels from undefined list", func() {
			Expect(g.Status.LibraryPanels).To(BeEmpty())
			Expect(g.RemoveNamespacedResource(ctx, k8sClient, lp1)).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			Expect(idx).To(Equal(-1))
			Expect(g.Status.LibraryPanels).To(BeEmpty())
		})
	})

	Context("Update entry in NamespacedResourceList", Ordered, func() {
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

		It("Does not add new Datasource when entry exists", func() {
			Expect(g.AddNamespacedResource(ctx, k8sClient, ds1, ds1.NamespacedResource())).Should(Succeed())

			// Intermediate
			im := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)).To(Succeed())
			Expect(im.AddNamespacedResource(ctx, k8sClient, ds1, ds1.NamespacedResource())).Should(Succeed())

			result := &Grafana{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)).To(Succeed())

			Expect(result.Status.Datasources).ToNot(BeEmpty())
			Expect(result.Status.Datasources).To(HaveLen(1))
			idx := result.Status.Datasources.IndexOf(ds1.Namespace, ds1.Name)
			Expect(idx).To(Equal(0))
			Expect(result.Status.Datasources[idx]).To(Equal(ds1.NamespacedResource()))
		})

		It("Updates existing Datasource on uid changed", func() {
			Expect(g.AddNamespacedResource(ctx, k8sClient, ds2, ds2.NamespacedResource())).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())

			Expect(g.AddNamespacedResource(ctx, k8sClient, ds3, ds3.NamespacedResource())).Should(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)).To(Succeed())

			Expect(g.Status.Datasources).ToNot(BeEmpty())
			Expect(g.Status.Datasources).To(HaveLen(3))

			// Update entry at the middle of the list
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
	})
})
