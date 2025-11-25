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

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Successfully delete GrafanaFolder with GrafanaAlertRuleGroup referencing it", func() {
	t := GinkgoT()

	f := &v1beta1.GrafanaFolder{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "force-delete",
		},
		Spec: v1beta1.GrafanaFolderSpec{
			GrafanaCommonSpec: commonSpecSynchronized,
			CustomUID:         "force-delete",
		},
	}

	arg := &v1beta1.GrafanaAlertRuleGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "force-delete",
		},
		Spec: v1beta1.GrafanaAlertRuleGroupSpec{
			GrafanaCommonSpec: commonSpecSynchronized,
			FolderRef:         f.Name,
			Interval:          metav1.Duration{Duration: 60 * time.Second},
			Rules: []v1beta1.AlertRule{
				{
					Title:     "TestRule",
					UID:       "akdj-wonvo",
					Condition: "A",
					Data: []*v1beta1.AlertQuery{
						{
							RefID:             "A",
							RelativeTimeRange: nil,
							DatasourceUID:     "__expr__",
							Model: &v1.JSON{Raw: []byte(`{
                                "conditions": [
                                    {
                                        "evaluator": {
                                            "params": [
                                                0,
                                                0
                                            ],
                                            "type": "gt"
                                        },
                                        "operator": {
                                            "type": "and"
                                        },
                                        "query": {
                                            "params": []
                                        },
                                        "reducer": {
                                            "params": [],
                                            "type": "avg"
                                        },
                                        "type": "query"
                                    }
                                ],
                                "datasource": {
                                    "name": "Expression",
                                    "type": "__expr__",
                                    "uid": "__expr__"
                                },
                                "expression": "1 > 0",
                                "hide": false,
                                "intervalMs": 1000,
                                "maxDataPoints": 100,
                                "refId": "B",
                                "type": "math"
                            }`)},
						},
					},
					NoDataState:  ptr.To("NoData"),
					ExecErrState: "Error",
					For:          ptr.To("60s"),
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
					IsPaused:     true,
				},
			},
		},
	}

	It("Creates folder and rule group, deletes folder, checks folder got deleted", func() {
		err := k8sClient.Create(testCtx, f)
		require.NoError(t, err)

		err = k8sClient.Create(testCtx, arg)
		require.NoError(t, err)

		By("Reconcile Folder")
		req := requestFromMeta(f.ObjectMeta)
		fr := GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

		_, err = fr.Reconcile(testCtx, req)
		require.NoError(t, err)

		By("Reconcile AlertRuleGroup")
		req = requestFromMeta(arg.ObjectMeta)
		argr := GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

		_, err = argr.Reconcile(testCtx, req)
		require.NoError(t, err)

		cl, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		By("Verifying folder and alert rule group exist")
		_, err = cl.Folders.GetFolderByUID(f.Spec.CustomUID) //nolint:errcheck
		require.NoErrorf(t, err, "Folder should exist in Grafana")

		_, err = cl.Provisioning.GetAlertRuleGroup(arg.GroupName(), f.Spec.CustomUID) //nolint:errcheck
		require.NoErrorf(t, err, "AlertRuleGroup should exist in Grafana")

		By("Deleting folder")
		err = k8sClient.Delete(testCtx, f)
		require.NoError(t, err)

		_, err = fr.Reconcile(testCtx, req)
		require.NoError(t, err)

		By("Verifying folder is missing")
		_, err = cl.Folders.GetFolderByUID(f.Spec.CustomUID) //nolint:errcheck
		require.Error(t, err)
		assert.IsType(t, &folders.GetFolderByUIDNotFound{}, err)
	})
})
