package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListDashboard(t *testing.T) {
	t.Run("&Dashboard{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaDashboard{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "Dashboard does not have a case in Grafana.Status.StatusList")
	})
}

func newDashboard(name, uid string) *GrafanaDashboard {
	return &GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaDashboardSpec{
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
			PublicSharing: &GrafanaDashboardPublicSharing{},
		},
	}
}

var _ = Describe("Dashboard type", func() {
	Context("Ensure Dashboard spec.uid is immutable", func() {
		t := GinkgoT()
		ctx := context.Background()

		It("Should block adding 'uid' when missing", func() {
			dash := newDashboard("missing-uid", "")

			// create: Dashboard without uid
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Add uid
			dash.Spec.CustomUID = "new-dash-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block removing 'uid' when set", func() {
			dash := newDashboard("existing-uid", "existing-uid")

			// create: Dashboard with uid
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Remove uid
			dash.Spec.CustomUID = ""
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block updating 'uid'", func() {
			dash := newDashboard("removing-uid", "existing-uid")

			// create: Dashboard with uid
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Update uid
			dash.Spec.CustomUID = "new-dash-uid"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})
	})

	Context("Ensure Public Dashboard 'accessToken' is immutable", func() {
		t := GinkgoT()
		ctx := context.Background()

		It("Should block adding 'accessToken' when missing", func() {
			dash := newDashboard("missing-public-at", "dash-uid")

			// create: Dashboard without accessToken
			dash.Spec.PublicSharing.AccessToken = ""
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Add accessToken
			// The accessToken of public dashboards must be a uuid
			dash.Spec.PublicSharing.AccessToken = "409024ac-1b3a-4e93-a6d1-94002c118530"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block removing 'accessToken' when set", func() {
			dash := newDashboard("existing-public-at", "dash-uid")

			// create: Dashboard with accessToken
			dash.Spec.PublicSharing.AccessToken = "f6055ffe-5d8d-4073-b104-bae0ce2761fa"
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Remove accessToken
			dash.Spec.PublicSharing.AccessToken = ""
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should block updating 'accessToken'", func() {
			dash := newDashboard("removing-public-at", "dash-uid")

			// create: Dashboard with accessToken
			dash.Spec.PublicSharing.AccessToken = "b2516e51-3709-4dd9-ae4b-ae641a67ade5"
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Update accessToken
			dash.Spec.PublicSharing.AccessToken = "5f1393bc-d389-4eec-a416-0e738bb0f30f"
			err = cl.Update(ctx, dash)
			require.Error(t, err)
		})

		It("Should allow updating 'accessToken' when publicDashboard is recreated", func() {
			dash := newDashboard("update-public-at", "dash-uid")

			// create: Dashboard with accessToken
			dash.Spec.PublicSharing.AccessToken = "00ab8af7-6995-42fa-b7da-db3500ec1f9e"
			err := cl.Create(ctx, dash)
			require.NoError(t, err)

			// edit: Disable public dashboard
			dash.Spec.PublicSharing = nil
			err = cl.Update(ctx, dash)
			require.NoError(t, err)

			// edit: Enable public dashboard with new accessToken
			dash.Spec.PublicSharing = &GrafanaDashboardPublicSharing{AccessToken: "87ff0673-dbfd-493d-8834-70a3d300b920"}
			err = cl.Update(ctx, dash)
			require.NoError(t, err)
		})
	})
})
