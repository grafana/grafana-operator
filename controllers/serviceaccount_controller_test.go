/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strings"
	"time"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ServiceAccount Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name    string
		meta    metav1.ObjectMeta
		spec    v1beta1.GrafanaServiceAccountSpec
		want    metav1.Condition
		wantErr string
	}{
		{
			name: ".spec.suspend=true",
			meta: objectMetaSuspended,
			spec: v1beta1.GrafanaServiceAccountSpec{
				Name:         objectMetaSuspended.Name,
				InstanceName: grafanaName,
				Suspend:      true,
			},
			want: metav1.Condition{
				Type:   conditionSuspended,
				Reason: conditionReasonApplySuspended,
			},
		},
		{
			name: "LookupGrafana returns nil",
			meta: objectMetaNoMatchingInstances,
			spec: v1beta1.GrafanaServiceAccountSpec{
				Name:         objectMetaNoMatchingInstances.Name,
				InstanceName: "does-not-exist",
			},
			want: metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: conditionReasonEmptyAPIReply,
			},
			wantErr: ErrNoMatchingInstances.Error(),
		},
		{
			name: "Failed to create client",
			meta: objectMetaApplyFailed,
			spec: v1beta1.GrafanaServiceAccountSpec{
				Name:         objectMetaApplyFailed.Name,
				InstanceName: "dummy",
			},
			want: metav1.Condition{
				Type:   conditionServiceAccountSynchronized,
				Reason: conditionReasonApplyFailed,
			},
			wantErr: "building grafana client",
		},
		{
			name: "Successfully applied resource to instance",
			meta: objectMetaSynchronized,
			spec: v1beta1.GrafanaServiceAccountSpec{
				Name:         objectMetaSynchronized.Name,
				InstanceName: grafanaName,
			},
			want: metav1.Condition{
				Type:   conditionServiceAccountSynchronized,
				Reason: conditionReasonApplySuccessful,
			},
		},
	}

	t := GinkgoT()

	for _, tt := range tests {
		It(tt.name, func() {
			cr := &v1beta1.GrafanaServiceAccount{
				ObjectMeta: tt.meta,
				Spec:       tt.spec,
			}
			cr.Spec.Role = "Viewer"
			cr.Spec.ResyncPeriod = metav1.Duration{Duration: 60 * time.Second}

			r := &GrafanaServiceAccountReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			err := k8sClient.Create(testCtx, cr)
			require.NoError(t, err)

			req := requestFromMeta(cr.ObjectMeta)

			_, err = r.Reconcile(testCtx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}

			err = r.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			containsEqualCondition(cr.CommonStatus().Conditions, tt.want)

			err = k8sClient.Delete(testCtx, cr)
			require.NoError(t, err)

			_, err = r.Reconcile(testCtx, req)
			if err != nil && strings.Contains(err.Error(), "dummy-deployment") {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
})

var _ = Describe("ServiceAccount: Tampering with CR or Created ServiceAccount in Grafana", func() {
	tests := []struct {
		name         string
		spec         v1beta1.GrafanaServiceAccountSpec
		scenarioFunc func(client.Client, *v1beta1.GrafanaServiceAccount, *genapi.GrafanaHTTPAPI) error
	}{
		{
			name: "Recreate account deleted from Grafana instance",
			spec: v1beta1.GrafanaServiceAccountSpec{},
			scenarioFunc: func(cl client.Client, cr *v1beta1.GrafanaServiceAccount, gClient *genapi.GrafanaHTTPAPI) error {
				_, err := gClient.ServiceAccounts.DeleteServiceAccount(cr.Status.Account.ID) //nolint:errcheck
				return err
			},
		},
		{
			name: "Recreate token secret when deleted",
			spec: v1beta1.GrafanaServiceAccountSpec{
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{{
					Name: "should-be-recreated",
				}},
			},
			scenarioFunc: func(cl client.Client, cr *v1beta1.GrafanaServiceAccount, gClient *genapi.GrafanaHTTPAPI) error {
				secret := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cr.Status.Account.Tokens[0].Secret.Name,
						Namespace: cr.Status.Account.Tokens[0].Secret.Namespace,
					},
				}
				return cl.Delete(testCtx, &secret)
			},
		},
		{
			name: "Revert account if updated in Grafana",
			spec: v1beta1.GrafanaServiceAccountSpec{},
			scenarioFunc: func(cl client.Client, cr *v1beta1.GrafanaServiceAccount, gClient *genapi.GrafanaHTTPAPI) error {
				_, err := gClient.ServiceAccounts.UpdateServiceAccount( //nolint:errcheck
					service_accounts.
						NewUpdateServiceAccountParams().
						WithServiceAccountID(cr.Status.Account.ID).
						WithBody(&models.UpdateServiceAccountForm{
							Role:       "Admin",
							Name:       "new-name",
							IsDisabled: ptr.To(true),
						}),
				)
				return err
			},
		},
	}

	for _, tt := range tests {
		It(tt.name, func() {
			t := GinkgoT()

			cr := &v1beta1.GrafanaServiceAccount{ObjectMeta: metav1.ObjectMeta{
				Name:      "service-account",
				Namespace: "default",
			}}
			req := requestFromMeta(cr.ObjectMeta)

			r := &GrafanaServiceAccountReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

			cr.Spec = tt.spec
			// 0. Apply defaults
			cr.Spec.InstanceName = grafanaName
			cr.Spec.Name = "serviceaccount"
			cr.Spec.Role = "Viewer"
			cr.Spec.ResyncPeriod = metav1.Duration{Duration: 60 * time.Second}

			// 1. Create
			err := k8sClient.Create(testCtx, cr)
			require.NoError(t, err)

			// 2. Reconcile
			_, err = r.Reconcile(testCtx, req)
			require.NoError(t, err)

			err = k8sClient.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			// 3. Scenario
			gClient, err := client2.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
			require.NoError(t, err)

			err = tt.scenarioFunc(r.Client, cr, gClient)
			require.NoError(t, err)

			// 4. Reconcile
			_, err = r.Reconcile(testCtx, req)
			require.NoError(t, err)

			err = k8sClient.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			// 5. Verify
			err = r.Get(testCtx, req.NamespacedName, cr)
			require.NoError(t, err)

			sa, err := gClient.ServiceAccounts.RetrieveServiceAccount(cr.Status.Account.ID)
			require.NoError(t, err)
			require.NotNil(t, sa)
			require.NotNil(t, sa.Payload)

			require.Equal(t, cr.Spec.Name, sa.Payload.Name)
			require.Equal(t, cr.Spec.Role, sa.Payload.Role)
			require.Equal(t, cr.Spec.IsDisabled, sa.Payload.IsDisabled)
			require.Equal(t, len(cr.Spec.Tokens), int(sa.Payload.Tokens))

			// Status Tokens values retrieved from Grafana, rely on reconciler to update status
			for _, tkSpec := range cr.Spec.Tokens {
				for _, tkStatus := range cr.Status.Account.Tokens {
					if tkSpec.Name != tkStatus.Name {
						continue
					}

					if !tkSpec.Expires.IsZero() {
						require.True(t, isEqualExpirationTime(tkSpec.Expires, tkSpec.Expires))
					}

					if tkSpec.SecretName == "" {
						continue
					}

					require.Contains(t, tkStatus.Secret.Name, tkSpec.SecretName)
				}
			}

			// Check secrets are correctly represented
			for _, tkStatus := range cr.Status.Account.Tokens {
				secretRequest := types.NamespacedName{
					Name:      tkStatus.Secret.Name,
					Namespace: tkStatus.Secret.Namespace,
				}
				s := corev1.Secret{}
				err = k8sClient.Get(testCtx, secretRequest, &s)
				require.NoError(t, err)
				require.NotNil(t, s)
				require.NotEmpty(t, s.Data["token"])
			}

			// 6. Cleanup
			err = k8sClient.Delete(testCtx, cr)
			require.NoError(t, err)

			_, err = r.Reconcile(testCtx, req)
			require.NoError(t, err)
		})
	}
})

var _ = Describe("ServiceAccount Controller: Integration Tests", func() {
	Context("When creating service account with token", func() {
		It("should create secret and verify Grafana API state", func() {
			ctx := context.Background()
			const namespace = "default"
			const name = "test-sa-with-token"
			const secretName = "test-sa-token-secret" // nolint:gosec

			// Create a GrafanaServiceAccount with a token
			sa := &v1beta1.GrafanaServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: v1beta1.GrafanaServiceAccountSpec{
					ResyncPeriod: metav1.Duration{Duration: 10 * time.Minute},
					InstanceName: grafanaName,
					Name:         "test-account-with-token",
					Role:         "Admin",
					Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{
						{
							Name:       "test-token",
							SecretName: secretName,
						},
					},
				},
			}

			t := GinkgoT()
			err := k8sClient.Create(ctx, sa)
			require.NoError(t, err)

			// Reconcile
			r := &GrafanaServiceAccountReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := requestFromMeta(sa.ObjectMeta)
			_, err = r.Reconcile(ctx, req)
			require.NoError(t, err)

			// Check that the secret was created with correct metadata and data
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: namespace,
			}, secret)
			require.NoError(t, err)

			require.Equal(t, "test-token", secret.Annotations["operator.grafana.com/service-account-token-name"])
			require.Equal(t, name, secret.Labels["operator.grafana.com/service-account-name"])
			require.NotEmpty(t, secret.Data["token"])

			// Check the status
			updatedSA := &v1beta1.GrafanaServiceAccount{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, updatedSA)
			require.NoError(t, err)

			// Verify that ServiceAccountSynchronized condition is set to success
			containsEqualCondition(updatedSA.Status.Conditions, metav1.Condition{
				Type:   conditionServiceAccountSynchronized,
				Reason: conditionReasonApplySuccessful,
			})

			// Verify that tokens are populated in status
			require.Equal(t, "test-token", updatedSA.Status.Account.Tokens[0].Name)
			require.Equal(t, secretName, updatedSA.Status.Account.Tokens[0].Secret.Name)
			require.Equal(t, namespace, updatedSA.Status.Account.Tokens[0].Secret.Namespace)

			// Verify that the service account and token were actually created in Grafana
			// Get Grafana client
			gClient, err := client2.NewGeneratedGrafanaClient(ctx, k8sClient, externalGrafanaCr)
			require.NoError(t, err)

			// Retrieve the service account from Grafana API
			saFromGrafana, err := gClient.ServiceAccounts.RetrieveServiceAccountWithParams(
				service_accounts.
					NewRetrieveServiceAccountParamsWithContext(ctx).
					WithServiceAccountID(updatedSA.Status.Account.ID),
			)
			require.NoError(t, err)
			require.Equal(t, "Admin", saFromGrafana.Payload.Role)
			require.Equal(t, "test-account-with-token", saFromGrafana.Payload.Name)
			require.False(t, saFromGrafana.Payload.IsDisabled)

			// Verify that the token exists in Grafana
			tokensFromGrafana, err := gClient.ServiceAccounts.ListTokensWithParams(
				service_accounts.
					NewListTokensParamsWithContext(ctx).
					WithServiceAccountID(updatedSA.Status.Account.ID),
			)
			require.NoError(t, err)
			require.Equal(t, "test-token", tokensFromGrafana.Payload[0].Name)
			require.Equal(t, updatedSA.Status.Account.Tokens[0].ID, tokensFromGrafana.Payload[0].ID)
		})
	})
})
