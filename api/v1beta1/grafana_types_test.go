package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Grafana status NamespacedResourceList all CRs works", func() {
	t := GinkgoT()

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

		crList := []struct {
			obj client.Object
			nr  NamespacedResource
		}{
			{
				obj: alertRuleGroup,
				nr:  alertRuleGroup.NamespacedResource(),
			},
			{
				obj: contactPoint,
				nr:  contactPoint.NamespacedResource(),
			},
			{
				obj: dashboard,
				nr:  dashboard.NamespacedResource(dashboard.Spec.CustomUID),
			},
			{
				obj: datasource,
				nr:  datasource.NamespacedResource(),
			},
			{
				obj: folder,
				nr:  folder.NamespacedResource(folder.Spec.CustomUID),
			},
			{
				obj: libraryPanel,
				nr:  libraryPanel.NamespacedResource(libraryPanel.Spec.CustomUID),
			},
			{
				obj: muteTiming,
				nr:  muteTiming.NamespacedResource(),
			},
			{
				obj: notificationTemplate,
				nr:  notificationTemplate.NamespacedResource(),
			},
		}

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
			err := cl.Create(ctx, crGrafana)
			require.NoError(t, err)

			crGrafana.Status.Stage = OperatorStageComplete
			crGrafana.Status.StageStatus = OperatorStageResultSuccess

			err = cl.Status().Update(ctx, crGrafana)
			require.NoError(t, err)
		})

		It("Adds item to status of Grafana", func() {
			for _, cr := range crList {
				err := crGrafana.AddNamespacedResource(ctx, cl, cr.obj, cr.nr)
				require.NoError(t, err)
			}

			im := &Grafana{}
			err := cl.Get(ctx, types.NamespacedName{
				Namespace: crGrafana.Namespace,
				Name:      crGrafana.Name,
			}, im)
			require.NoError(t, err)

			for _, cr := range crList {
				list, _, err := im.Status.StatusList(cr.obj)
				require.NoError(t, err)
				require.NotNil(t, list)
				assert.NotEmpty(t, *list)
				assert.Len(t, *list, 1)

				idx := im.Status.Datasources.IndexOf(cr.obj.GetNamespace(), cr.obj.GetName())
				assert.Equal(t, 0, idx)
			}

			for _, cr := range crList {
				err = crGrafana.RemoveNamespacedResource(ctx, cl, cr.obj)
				require.NoError(t, err)
			}

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: crGrafana.Namespace,
				Name:      crGrafana.Name,
			}, result)
			require.NoError(t, err)

			for _, cr := range crList {
				list, _, err := result.Status.StatusList(cr.obj)
				require.NoError(t, err)
				assert.Empty(t, *list)

				idx := result.Status.Datasources.IndexOf(cr.obj.GetNamespace(), cr.obj.GetName())
				assert.Equal(t, -1, idx)
			}
		})
	})
})

var _ = Describe("Grafana Status NamespacedResourceList CRUD", Ordered, func() {
	t := GinkgoT()

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
		err := cl.Create(ctx, g)
		require.NoError(t, err)

		g.Status.Stage = OperatorStageComplete
		g.Status.StageStatus = OperatorStageResultSuccess

		err = cl.Status().Update(ctx, g)
		require.NoError(t, err)
	})

	// Fetch latest status before each Spec
	BeforeEach(func() {
		By("fetching latest Grafana manifest")
		tmpGrafana := &Grafana{}

		err := cl.Get(ctx, types.NamespacedName{
			Namespace: g.Namespace,
			Name:      g.Name,
		}, tmpGrafana)
		require.NoError(t, err)

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
			err := g.AddNamespacedResource(ctx, cl, lp1, lp1.NamespacedResource(lp1.Spec.CustomUID))
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			assert.NotEmpty(t, result.Status.LibraryPanels)
			assert.Len(t, result.Status.LibraryPanels, 1)

			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			assert.Equal(t, 0, idx)

			assert.Equal(t, lp1.NamespacedResource(lp1.Spec.CustomUID), result.Status.LibraryPanels[idx])
		})

		It("Adds an additional LibraryPanel entries when list is not empty", func() {
			err := g.AddNamespacedResource(ctx, cl, lp2, lp2.NamespacedResource(lp2.Spec.CustomUID))
			require.NoError(t, err)

			err = g.AddNamespacedResource(ctx, cl, lp3, lp3.NamespacedResource(lp3.Spec.CustomUID))
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			assert.NotEmpty(t, result.Status.LibraryPanels)
			assert.Len(t, result.Status.LibraryPanels, 3)

			idx := result.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			assert.Equal(t, 1, idx)
			assert.Equal(t, lp2.NamespacedResource(lp2.Spec.CustomUID), result.Status.LibraryPanels[idx])

			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			assert.Equal(t, 2, idx)
			assert.Equal(t, lp3.NamespacedResource(lp3.Spec.CustomUID), result.Status.LibraryPanels[idx])
		})

		It("Removes LibraryPanel from the middle of a list with multiple entries", func() {
			// Verify state before removal
			idx := g.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			assert.Equal(t, 2, idx)
			assert.Equal(t, lp3.NamespacedResource(lp3.Spec.CustomUID), g.Status.LibraryPanels[idx])

			// Remove middle entry
			err := g.RemoveNamespacedResource(ctx, cl, lp2)
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			assert.NotEmpty(t, result.Status.LibraryPanels)
			assert.Len(t, result.Status.LibraryPanels, 2)

			idx = result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			assert.Equal(t, 0, idx)
			assert.Equal(t, lp1.NamespacedResource(lp1.Spec.CustomUID), result.Status.LibraryPanels[idx])

			// Was removed and should not be found
			idx = result.Status.LibraryPanels.IndexOf(lp2.Namespace, lp2.Name)
			assert.Equal(t, -1, idx)

			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			assert.Equal(t, 1, idx)
			assert.Equal(t, lp3.NamespacedResource(lp3.Spec.CustomUID), result.Status.LibraryPanels[idx])
		})

		It("Removes LibraryPanels from list", func() {
			// Only lp1 and lp3 remains in the Status at this time
			assert.NotEmpty(t, g.Status.LibraryPanels)
			assert.Len(t, g.Status.LibraryPanels, 2)

			err := g.RemoveNamespacedResource(ctx, cl, lp1)
			require.NoError(t, err)

			err = g.RemoveNamespacedResource(ctx, cl, lp3)
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			assert.Equal(t, -1, idx)

			idx = result.Status.LibraryPanels.IndexOf(lp3.Namespace, lp3.Name)
			assert.Equal(t, -1, idx)

			assert.Empty(t, g.Status.LibraryPanels)
		})

		It("Removes LibraryPanels from undefined list", func() {
			assert.Empty(t, g.Status.LibraryPanels)

			err := g.RemoveNamespacedResource(ctx, cl, lp1)
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			idx := result.Status.LibraryPanels.IndexOf(lp1.Namespace, lp1.Name)
			assert.Equal(t, -1, idx)
			assert.Empty(t, g.Status.LibraryPanels)
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
			err := g.AddNamespacedResource(ctx, cl, ds1, ds1.NamespacedResource())
			require.NoError(t, err)

			// Intermediate
			im := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, im)
			require.NoError(t, err)

			err = im.AddNamespacedResource(ctx, cl, ds1, ds1.NamespacedResource())
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			assert.NotEmpty(t, result.Status.Datasources)
			assert.Len(t, result.Status.Datasources, 1)

			idx := result.Status.Datasources.IndexOf(ds1.Namespace, ds1.Name)
			assert.Equal(t, 0, idx)

			assert.Equal(t, ds1.NamespacedResource(), result.Status.Datasources[idx])
		})

		It("Updates existing Datasource on uid changed", func() {
			err := g.AddNamespacedResource(ctx, cl, ds2, ds2.NamespacedResource())
			require.NoError(t, err)

			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)
			require.NoError(t, err)

			err = g.AddNamespacedResource(ctx, cl, ds3, ds3.NamespacedResource())
			require.NoError(t, err)

			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, g)
			require.NoError(t, err)

			assert.NotEmpty(t, g.Status.Datasources)
			assert.Len(t, g.Status.Datasources, 3)

			// Update entry at the middle of the list
			ds2.Spec.CustomUID = "ds-2-unique-identifier"
			err = g.AddNamespacedResource(ctx, cl, ds2, ds2.NamespacedResource())
			require.NoError(t, err)

			result := &Grafana{}
			err = cl.Get(ctx, types.NamespacedName{
				Namespace: g.Namespace,
				Name:      g.Name,
			}, result)
			require.NoError(t, err)

			assert.NotEmpty(t, result.Status.Datasources)
			assert.Len(t, result.Status.Datasources, 3)

			idx := result.Status.Datasources.IndexOf(ds1.Namespace, ds1.Name)
			assert.Equal(t, 0, idx)
			assert.Equal(t, ds1.NamespacedResource(), result.Status.Datasources[idx])

			idx = result.Status.Datasources.IndexOf(ds2.Namespace, ds2.Name)
			assert.Equal(t, 1, idx)
			assert.Equal(t, ds2.NamespacedResource(), result.Status.Datasources[idx])

			idx = result.Status.Datasources.IndexOf(ds3.Namespace, ds3.Name)
			assert.Equal(t, 2, idx)
			assert.Equal(t, ds3.NamespacedResource(), result.Status.Datasources[idx])
		})
	})
})

func TestGetConfigSection(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]map[string]string
		want   map[string]string
	}{
		{
			name:   "nil config",
			config: nil,
			want:   map[string]string{},
		},
		{
			name: "nil config section",
			config: map[string]map[string]string{
				"section": nil,
			},
			want: map[string]string{},
		},
		{
			name: "non-empty config section",
			config: map[string]map[string]string{
				"section": {
					"key": "value",
				},
			},
			want: map[string]string{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := Grafana{
				Spec: GrafanaSpec{
					Config: tt.config,
				},
			}

			got := cr.GetConfigSection("section")
			assert.Equal(t, tt.want, got)
		})
	}
}

var _ = Describe("Grafana External URL validation", func() {
	t := GinkgoT()
	Context("Ensure External.URL follows http/https pattern", func() {
		ctx := context.Background()

		It("Should accept valid http URLs", func() {
			g := &Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-http",
					Namespace: "default",
				},
				Spec: GrafanaSpec{
					External: &External{
						URL: "http://grafana.example.com",
					},
				},
			}

			err := cl.Create(ctx, g)
			require.NoError(t, err)
		})

		It("Should accept valid https URLs", func() {
			g := &Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-https",
					Namespace: "default",
				},
				Spec: GrafanaSpec{
					External: &External{
						URL: "https://grafana.example.com:3000/subpath",
					},
				},
			}

			err := cl.Create(ctx, g)
			require.NoError(t, err)
		})

		It("Should reject URLs without protocol", func() {
			g := &Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-no-protocol",
					Namespace: "default",
				},
				Spec: GrafanaSpec{
					External: &External{
						URL: "grafana.example.com",
					},
				},
			}

			err := cl.Create(ctx, g)
			require.Error(t, err)
		})

		It("Should reject invalid protocols", func() {
			g := &Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-ftp",
					Namespace: "default",
				},
				Spec: GrafanaSpec{
					External: &External{
						URL: "ftp://grafana.example.com",
					},
				},
			}

			err := cl.Create(ctx, g)
			require.Error(t, err)
		})

		It("Should reject empty URLs after protocol", func() {
			g := &Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-empty",
					Namespace: "default",
				},
				Spec: GrafanaSpec{
					External: &External{
						URL: "http://",
					},
				},
			}

			err := cl.Create(ctx, g)
			require.Error(t, err)
		})
	})
})

func TestGetConfigSectionValue(t *testing.T) {
	cr := Grafana{
		Spec: GrafanaSpec{
			Config: map[string]map[string]string{
				"section": {
					"key": "value",
				},
			},
		},
	}

	want := "value"
	got := cr.GetConfigSectionValue("section", "key")

	assert.Equal(t, want, got)
}
