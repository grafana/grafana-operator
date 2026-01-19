package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListLibraryPanel(t *testing.T) {
	t.Run("&LibraryPanel{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaLibraryPanel{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "LibraryPanel does not have a case in Grafana.Status.StatusList")
	})
}

func newLibraryPanel(name, uid string) *GrafanaLibraryPanel {
	return &GrafanaLibraryPanel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaLibraryPanel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaLibraryPanelSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			GrafanaContentSpec: GrafanaContentSpec{
				CustomUID: uid,
				JSON:      "",
			},
		},
	}
}

var _ = Describe("LibraryPanel type", func() {
	t := GinkgoT()

	Context("Ensure LibraryPanel spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			dash := newLibraryPanel("missing-uid", "")
			By("Create new LibraryPanel without uid")
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("Adding a uid")
			dash.Spec.CustomUID = "new-library-panel-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block removing uid field when set", func() {
			dash := newLibraryPanel("existing-uid", "existing-uid")
			By("Creating LibraryPanel with existing UID")
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("And setting UID to ''")
			dash.Spec.CustomUID = ""
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block changing value of uid", func() {
			dash := newLibraryPanel("removing-uid", "existing-uid")
			By("Create new LibraryPanel with existing UID")
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			By("Changing the existing UID")
			dash.Spec.CustomUID = "new-library-panel-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})
	})
})

var _ = Describe("GrafanaDatasource URL validation", func() {
	t := GinkgoT()
	Context("Ensure datasource URL follows the required pattern", func() {
		ctx := context.Background()

		It("Should accept valid http URLs", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-http-url",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "http://prometheus.monitoring.svc:9090",
					},
				},
			}

			By("Creating GrafanaDatasource with http:// URL")
			err := cl.Create(ctx, ds)
			require.NoError(t, err)
		})

		It("Should accept valid https URLs", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-https-url",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "https://prometheus.example.com:9090/metrics",
					},
				},
			}

			By("Creating GrafanaDatasource with https:// URL")
			err := cl.Create(ctx, ds)
			require.NoError(t, err)
		})

		It("Should accept Grafana template variable URLs", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-template-url",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "${PROMETHEUS_URL}",
					},
				},
			}

			By("Creating GrafanaDatasource with template variable ${...}")
			err := cl.Create(ctx, ds)
			require.NoError(t, err)
		})

		It("Should reject URLs without protocol", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-no-protocol",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "prometheus.monitoring.svc:9090",
					},
				},
			}

			By("Creating GrafanaDatasource without http:// or https://")
			err := cl.Create(ctx, ds)
			require.Error(t, err)
		})

		It("Should reject URLs with invalid protocols", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-ftp-protocol",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "ftp://prometheus.monitoring.svc:9090",
					},
				},
			}

			By("Creating GrafanaDatasource with ftp:// protocol")
			err := cl.Create(ctx, ds)
			require.Error(t, err)
		})

		It("Should reject empty URLs after protocol", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-empty-url",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "http://",
					},
				},
			}

			By("Creating GrafanaDatasource with http:// but no host")
			err := cl.Create(ctx, ds)
			require.Error(t, err)
		})

		It("Should reject malformed template variables", func() {
			ds := &GrafanaDatasource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-template",
					Namespace: "default",
				},
				Spec: GrafanaDatasourceSpec{
					GrafanaCommonSpec: GrafanaCommonSpec{
						InstanceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "datasource",
							},
						},
					},
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						Type: "prometheus",
						URL:  "$PROMETHEUS_URL",
					},
				},
			}

			By("Creating GrafanaDatasource with malformed template variable")
			err := cl.Create(ctx, ds)
			require.Error(t, err)
		})
	})
})
