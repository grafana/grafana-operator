package controllers

import (
	"time"

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Folder reconciler", func() {
	t := GinkgoT()

	It("successfully deletes folder containing AlertRuleGroup", func() {
		folder := struct {
			cr  *v1beta1.GrafanaFolder
			r   GrafanaFolderReconciler
			req controllerruntime.Request
		}{}

		folder.cr = &v1beta1.GrafanaFolder{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "force-delete",
			},
			Spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				CustomUID:         "force-delete",
			},
		}
		folder.r = GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		folder.req = requestFromMeta(folder.cr.ObjectMeta)

		alertRuleGroup := struct {
			cr  *v1beta1.GrafanaAlertRuleGroup
			r   GrafanaAlertRuleGroupReconciler
			req controllerruntime.Request
		}{}

		alertRuleGroup.cr = &v1beta1.GrafanaAlertRuleGroup{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "force-delete",
			},
			Spec: v1beta1.GrafanaAlertRuleGroupSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
				FolderRef:         folder.cr.Name,
				Interval:          metav1.Duration{Duration: 60 * time.Second},
				Rules: []v1beta1.AlertRule{
					{
						Title:     "TestRule",
						UID:       "force-delete",
						Condition: "A",
						Data: []*v1beta1.AlertQuery{
							{
								RefID:         "A",
								DatasourceUID: "__expr__",
								Model:         &v1.JSON{Raw: []byte(`{"expression": "1", "refId": "A"}`)},
							},
						},
						ExecErrState: "Error",
						NoDataState:  ptr.To("NoData"),
					},
				},
			},
		}
		alertRuleGroup.r = GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		alertRuleGroup.req = requestFromMeta(alertRuleGroup.cr.ObjectMeta)

		grafanaClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		// Create folder
		err = k8sClient.Create(testCtx, folder.cr)
		require.NoError(t, err)

		_, err = folder.r.Reconcile(testCtx, folder.req)
		require.NoError(t, err)

		// Create AlertRuleGroup
		err = k8sClient.Create(testCtx, alertRuleGroup.cr)
		require.NoError(t, err)

		_, err = alertRuleGroup.r.Reconcile(testCtx, alertRuleGroup.req)
		require.NoError(t, err)

		// Make sure both resources exist in Grafana
		uid := folder.cr.Spec.CustomUID

		_, err = grafanaClient.Folders.GetFolderByUID(uid) //nolint:errcheck
		require.NoErrorf(t, err, "Folder should exist in Grafana")

		_, err = grafanaClient.Provisioning.GetAlertRuleGroup(alertRuleGroup.cr.GroupName(), uid) //nolint:errcheck
		require.NoErrorf(t, err, "AlertRuleGroup should exist in Grafana")

		// Deleting folder
		err = k8sClient.Delete(testCtx, folder.cr)
		require.NoError(t, err)

		_, err = folder.r.Reconcile(testCtx, folder.req)
		require.NoError(t, err)

		// Make sure the folder is gone
		_, err = grafanaClient.Folders.GetFolderByUID(uid) //nolint:errcheck
		require.Error(t, err)
		assert.IsType(t, &folders.GetFolderByUIDNotFound{}, err)
	})
})
