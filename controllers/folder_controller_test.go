package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	. "github.com/onsi/ginkgo/v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			wantErr: LogMsgApplyErrors,
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

			r := &GrafanaFolderReconciler{Client: cl, Scheme: cl.Scheme()}

			reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)
		})
	}
})

var _ = Describe("Folder reconciler", func() {
	t := GinkgoT()

	It("marks legacy title-fallback folders as unsafe for folderRef consumers", func() {
		const (
			folderName = "legacy-fallback-folder"
			remoteUID  = "legacy-fallback-uid"
		)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		_, err = gClient.Folders.CreateFolder(&models.CreateFolderCommand{ //nolint:errcheck
			Title: folderName,
			UID:   remoteUID,
		})
		require.NoError(t, err)

		folder := &v1beta1.GrafanaFolder{
			Namespace: "default",
			Name:      folderName,
			Spec: v1beta1.GrafanaFolderSpec{
				GrafanaCommonSpec: commonSpecSynchronized,
			},
		}

		req := tk8s.GetRequest(t, folder)
		reconciler := &GrafanaFolderReconciler{Client: cl, Scheme: cl.Scheme()}

		err = cl.Create(testCtx, folder)
		require.NoError(t, err)

		_, err = reconciler.Reconcile(testCtx, req)
		require.NoError(t, err)

		err = cl.Get(testCtx, req.NamespacedName, folder)
		require.NoError(t, err)
		require.NotEqual(t, remoteUID, folder.GetGrafanaUID())
		assert.True(t, tk8s.HasCondition(t, folder, metav1.Condition{
			Type:   conditionFolderUIDMismatch,
			Reason: conditionReasonFolderUIDInferred,
		}))
	})

	It("successfully deletes folder containing AlertRuleGroup", func() {
		folder := struct {
			cr  *v1beta1.GrafanaFolder
			r   GrafanaFolderReconciler
			req ctrl.Request
		}{
			cr: &v1beta1.GrafanaFolder{
				Namespace: "default",
				Name:      "force-delete",
				Spec: v1beta1.GrafanaFolderSpec{
					GrafanaCommonSpec: commonSpecSynchronized,
					CustomUID:         "force-delete",
				},
			},
			r: GrafanaFolderReconciler{Client: cl, Scheme: cl.Scheme()},
		}
		folder.req = tk8s.GetRequest(t, folder.cr)

		alertRuleGroup := struct {
			cr  *v1beta1.GrafanaAlertRuleGroup
			r   GrafanaAlertRuleGroupReconciler
			req ctrl.Request
		}{
			cr: &v1beta1.GrafanaAlertRuleGroup{
				Namespace: "default",
				Name:      "force-delete",
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
							NoDataState:  new("NoData"),
						},
					},
				},
			},
			r: GrafanaAlertRuleGroupReconciler{Client: cl, Scheme: cl.Scheme()},
		}
		alertRuleGroup.req = tk8s.GetRequest(t, alertRuleGroup.cr)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, cl, externalGrafanaCr)
		require.NoError(t, err)

		// Create folder
		err = cl.Create(testCtx, folder.cr)
		require.NoError(t, err)

		_, err = folder.r.Reconcile(testCtx, folder.req)
		require.NoError(t, err)

		// Create AlertRuleGroup
		err = cl.Create(testCtx, alertRuleGroup.cr)
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
		err = cl.Delete(testCtx, folder.cr)
		require.NoError(t, err)

		_, err = folder.r.Reconcile(testCtx, folder.req)
		require.NoError(t, err)

		// Make sure the folder is gone
		_, err = gClient.Folders.GetFolderByUID(uid) //nolint:errcheck
		require.Error(t, err)
		assert.IsType(t, &folders.GetFolderByUIDNotFound{}, err) //nolint:testifylint
	})
})

var genericFolderInterceptCnt = 0

func TestGenericReconcileRetryOnConflict(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cr := &v1beta1.GrafanaFolder{
		ObjectMeta: objectMetaSuspended,
		Spec: v1beta1.GrafanaFolderSpec{
			Suspend: true,
		},
	}

	interceptFns := &interceptor.Funcs{
		// "Get" Normally will never return a "409 Conflict".
		// But it's a very convenient way of testing the RetryOnConflict wrapper.
		Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if genericFolderInterceptCnt < 2 {
				genericFolderInterceptCnt++

				return apierrors.NewConflict(schema.GroupResource{
					Group:    obj.GetObjectKind().GroupVersionKind().Group,
					Resource: obj.GetObjectKind().GroupVersionKind().Kind,
				}, obj.GetName(), fmt.Errorf("Forced Conflict"))
			}

			return client.Get(ctx, key, obj, opts...)
		},
	}
	cl := tk8s.GetFakeInterceptingClient(t, interceptFns)
	err := v1beta1.AddToScheme(cl.Scheme())
	require.NoError(t, err)

	err = cl.Create(ctx, cr)
	require.NoError(t, err)

	r := NewGenericFolderReconciler(cl, &Config{
		ResyncPeriod: 5 * time.Second,
	})
	req := tk8s.GetRequest(t, cr)

	// Expect Conflict
	_, err = r.ReconcilerWithRetry(ctx, req)
	require.Error(t, err)
	assert.True(t, apierrors.IsConflict(err))
	assert.Equal(t, 1, genericFolderInterceptCnt, "Should return on the first error (conflict)")

	// Reset intercept count to re-enable intercept client
	genericFolderInterceptCnt = 0

	// Retry multiple times until conflict disappears
	_, err = r.Reconcile(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 2, genericFolderInterceptCnt, "Retrying more than once is expected")
}
