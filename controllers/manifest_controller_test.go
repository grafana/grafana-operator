package controllers

import (
	"github.com/grafana/grafana-openapi-client-go/client/playlists"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Manifest reconciler", func() {
	t := GinkgoT()

	It("successfully creates playlist from manifest", func() {
		cr := &v1beta1.GrafanaManifest{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "playlist-test",
			},
			Spec: v1beta1.GrafanaManifestSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				Template: v1beta1.GrafanaManifestTemplate{
					Metadata: v1beta1.RequiredObjectMeta{
						Name: "manifest-test",
					},
					RequiredTypeMeta: v1beta1.RequiredTypeMeta{
						APIVersion: "playlist.grafana.app/v0alpha1",
						Kind:       "Playlist",
					},
					Spec: &apiextensionsv1.JSON{
						Raw: []byte(`{
"interval": "5m",
"items": [
  {
    "type":  "dashboard_by_tag",
		"value": "test"
	}
],
"title": "playlist-test"
}`),
					},
				},
			},
		}
		r := GrafanaManifestReconciler{Client: cl, Scheme: cl.Scheme()}
		req := tk8s.GetRequest(t, cr)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		// Create playlist
		err = cl.Create(testCtx, cr)
		require.NoError(t, err)

		_, err = r.Reconcile(testCtx, req)
		require.NoError(t, err)
		playlists, err := gClient.Playlists.SearchPlaylists(playlists.NewSearchPlaylistsParams())
		require.NoError(t, err)
		require.Len(t, playlists.Payload, 1)
	})
})
