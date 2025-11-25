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
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/ptr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

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
	It("Recreate account deleted from Grafana instance", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "delete-and-recreate",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "delete-and-recreate",
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		_, err = gClient.ServiceAccounts.DeleteServiceAccount(cr.Status.Account.ID) //nolint:errcheck
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		updatedStatus := cr.Status.DeepCopy()
		require.NotEqual(t, originalStatus.Account.ID, updatedStatus.Account.ID)

		deleteCR(t, cr, r)
	})

	It("Recreate token secret when deleted", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "should-be-recreated",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "token-delete-and-recreated",
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{{
					Name: "should-be-recreated",
				}},
			},
		}

		r := createAndReconcileCR(t, cr)

		// Fetch secret for comparison
		originalSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Status.Account.Tokens[0].Secret.Name,
				Namespace: cr.Status.Account.Tokens[0].Secret.Namespace,
			},
		}
		err := k8sClient.Get(testCtx, types.NamespacedName{
			Name:      cr.Status.Account.Tokens[0].Secret.Name,
			Namespace: cr.Status.Account.Tokens[0].Secret.Namespace,
		}, &originalSecret)
		require.NoError(t, err)
		require.NotNil(t, originalSecret.Data)

		// Delete secret
		err = k8sClient.Delete(testCtx, &originalSecret)
		require.NoError(t, err)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		// Expect secret to be recreated during reconcile
		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		// Fetch new secret
		updatedSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Status.Account.Tokens[0].Secret.Name,
				Namespace: cr.Status.Account.Tokens[0].Secret.Namespace,
			},
		}
		err = k8sClient.Get(testCtx, types.NamespacedName{
			Name:      cr.Status.Account.Tokens[0].Secret.Name,
			Namespace: cr.Status.Account.Tokens[0].Secret.Namespace,
		}, &updatedSecret)
		require.NoError(t, err)
		require.NotEmpty(t, originalSecret.Data)
		require.NotEmpty(t, updatedSecret.Data)
		require.NotEqual(t, originalSecret.UID, updatedSecret.UID)
		require.NotEqual(t, originalSecret.Data["token"], updatedSecret.Data["token"])

		deleteCR(t, cr, r)
	})

	It("Revert account if updated in Grafana", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "revert-to-spec",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "revert-to-spec",
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		_, err = gClient.ServiceAccounts.UpdateServiceAccount( //nolint:errcheck
			service_accounts.
				NewUpdateServiceAccountParams().
				WithServiceAccountID(cr.Status.Account.ID).
				WithBody(&models.UpdateServiceAccountForm{
					Role:       "Admin",
					Name:       "new-name",
					IsDisabled: ptr.To(true),
				}),
		)
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		updatedStatus := cr.Status.DeepCopy()
		require.Equal(t, originalStatus.Account.ID, updatedStatus.Account.ID)
		require.Equal(t, originalStatus.Account.Role, updatedStatus.Account.Role)
		require.Equal(t, originalStatus.Account.Name, updatedStatus.Account.Name)
		require.Equal(t, originalStatus.Account.IsDisabled, updatedStatus.Account.IsDisabled)

		deleteCR(t, cr, r)
	})

	It("Add new token", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "add-token",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "add-token",
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{{
					Name: "first",
				}},
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		cr.Spec.Tokens = append(cr.Spec.Tokens, v1beta1.GrafanaServiceAccountTokenSpec{
			Name: "second",
		})
		err := k8sClient.Update(testCtx, cr)
		require.NoError(t, err)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		updatedStatus := cr.Status.DeepCopy()
		require.NotEqual(t, len(originalStatus.Account.Tokens), len(updatedStatus.Account.Tokens))

		deleteCR(t, cr, r)
	})

	It("Update token name", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-token-name",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "update-token-name",
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{{
					Name: "to-be-renamed",
				}},
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		cr.Spec.Tokens[0].Name = "new-token-name"
		err := k8sClient.Update(testCtx, cr)
		require.NoError(t, err)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		updatedStatus := cr.Status.DeepCopy()
		require.NotEqual(t, originalStatus.Account.Tokens[0].Name, updatedStatus.Account.Tokens[0].Name)

		deleteCR(t, cr, r)
	})

	It("Update token secret name", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-token-secret-name",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "update-token-secret-name",
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{{
					Name:       "secret-name-update",
					SecretName: "to-be-renamed",
				}},
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		cr.Spec.Tokens[0].SecretName = "new-secret-name"
		err := k8sClient.Update(testCtx, cr)
		require.NoError(t, err)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		// Ensure original secret was deleted
		originalSecret := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      originalStatus.Account.Tokens[0].Secret.Name,
				Namespace: originalStatus.Account.Tokens[0].Secret.Namespace,
			},
		}
		err = k8sClient.Get(testCtx, types.NamespacedName{
			Name:      originalStatus.Account.Tokens[0].Secret.Name,
			Namespace: originalStatus.Account.Tokens[0].Secret.Namespace,
		}, &originalSecret)
		require.Error(t, err)
		require.True(t, kuberr.IsNotFound(err))

		updatedStatus := cr.Status.DeepCopy()
		require.NotEqual(t, originalStatus.Account.Tokens[0].Secret.Name, updatedStatus.Account.Tokens[0].Secret.Name)

		deleteCR(t, cr, r)
	})

	It("Update token expirations", func() {
		t := GinkgoT()

		cr := &v1beta1.GrafanaServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-token-expirations",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaServiceAccountSpec{
				Name: "update-token-expirations",
				Tokens: []v1beta1.GrafanaServiceAccountTokenSpec{
					{
						Name:    "alter-expiration",
						Expires: &metav1.Time{Time: time.Now().Add(time.Hour)},
					},
					{
						Name:    "remove-expiration",
						Expires: &metav1.Time{Time: time.Now().Add(time.Hour)},
					},
				},
			},
		}

		r := createAndReconcileCR(t, cr)

		originalStatus := cr.Status.DeepCopy()

		cr.Spec.Tokens[0].Expires.Time = time.Now().Add(2 * time.Hour)
		cr.Spec.Tokens[1].Expires = nil
		err := k8sClient.Update(testCtx, cr)
		require.NoError(t, err)

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(testCtx, k8sClient, externalGrafanaCr)
		require.NoError(t, err)

		reconcileAndCompareSpecWithStatus(t, cr, r, gClient)

		updatedStatus := cr.Status.DeepCopy()
		require.NotNil(t, updatedStatus.Account.Tokens[0].Expires)
		require.Nil(t, updatedStatus.Account.Tokens[1].Expires)

		require.WithinDuration(t, time.Now().Add(time.Hour), originalStatus.Account.Tokens[0].Expires.Time, 30*time.Second)
		require.WithinDuration(t, time.Now().Add(2*time.Hour), updatedStatus.Account.Tokens[0].Expires.Time, 30*time.Second)

		deleteCR(t, cr, r)
	})
})

func createAndReconcileCR(t FullGinkgoTInterface, cr *v1beta1.GrafanaServiceAccount) *GrafanaServiceAccountReconciler {
	req := requestFromMeta(cr.ObjectMeta)

	r := &GrafanaServiceAccountReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}

	// Apply defaults
	cr.Spec.InstanceName = grafanaName
	cr.Spec.Role = "Viewer"
	cr.Spec.ResyncPeriod = metav1.Duration{Duration: 60 * time.Second}

	err := k8sClient.Create(testCtx, cr)
	require.NoError(t, err)

	_, err = r.Reconcile(testCtx, req)
	require.NoError(t, err)

	err = k8sClient.Get(testCtx, req.NamespacedName, cr)
	require.NoError(t, err)

	return r
}

func reconcileAndCompareSpecWithStatus(t FullGinkgoTInterface, cr *v1beta1.GrafanaServiceAccount, r *GrafanaServiceAccountReconciler, gClient *genapi.GrafanaHTTPAPI) {
	req := requestFromMeta(cr.ObjectMeta)

	// Reconcile and fetch new object
	_, err := r.Reconcile(testCtx, req)
	require.NoError(t, err)

	err = k8sClient.Get(testCtx, req.NamespacedName, cr)
	require.NoError(t, err)

	// Verify status reflects spec
	err = r.Get(testCtx, req.NamespacedName, cr)
	require.NoError(t, err)

	sa, err := gClient.ServiceAccounts.RetrieveServiceAccount(cr.Status.Account.ID)
	require.NoError(t, err)
	require.NotNil(t, sa)
	require.NotNil(t, sa.Payload)

	require.Equal(t, cr.Spec.Name, sa.Payload.Name)
	require.Equal(t, cr.Spec.Role, sa.Payload.Role)
	require.Equal(t, cr.Spec.IsDisabled, sa.Payload.IsDisabled)
	require.Len(t, cr.Spec.Tokens, int(sa.Payload.Tokens))
	require.Len(t, cr.Status.Account.Tokens, len(cr.Spec.Tokens))

	// Status Tokens values retrieved from Grafana, rely on reconciler to update status
	for _, tkSpec := range cr.Spec.Tokens {
		for _, tkStatus := range cr.Status.Account.Tokens {
			if tkSpec.Name != tkStatus.Name {
				continue
			}

			require.Equal(t, tkSpec.Expires.IsZero(), tkStatus.Expires.IsZero())

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
}

func deleteCR(t FullGinkgoTInterface, cr *v1beta1.GrafanaServiceAccount, r *GrafanaServiceAccountReconciler) {
	req := requestFromMeta(cr.ObjectMeta)

	err := k8sClient.Delete(testCtx, cr)
	require.NoError(t, err)

	_, err = r.Reconcile(testCtx, req)
	require.NoError(t, err)
}

var _ = Describe("ServiceAccount Controller: Integration Tests", func() {
	Context("When creating service account with token", func() {
		It("should create secret and verify Grafana API state", func() {
			ctx := context.Background()
			const namespace = "default"
			const name = "test-sa-with-token"
			const secretName = "test-sa-token-secret" //nolint:gosec

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
			gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, k8sClient, externalGrafanaCr)
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
