package controllers

import (
	"time"

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Successfully delete GrafanaFolder with GrafanaAlertRuleGroup referencing it", func() {
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
	noDataState := "NoData"
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
					NoDataState:  &noDataState,
					ExecErrState: "Error",
					For:          &metav1.Duration{Duration: 60 * time.Second},
					Annotations:  map[string]string{},
					Labels:       map[string]string{},
					IsPaused:     true,
				},
			},
		},
	}

	It("Creates folder and rule group, deletes folder, checks folder got deleted", func() {
		Expect(k8sClient.Create(testCtx, f)).To(Succeed())
		Expect(k8sClient.Create(testCtx, arg)).To(Succeed())

		By("Reconcile Folder")
		req := requestFromMeta(f.ObjectMeta)
		fr := GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err := fr.Reconcile(testCtx, req)
		Expect(err).ToNot(HaveOccurred())

		By("Reconcile AlertRuleGroup")
		req = requestFromMeta(arg.ObjectMeta)
		argr := GrafanaAlertRuleGroupReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		_, err = argr.Reconcile(testCtx, req)
		Expect(err).ToNot(HaveOccurred())

		cl, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		Expect(err).ToNot(HaveOccurred())

		By("Verifying folder and alert rule group exist")
		_, err = cl.Folders.GetFolderByUID(f.Spec.CustomUID) // nolint:errcheck
		Expect(err).NotTo(HaveOccurred(), "Folder should exist in Grafana")

		_, err = cl.Provisioning.GetAlertRuleGroup(arg.GroupName(), f.Spec.CustomUID) // nolint:errcheck
		Expect(err).NotTo(HaveOccurred(), "AlertRuleGroup should exist in Grafana")

		By("Deleting folder")
		Expect(k8sClient.Delete(testCtx, f)).Should(Succeed())

		_, err = fr.Reconcile(testCtx, req)
		Expect(err).ToNot(HaveOccurred())

		By("Verifying folder is missing")
		_, err = cl.Folders.GetFolderByUID(f.Spec.CustomUID) // nolint:errcheck
		Expect(err).To(HaveOccurred())

		var notFound *folders.GetFolderByUIDNotFound
		Expect(err).Should(BeAssignableToTypeOf(notFound), "Folder should have been removed from Grafana")
	})
})
