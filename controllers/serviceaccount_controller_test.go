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

	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
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
			// reconcileAndValidateCondition(r, cr, tt.want, tt.wantErr)

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

			Expect(k8sClient.Create(ctx, sa)).Should(Succeed())

			// Reconcile
			r := &GrafanaServiceAccountReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := requestFromMeta(sa.ObjectMeta)
			_, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())

			// Check that the secret was created with correct metadata and data
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: namespace,
			}, secret)).Should(Succeed())

			Expect(secret).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Labels":      HaveKeyWithValue("operator.grafana.com/service-account-name", name),
					"Annotations": HaveKeyWithValue("operator.grafana.com/service-account-token-name", "test-token"),
				}),
				"Data": HaveKeyWithValue("token", Not(BeEmpty())),
			})))

			// Check the status
			updatedSA := &v1beta1.GrafanaServiceAccount{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, updatedSA)).Should(Succeed())

			// Verify that ServiceAccountSynchronized condition is set to success
			containsEqualCondition(updatedSA.Status.Conditions, metav1.Condition{
				Type:   conditionServiceAccountSynchronized,
				Reason: conditionReasonApplySuccessful,
			})

			// Verify that tokens are populated in status
			Expect(updatedSA.Status.Account).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Tokens": ConsistOf(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("test-token"),
					"Secret": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":      Equal(secretName),
						"Namespace": Equal(namespace),
					})),
				})),
			})))

			// Verify that the service account and token were actually created in Grafana
			// Get Grafana client
			gClient, err := client2.NewGeneratedGrafanaClient(ctx, k8sClient, externalGrafanaCr)
			Expect(err).ToNot(HaveOccurred())

			// Retrieve the service account from Grafana API
			saFromGrafana, err := gClient.ServiceAccounts.RetrieveServiceAccountWithParams(
				service_accounts.
					NewRetrieveServiceAccountParamsWithContext(ctx).
					WithServiceAccountID(updatedSA.Status.Account.ID),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(saFromGrafana.Payload).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("test-account-with-token"),
				"Role":       Equal("Admin"),
				"IsDisabled": BeFalse(),
			})))

			// Verify that the token exists in Grafana
			tokensFromGrafana, err := gClient.ServiceAccounts.ListTokensWithParams(
				service_accounts.
					NewListTokensParamsWithContext(ctx).
					WithServiceAccountID(updatedSA.Status.Account.ID),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(tokensFromGrafana.Payload).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("test-token"),
					"ID":   Equal(updatedSA.Status.Account.Tokens[0].ID),
				})),
			))
		})
	})

	Context("When Grafana instance is not ready", func() {
		It("should handle gracefully", func() {
			ctx := context.Background()
			const namespace = "default"
			const name = "test-sa-instance-not-ready"
			const grafanaNotReady = "grafana-not-ready"

			// Create a Grafana instance that is not ready
			notReadyGrafana := &v1beta1.Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      grafanaNotReady,
					Namespace: namespace,
				},
				Spec: v1beta1.GrafanaSpec{
					Config: map[string]map[string]string{
						"security": {
							"admin_user":     "admin",
							"admin_password": "admin",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, notReadyGrafana)).Should(Succeed())

			// Update status to be not ready
			notReadyGrafana.Status = v1beta1.GrafanaStatus{
				Stage:       v1beta1.OperatorStageDeployment,
				StageStatus: v1beta1.OperatorStageResultInProgress,
			}
			Expect(k8sClient.Status().Update(ctx, notReadyGrafana)).Should(Succeed())

			// Create a GrafanaServiceAccount that references the not-ready instance
			sa := &v1beta1.GrafanaServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: v1beta1.GrafanaServiceAccountSpec{
					ResyncPeriod: metav1.Duration{Duration: 10 * time.Minute},
					InstanceName: grafanaNotReady,
					Name:         "test-account-not-ready",
					Role:         "Viewer",
				},
			}

			Expect(k8sClient.Create(ctx, sa)).Should(Succeed())

			// Reconcile
			r := &GrafanaServiceAccountReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			req := requestFromMeta(sa.ObjectMeta)
			_, err := r.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not ready"))

			// Check the status condition
			updatedSA := &v1beta1.GrafanaServiceAccount{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}, updatedSA)).Should(Succeed())

			// Verify that NoMatchingInstance condition is set
			containsEqualCondition(updatedSA.Status.Conditions, metav1.Condition{
				Type:   conditionNoMatchingInstance,
				Reason: "ErrFetchingInstances",
			})
		})
	})
})
