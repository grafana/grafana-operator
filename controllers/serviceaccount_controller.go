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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/grafana/grafana-openapi-client-go/client/access_control"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/client/teams"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"

	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	conditionServiceAccountSynchronized = "ServiceAccountSynchronized"
)

// GrafanaServiceAccountReconciler reconciles a GrafanaServiceAccount object.
type GrafanaServiceAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaserviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaserviceaccounts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanaserviceaccounts/finalizers,verbs=update

// syncServiceAccounts removes service accounts in Grafana that no longer have a corresponding CR.
func (r *GrafanaServiceAccountReconciler) syncServiceAccounts(ctx context.Context) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// List Grafana instances
	var grafanas v1beta1.GrafanaList
	err := r.Client.List(ctx, &grafanas)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("listing Grafana instances: %w", err)
	}

	// List GrafanaServiceAccount CRs
	var crs v1beta1.GrafanaServiceAccountList
	err = r.Client.List(ctx, &crs)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("listing GrafanaServiceAccount resources: %w", err)
	}

	removed, remained := 0, 0
	for _, grafana := range grafanas.Items {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating Grafana client: %w", err)
		}

		if len(grafana.Status.ServiceAccounts) == 0 {
			continue
		}

		for _, serviceAccountResource := range grafana.Status.ServiceAccounts {
			// Limit deletions per cycle to avoid flooding requests
			if removed >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := serviceAccountResource.Split()
			serviceAccountID, err := strconv.ParseInt(uid, 10, 64)
			if err != nil {
				log.Error(err, "failed to parse service account ID, skip", "resource", string(serviceAccountResource), "uid", uid)
				continue
			}

			// Check if service account exists in the cluster
			cr := crs.Find(serviceAccountResource.Namespace(), serviceAccountResource.Name())
			if cr != nil && cr.Status.ID == serviceAccountID {
				remained++
				continue
			}

			err = r.cleanupServiceAccount(ctx, grafanaClient.ServiceAccounts, cr)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("removing service account: %w", err)
			}

			grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Remove(namespace, name)
			removed++
		}

		// Update Grafana status after processing each instance
		err = r.Client.Status().Update(ctx, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Grafana status: %w", err)
		}
	}

	if removed > 0 {
		log.Info("service accounts sync complete", "removed", removed, "preserved", remained)
	}
	return ctrl.Result{}, nil
}

// cleanupServiceAccount removes tokens and the service account from Grafana.
func (r *GrafanaServiceAccountReconciler) cleanupServiceAccount(
	ctx context.Context,
	serviceAccountsClient service_accounts.ClientService,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	for _, tokenStatus := range cr.Status.Tokens {
		err := r.revokeTokenResources(ctx, serviceAccountsClient, cr.Status.ID, tokenStatus.TokenID, cr.Namespace, tokenStatus.SecretName)
		if err != nil {
			return fmt.Errorf("removing token resources %s: %w", tokenStatus.Name, err)
		}
	}
	cr.Status.Tokens = nil

	err := r.removeServiceAccount(ctx, serviceAccountsClient, cr.Status.ID)
	if err != nil {
		return fmt.Errorf("removing service account %q: %w", cr.Status.ID, err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeServiceAccount(
	ctx context.Context,
	serviceAccountsClient service_accounts.ClientService,
	serviceAccountID int64,
) error {
	log := logf.FromContext(ctx)

	log.Info("deleting service account from Grafana")
	if _, err := serviceAccountsClient.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID),
	); err != nil {
		var notFound *service_accounts.DeleteServiceAccountInternalServerError //TODO: check if this is the correct error type
		if !errors.As(err, &notFound) {
			return fmt.Errorf("deleting service account %d: %w", serviceAccountID, err)
		}
		log.Info("service account not found in Grafana, skip")
	}

	return nil
}

// Reconcile contains the main reconciliation logic for GrafanaServiceAccount.
func (r *GrafanaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler")
	ctx = logf.IntoContext(ctx, log)

	var cr v1beta1.GrafanaServiceAccount
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, &cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil // CR no longer exists
		}
		return ctrl.Result{}, fmt.Errorf("getting GrafanaServiceAccount %q: %w", req, err)
	}

	// If CR is marked for deletion, run finalization logic.
	if cr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(&cr, grafanaFinalizer) {
			err := r.finalize(ctx, &cr)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("finalizing GrafanaServiceAccount %q: %w", req, err)
			}

			err = removeFinalizer(ctx, r.Client, &cr)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("removing finalizer %q: %w", req, err)
			}
		}
		return ctrl.Result{}, nil
	}

	// Update status and finalizer after reconciliation.
	defer func() {
		cr.Status.LastResync = metav1.Time{Time: time.Now()}
		err := r.Status().Update(ctx, &cr)
		if err != nil {
			log.Error(err, "failed to update GrafanaServiceAccount status")
		}
		if meta.IsStatusConditionTrue(cr.Status.Conditions, conditionNoMatchingInstance) {
			err := removeFinalizer(ctx, r.Client, &cr)
			if err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			err := addFinalizer(ctx, r.Client, &cr)
			if err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	// Get Grafana grafanas matching the CR.
	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, &cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		return ctrl.Result{}, fmt.Errorf("fetching Grafana instances: %w", err)
	}
	if len(grafanas) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, nil)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}
	removeNoMatchingInstance(&cr.Status.Conditions)
	removeInvalidSpec(&cr.Status.Conditions)
	log.Info("found matching Grafana instances", "count", len(grafanas))

	// Apply changes to each matching Grafana instance.
	applyErrors := map[string]string{}
	for _, grafana := range grafanas {
		err := r.setupServiceAccount(ctx, &grafana, &cr)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}

		err = r.Client.Status().Update(ctx, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Grafana status: %w", err)
		}
	}

	condition := buildSynchronizedCondition("ServiceAccount", conditionServiceAccountSynchronized, cr.Generation, applyErrors, len(grafanas))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("applying changes to some instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: cr.Spec.ResyncPeriod.Duration}, nil
}

// finalize removes the service account from Grafana when the CR is deleted.
func (r *GrafanaServiceAccountReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) error {
	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances for finalization: %w", err)
	}

	for _, grafana := range grafanas {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("creating Grafana client: %w", err)
		}

		err = r.cleanupServiceAccount(ctx, grafanaClient.ServiceAccounts, cr)
		if err != nil {
			return fmt.Errorf("cleaning up service account: %w", err)
		}

		grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Remove(cr.Namespace, cr.Name)
		err = r.Client.Status().Update(ctx, &grafana)
		if err != nil {
			return fmt.Errorf("updating Grafana status after service account removal: %w", err)
		}
	}

	err = r.Client.Status().Update(ctx, cr)
	if err != nil {
		return fmt.Errorf("updating GrafanaServiceAccount status: %w", err)
	}

	return nil
}

// setupServiceAccount creates or updates the service account in Grafana.
func (r *GrafanaServiceAccountReconciler) setupServiceAccount(
	ctx context.Context,
	grafana *v1beta1.Grafana,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return fmt.Errorf("creating Grafana client: %w", err)
	}

	{
		search, err := grafanaClient.ServiceAccounts.SearchOrgServiceAccountsWithPaging(
			service_accounts.NewSearchOrgServiceAccountsWithPagingParamsWithContext(ctx).
				WithQuery(&cr.Spec.Name),
		)
		if err != nil {
			return fmt.Errorf("searching service accounts: %w", err)
		}
		for _, serviceAccount := range search.Payload.ServiceAccounts {
			if serviceAccount.Name == cr.Spec.Name {
				if cr.Status.ID == 0 || serviceAccount.ID == cr.Status.ID {
					cr.Status.ID = serviceAccount.ID
					break
				}

				err := fmt.Errorf("service account name %q in Grafana has ID=%d, but CR status ID=%d", cr.Spec.Name, serviceAccount.ID, cr.Status.ID)
				meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
					Type:               conditionServiceAccountSynchronized,
					Status:             metav1.ConditionFalse,
					Reason:             "Conflict",
					Message:            err.Error(),
					ObservedGeneration: cr.GetGeneration(),
				})
				logf.FromContext(ctx).Error(err, "service account conflict")
				return err
			}
		}
	}

	if cr.Status.ID == 0 {
		serviceAccount, err := grafanaClient.ServiceAccounts.CreateServiceAccount(
			service_accounts.NewCreateServiceAccountParamsWithContext(ctx).
				WithBody(&models.CreateServiceAccountForm{
					IsDisabled: cr.Spec.IsDisabled,
					Name:       cr.Spec.Name,
					Role:       cr.Spec.Role,
				}),
		)
		if err != nil {
			return fmt.Errorf("creating service account: %w", err)
		}
		cr.Status.ID = serviceAccount.Payload.ID
	} else {
		if _, err := grafanaClient.ServiceAccounts.UpdateServiceAccount( // nolint:errcheck
			service_accounts.NewUpdateServiceAccountParamsWithContext(ctx).
				WithBody(&models.UpdateServiceAccountForm{
					IsDisabled:       cr.Spec.IsDisabled,
					Name:             cr.Spec.Name,
					Role:             cr.Spec.Role,
					ServiceAccountID: cr.Status.ID,
				}).
				WithServiceAccountID(cr.Status.ID),
		); err != nil {
			return fmt.Errorf("updating service account: %w", err)
		}
	}

	// Save the service account ID in CR status and update Grafana status.
	grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Add(cr.Namespace, cr.Name, strconv.FormatInt(cr.Status.ID, 10))

	err = r.reconcileTokens(ctx, grafanaClient.ServiceAccounts, cr)
	if err != nil {
		return fmt.Errorf("reconciling tokens: %w", err)
	}

	err = r.reconcilePermissions(ctx, grafanaClient.AccessControl, grafanaClient.Teams, grafanaClient.Users, cr)
	if err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}

	return nil
}

// SetupWithManager registers the reconciler with the manager.
func (r *GrafanaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaServiceAccount{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
	if err != nil {
		return err
	}

	// Perform initial sync when the manager is ready.
	go func() {
		log := logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler.SetupWithManager")
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(initialSyncDelay):
				start := time.Now()
				res, err := r.syncServiceAccounts(ctx)
				elapsed := time.Since(start).Milliseconds()
				metrics.InitialServiceAccountSyncDuration.Set(float64(elapsed))

				if err != nil {
					log.Error(err, "synchronizing service accounts")
					continue
				}
				if res.Requeue {
					log.Info("more service accounts to synchronize")
					continue
				}
				log.Info("service account sync complete")
				return
			}
		}
	}()

	return nil
}

// reconcileTokens creates or updates tokens in Grafana based on the CR spec and removes expired or stale tokens.
func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	serviceAccountsClient service_accounts.ClientService,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	expectedTokens := cr.Spec.Tokens
	if cr.Spec.GenerateTokenSecret && len(cr.Spec.Tokens) == 0 {
		expectedTokens = []v1beta1.GrafanaServiceAccountToken{{Name: fmt.Sprintf("%s-default-token", cr.Name)}}
	}

	now := time.Now()

	// Remove stale or expired tokens.
	{
		tokensMap := make(map[string]*metav1.Time, len(expectedTokens))
		for _, token := range expectedTokens {
			tokensMap[token.Name] = token.Expires
		}

		var statusTokens []v1beta1.GrafanaServiceAccountTokenStatus
		for _, tokenStatus := range cr.Status.Tokens {
			if expires, exists := tokensMap[tokenStatus.Name]; exists && (expires == nil || now.Before(expires.Time)) {
				statusTokens = append(statusTokens, tokenStatus)
				continue
			}

			err := r.revokeTokenResources(ctx, serviceAccountsClient, cr.Status.ID, tokenStatus.TokenID, cr.Namespace, tokenStatus.SecretName)
			if err != nil {
				return fmt.Errorf("removing token %s: %w", tokenStatus.Name, err)
			}
		}

		cr.Status.Tokens = statusTokens
	}

	// Create tokens that are required but missing.
	existingSecrets := map[string]struct{}{}
	for _, tokenStatus := range cr.Status.Tokens {
		existingSecrets[tokenStatus.SecretName] = struct{}{}
	}
	for _, token := range expectedTokens {
		secretName := token.Name

		if _, exists := existingSecrets[secretName]; exists {
			continue
		}

		cmd := models.AddServiceAccountTokenCommand{
			Name: token.Name,
		}
		if token.Expires != nil {
			cmd.SecondsToLive = int64(time.Until(token.Expires.Time).Seconds())
		}
		resp, err := serviceAccountsClient.CreateToken(
			service_accounts.
				NewCreateTokenParamsWithContext(ctx).
				WithServiceAccountID(cr.Status.ID).
				WithBody(&cmd),
		)
		if err != nil {
			return fmt.Errorf("creating token for service account: %w", err)
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Namespace,
				Labels:    map[string]string{"app": "grafana-serviceaccount-token"},
			},
			Data: map[string][]byte{
				"token": []byte(resp.Payload.Key),
			},
		}
		if token.Expires != nil {
			secret.Annotations = map[string]string{
				"grafana.integreatly.org/token-expiry": token.Expires.Format(time.RFC3339),
			}
		}
		err = controllerutil.SetControllerReference(cr, secret, r.Scheme)
		if err != nil {
			logf.FromContext(ctx).Error(err, "failed to set owner reference on token secret")
		}
		err = r.Client.Create(ctx, secret)
		if err != nil {
			return fmt.Errorf("creating token secret %s: %w", secretName, err)
		}

		cr.Status.Tokens = append(cr.Status.Tokens, v1beta1.GrafanaServiceAccountTokenStatus{
			Name:       token.Name,
			TokenID:    resp.Payload.ID,
			SecretName: secretName,
		})
		existingSecrets[secretName] = struct{}{}
	}

	return nil
}

// reconcilePermissions assigns or removes RBAC roles based on the CR spec.
func (r *GrafanaServiceAccountReconciler) reconcilePermissions(
	ctx context.Context,
	accessControlClient access_control.ClientService,
	teamsClient teams.ClientService,
	usersClient users.ClientService,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	const resource = "serviceaccounts"

	log := logf.FromContext(ctx)

	resourceID := strconv.FormatInt(cr.Status.ID, 10)

	resp, err := accessControlClient.GetResourcePermissionsWithParams(
		access_control.NewGetResourcePermissionsParamsWithContext(ctx).
			WithResource(resource).
			WithResourceID(resourceID),
	)
	if err != nil {
		return fmt.Errorf("listing current resource permissions for service account %d: %w", cr.Status.ID, err)
	}

	type subject struct {
		teamID int64
		userID int64
	}

	desired := map[subject]string{}
	for _, p := range cr.Spec.Permissions {
		switch {
		case p.Team != "" && p.User == "":
			resp, err := teamsClient.SearchTeams(
				teams.NewSearchTeamsParamsWithContext(ctx).
					WithQuery(&p.Team),
			)
			if err != nil {
				return fmt.Errorf("searching teams for %q: %w", p.Team, err)
			}
			if resp.Payload.TotalCount > 1 {
				return fmt.Errorf("multiple teams found with name %q", p.Team)
			}
			if resp.Payload.TotalCount != int64(len(resp.Payload.Teams)) {
				return fmt.Errorf("inconsistent team count: %d != %d", resp.Payload.TotalCount, len(resp.Payload.Teams))
			}
			if resp.Payload.TotalCount == 0 {
				return fmt.Errorf("team %q not found in Grafana", p.Team)
			}
			desired[subject{teamID: resp.Payload.Teams[0].ID}] = p.Permission
		case p.Team == "" && p.User != "":
			resp, err := usersClient.GetUserByLoginOrEmailWithParams(
				users.NewGetUserByLoginOrEmailParamsWithContext(ctx).
					WithLoginOrEmail(p.User),
			)
			if err != nil {
				return fmt.Errorf("searching user %q: %w", p.User, err)
			}
			desired[subject{userID: resp.Payload.ID}] = p.Permission
		default:
			setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidSpec", "either user or team must be set in each permission")
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
			return fmt.Errorf("malformed permission entry: team=%q user=%q", p.Team, p.User)
		}
	}

	var cmds []*models.SetResourcePermissionCommand

	existing := map[subject]struct{}{}
	for _, curr := range resp.Payload {
		s := subject{teamID: curr.TeamID, userID: curr.UserID}
		desiredPerm, inDesired := desired[s]
		if inDesired {
			existing[s] = struct{}{}
		}
		// Assuming that the desired permission is empty if it's not found
		if curr.Permission != desiredPerm {
			cmds = append(cmds, &models.SetResourcePermissionCommand{TeamID: s.teamID, UserID: s.userID, Permission: desiredPerm})
		}
	}

	for s, desiredPerm := range desired {
		if _, exists := existing[s]; !exists {
			cmds = append(cmds, &models.SetResourcePermissionCommand{TeamID: s.teamID, UserID: s.userID, Permission: desiredPerm})
		}
	}

	if len(cmds) == 0 {
		log.V(1).Info("permissions are already up to date")
		return nil
	}

	_, err = accessControlClient.SetResourcePermissions( // nolint:errcheck
		access_control.NewSetResourcePermissionsParamsWithContext(ctx).
			WithBody(&models.SetPermissionsCommand{Permissions: cmds}).
			WithResource(resource).
			WithResourceID(resourceID),
	)
	if err != nil {
		return fmt.Errorf("setting resource permissions for service account %d: %w", cr.Status.ID, err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) revokeTokenResources(
	ctx context.Context,
	serviceAccountsClient service_accounts.ClientService,
	serviceAccountID int64,
	tokenID int64,
	tokenNamespace string,
	tokenSecretName string,
) error {
	log := logf.FromContext(ctx)

	if _, err := serviceAccountsClient.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithTokenID(tokenID),
	); err != nil {
		var notFound *service_accounts.DeleteTokenInternalServerError //TODO: check if this is the correct error type
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("deleting token %d: %w", tokenID, err)
	}

	// Delete the Secret if it exists.
	var secret corev1.Secret
	err := r.Client.Get(ctx, client.ObjectKey{Name: tokenSecretName, Namespace: tokenNamespace}, &secret)
	if err != nil {
		if kuberr.IsNotFound(err) {
			log.Info("token secret not found, skip", "secretName", tokenSecretName)
			return nil
		}
		return fmt.Errorf("getting token secret %s: %w", tokenSecretName, err)
	}

	err = r.Client.Delete(ctx, &secret)
	if err != nil {
		return fmt.Errorf("deleting token secret %s: %w", tokenSecretName, err)
	}

	return nil
}
