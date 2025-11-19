package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListDatasource(t *testing.T) {
	t.Run("&Datasource{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaDatasource{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "Datasource does not have a case in Grafana.Status.StatusList")
	})
}

func newDatasource(name string, uid string) *GrafanaDatasource {
	return &GrafanaDatasource{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDatasourceSpec{
			CustomUID: uid,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "datasource",
					},
				},
			},
			Datasource: &GrafanaDatasourceInternal{
				Name:   "testdata",
				Type:   "grafana-testdata-datasource",
				Access: "proxy",
			},
		},
	}
}

var _ = Describe("Datasource type", func() {
	t := GinkgoT()

	Context("Ensure Datasource spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			ds := newDatasource("missing-uid", "")
			By("Create new Datasource without uid")
			err := k8sClient.Create(ctx, ds)
			require.NoError(t, err)

			By("Adding a uid")
			ds.Spec.CustomUID = "new-ds-uid"
			err = k8sClient.Update(ctx, ds)
			require.Error(t, err)
		})

		It("Should block removing uid field when set", func() {
			ds := newDatasource("existing-uid", "existing-uid")
			By("Creating Datasource with existing UID")
			err := k8sClient.Create(ctx, ds)
			require.NoError(t, err)

			By("And setting UID to ''")
			ds.Spec.CustomUID = ""
			err = k8sClient.Update(ctx, ds)
			require.Error(t, err)
		})

		It("Should block changing value of uid", func() {
			ds := newDatasource("removing-uid", "existing-uid")
			By("Create new Datasource with existing UID")
			err := k8sClient.Create(ctx, ds)
			require.NoError(t, err)

			By("Changing the existing UID")
			ds.Spec.CustomUID = "new-ds-uid"
			err = k8sClient.Update(ctx, ds)
			require.Error(t, err)
		})
	})
})

var _ = Describe("Fail on field behavior changes", func() {
	t := GinkgoT()

	emptyDatasource := &GrafanaDatasource{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaDatasource",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-nil-datasource",
			Namespace: "default",
		},
		Spec: GrafanaDatasourceSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &v1.LabelSelector{},
			},
			Datasource: nil,
		},
	}

	ctx := context.Background()
	It("Fails creating GrafanaDatasource with undefined spec.datasource", func() {
		err := k8sClient.Create(ctx, emptyDatasource)
		require.Error(t, err)
	})
})
