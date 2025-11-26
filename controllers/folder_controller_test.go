package controllers

import (
	"time"

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Folder Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaFolderSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSuspended,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "GetScopedMatchingInstances returns empty list",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecNoMatchingInstances,
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to apply to instance",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecApplyFailed,
			},
			want: metav1.Condition{
				Type:   conditionFolderSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "failed to apply to all instances",
		},
		{
			name: "InvalidSpec Condition",
			meta: objectMetaInvalidSpec,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecInvalidSpec,
				CustomUID:         "self-ref",
				ParentFolderUID:   "self-ref",
			},
			want: metav1.Condition{
				Type:   conditionInvalidSpec,
				Reason: conditionReasonCyclicParent,
			},
			wantErr: "cyclic folder reference",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
			},
			want: metav1.Condition{
				Type:   conditionFolderSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaFolder{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}

			r := &GrafanaFolderReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("Folder reconciler", func() {
	t := GinkgoT()

	It("successfully deletes folder containing AlertRuleGroup", func() {
		folder := struct {
			cr  *v1beta1.GrafanaFolder
			r   GrafanaFolderReconciler
			req ctrl.Request
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
			req ctrl.Request
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
								Model:         &apiextensionsv1.JSON{Raw: []byte(`{"expression": "1", "refId": "A"}`)},
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

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
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

		_, err = gClient.Folders.GetFolderByUID(uid) //nolint:errcheck
		require.NoErrorf(t, err, "Folder should exist in Grafana")

		_, err = gClient.Provisioning.GetAlertRuleGroup(alertRuleGroup.cr.GroupName(), uid) //nolint:errcheck
		require.NoErrorf(t, err, "AlertRuleGroup should exist in Grafana")

		// Delete folder
		err = k8sClient.Delete(testCtx, folder.cr)
		require.NoError(t, err)

		_, err = folder.r.Reconcile(testCtx, folder.req)
		require.NoError(t, err)

		// Make sure the folder is gone
		_, err = gClient.Folders.GetFolderByUID(uid) //nolint:errcheck
		require.Error(t, err)
		assert.IsType(t, &folders.GetFolderByUIDNotFound{}, err)
	})
})
