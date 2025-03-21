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
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	// List Grafana instances
	var grafanas v1beta1.GrafanaList
	if err := r.Client.List(ctx, &grafanas); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing Grafana instances: %w", err)
	}

	// List GrafanaServiceAccount CRs
	var crs v1beta1.GrafanaServiceAccountList
	if err := r.Client.List(ctx, &crs); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing GrafanaServiceAccount resources: %w", err)
	}

	removed, preserved := 0, 0
	for _, grafana := range grafanas.Items {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating Grafana client: %w", err)
		}

		for _, serviceAccountResource := range grafana.Status.ServiceAccounts {
			// Limit deletions per cycle to avoid flooding requests
			if removed >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := serviceAccountResource.Split()
			serviceAccountID, err := strconv.ParseInt(uid, 10, 64)
			if err != nil {
				log.FromContext(ctx).Error(err, "failed to parse service account ID, skip", "namespace", namespace, "name", name, "uid", uid)
				continue
			}

			// Check if service account exists in the cluster
			if cr := crs.Find(serviceAccountResource.Namespace(), serviceAccountResource.Name()); cr != nil && cr.Status.ID == serviceAccountID {
				preserved++
				continue
			}

			if err := r.removeServiceAccount(ctx, serviceAccountID, grafanaClient.ServiceAccounts); err != nil {
				return ctrl.Result{}, fmt.Errorf("removing service account: %w", err)
			}

			grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Remove(namespace, name)
			removed++
		}

		// Update Grafana status after processing each instance
		if err := r.Client.Status().Update(ctx, &grafana); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating Grafana status: %w", err)
		}
	}

	if removed > 0 {
		log.FromContext(ctx).Info("service accounts sync complete", "removed", removed, "preserved", preserved)
	}
	return ctrl.Result{}, nil
}

// cleanupServiceAccount removes tokens and the service account from Grafana.
func (r *GrafanaServiceAccountReconciler) cleanupServiceAccount(
	ctx context.Context,
	grafana *v1beta1.Grafana,
	cr *v1beta1.GrafanaServiceAccount,
	serviceAccountID int64,
	serviceAccountsClient service_accounts.ClientService,
) error {
	l := log.FromContext(ctx).WithValues("serviceAccountID", serviceAccountID)

	// Remove tokens from Grafana and delete their Kubernetes Secrets.
	for _, tokenStatus := range cr.Status.Tokens {
		if err := r.removeToken(ctx, serviceAccountsClient, serviceAccountID, tokenStatus.Name, tokenStatus.TokenID, cr.Namespace, tokenStatus.SecretName); err != nil {
			return fmt.Errorf("removing token %s: %w", tokenStatus.Name, err)
		}
	}

	cr.Status.Tokens = nil
	if err := r.Client.Status().Update(ctx, cr); err != nil {
		l.Error(err, "failed to update CR status after token removal")
	}

	if err := r.removeServiceAccount(ctx, serviceAccountID, serviceAccountsClient); err != nil {
		return fmt.Errorf("removing service account: %w", err)
	}

	grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Remove(cr.Namespace, cr.Name)
	if err := r.Client.Status().Update(ctx, grafana); err != nil {
		return fmt.Errorf("updating Grafana status after service account removal: %w", err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeServiceAccount(
	ctx context.Context,
	serviceAccountID int64,
	serviceAccountsClient service_accounts.ClientService,
) error {
	l := log.FromContext(ctx)

	l.Info("deleting service account from Grafana")
	if _, err := serviceAccountsClient.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID),
	); err != nil {
		var notFound *service_accounts.DeleteServiceAccountInternalServerError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("deleting service account %d: %w", serviceAccountID, err)
		}
		l.Info("service account not found in Grafana, skip")
	}

	return nil
}

// Reconcile contains the main reconciliation logic for GrafanaServiceAccount.
func (r *GrafanaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("GrafanaServiceAccountReconciler")
	ctx = log.IntoContext(ctx, l)

	var cr v1beta1.GrafanaServiceAccount
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, &cr); err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil // CR no longer exists
		}
		return ctrl.Result{}, fmt.Errorf("getting GrafanaServiceAccount: %w", err)
	}

	// If CR is marked for deletion, run finalization logic.
	if cr.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(&cr, grafanaFinalizer) {
			if err := r.finalize(ctx, &cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("finalizing GrafanaServiceAccount: %w", err)
			}
			if err := removeFinalizer(ctx, r.Client, &cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("removing finalizer: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	// Update status and finalizer after reconciliation.
	defer func() {
		cr.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, &cr); err != nil {
			l.Error(err, "failed to update GrafanaServiceAccount status")
		}
		if meta.IsStatusConditionTrue(cr.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, &cr); err != nil {
				l.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, &cr); err != nil {
				l.Error(err, "failed to set finalizer")
			}
		}
	}()

	// Get Grafana grafanas matching the CR.
	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, &cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{}, fmt.Errorf("fetching matching Grafana instances: %w", err)
	}
	if len(grafanas) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, nil)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		cr.Status.NoMatchingInstances = true
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}
	removeNoMatchingInstance(&cr.Status.Conditions)
	cr.Status.NoMatchingInstances = false
	removeInvalidSpec(&cr.Status.Conditions)
	l.Info("found matching Grafana instances", "count", len(grafanas))

	// Apply changes to each matching Grafana instance.
	updateForm := models.CreateServiceAccountForm{
		IsDisabled: cr.Spec.IsDisabled,
		Name:       cr.Spec.Name,
		Role:       cr.Spec.Role,
	}
	applyErrors := map[string]string{}
	for _, grafana := range grafanas {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating Grafana client: %w", err)
		}

		if err = r.setupServiceAccount(ctx, &grafana, &cr, &updateForm, grafanaClient.AccessControl, grafanaClient.ServiceAccounts, grafanaClient.Teams, grafanaClient.Users); err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
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
		found, uid := grafana.Status.ServiceAccounts.Find(cr.Namespace, cr.Name)
		if !found {
			continue
		}
		serviceAccountID, err := strconv.ParseInt(*uid, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing service account ID: %w", err)
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("creating Grafana client: %w", err)
		}

		if err := r.cleanupServiceAccount(ctx, &grafana, cr, serviceAccountID, grafanaClient.ServiceAccounts); err != nil {
			return fmt.Errorf("cleaning up service account: %w", err)
		}
	}

	return nil
}

// setupServiceAccount creates or updates the service account in Grafana.
func (r *GrafanaServiceAccountReconciler) setupServiceAccount(
	ctx context.Context,
	grafana *v1beta1.Grafana,
	cr *v1beta1.GrafanaServiceAccount,
	saForm *models.CreateServiceAccountForm,
	accessControlClient access_control.ClientService,
	serviceAccountsClient service_accounts.ClientService,
	teamsClient teams.ClientService,
	usersClient users.ClientService,
) error {
	var perPage int64 = 1
	serviceAccounts, err := serviceAccountsClient.SearchOrgServiceAccountsWithPaging(
		service_accounts.NewSearchOrgServiceAccountsWithPagingParamsWithContext(ctx).
			WithQuery(&saForm.Name).
			WithPerpage(&perPage),
	)
	if err != nil {
		return fmt.Errorf("searching service accounts: %w", err)
	}
	if serviceAccounts.Payload.TotalCount > 1 {
		return fmt.Errorf("multiple service accounts found with name %s", saForm.Name)
	}
	if serviceAccounts.Payload.TotalCount != int64(len(serviceAccounts.Payload.ServiceAccounts)) {
		return fmt.Errorf("inconsistent service account count: %d != %d", serviceAccounts.Payload.TotalCount, len(serviceAccounts.Payload.ServiceAccounts))
	}

	var serviceAccountID int64
	if len(serviceAccounts.Payload.ServiceAccounts) == 1 {
		if !cr.ResyncPeriodHasElapsed() {
			return nil
		}

		serviceAccountID = serviceAccounts.Payload.ServiceAccounts[0].ID
		if serviceAccountID != cr.Status.ID {
			log.FromContext(ctx).Info("service account ID mismatch, skip", "expected", cr.Status.ID, "actual", serviceAccountID)
			return nil
		}
		if _, err := serviceAccountsClient.UpdateServiceAccount( // nolint:errcheck
			service_accounts.NewUpdateServiceAccountParamsWithContext(ctx).
				WithBody(&models.UpdateServiceAccountForm{
					IsDisabled:       saForm.IsDisabled,
					Name:             saForm.Name,
					Role:             saForm.Role,
					ServiceAccountID: serviceAccountID,
				}).
				WithServiceAccountID(serviceAccountID),
		); err != nil {
			return fmt.Errorf("updating service account: %w", err)
		}
	} else {
		serviceAccount, err := serviceAccountsClient.CreateServiceAccount(
			service_accounts.NewCreateServiceAccountParamsWithContext(ctx).
				WithBody(saForm),
		)
		if err != nil {
			return fmt.Errorf("creating service account: %w", err)
		}
		serviceAccountID = serviceAccount.Payload.ID
	}

	// Save the service account ID in CR status and update Grafana status.
	cr.Status.ID = serviceAccountID
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("serviceAccountID", serviceAccountID))
	grafana.Status.ServiceAccounts = grafana.Status.ServiceAccounts.Add(cr.Namespace, cr.Name, strconv.FormatInt(serviceAccountID, 10))

	if err := r.reconcileTokens(ctx, cr, serviceAccountID, serviceAccountsClient); err != nil {
		return fmt.Errorf("reconciling tokens: %w", err)
	}
	if err := r.reconcilePermissions(ctx, cr.Spec.Permissions, serviceAccountID, accessControlClient, teamsClient, usersClient); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}

	return r.Client.Status().Update(ctx, grafana)
}

// SetupWithManager registers the reconciler with the manager.
func (r *GrafanaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaServiceAccount{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r); err != nil {
		return err
	}

	// Perform initial sync when the manager is ready.
	go func() {
		l := log.FromContext(ctx).WithName("GrafanaServiceAccountReconciler.SetupWithManager")
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
					l.Error(err, "synchronizing service accounts")
					continue
				}
				if res.Requeue {
					l.Info("more service accounts to synchronize")
					continue
				}
				l.Info("service account sync complete")
				return
			}
		}
	}()

	return nil
}

// reconcileTokens creates or updates tokens in Grafana based on the CR spec and removes expired or stale tokens.
func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
	serviceAccountID int64,
	serviceAccountsClient service_accounts.ClientService,
) error {
	expectedTokens := cr.Spec.Tokens
	if cr.Spec.GenerateTokenSecret && len(cr.Spec.Tokens) == 0 {
		expectedTokens = []v1beta1.GrafanaServiceAccountToken{{Name: fmt.Sprintf("%s-default-token", cr.Name)}}
	}

	// Remove stale or expired tokens.
	{
		tokensMap := make(map[string]*metav1.Time, len(expectedTokens))
		for _, token := range expectedTokens {
			tokensMap[token.Name] = token.Expires
		}

		var statusTokens []v1beta1.GrafanaServiceAccountTokenStatus
		for _, tokenStatus := range cr.Status.Tokens {
			if expires, exists := tokensMap[tokenStatus.Name]; exists && (expires == nil || time.Now().Before(expires.Time)) {
				statusTokens = append(statusTokens, tokenStatus)
				continue
			}

			if err := r.removeToken(ctx, serviceAccountsClient, serviceAccountID, tokenStatus.Name, tokenStatus.TokenID, cr.Namespace, tokenStatus.SecretName); err != nil {
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
		if token.Expires != nil && time.Now().After(token.Expires.Time) {
			log.FromContext(ctx).Info("skipping already expired token", "tokenName", token.Name)
			continue
		}

		secretName := token.Name
		if secretName == "" {
			secretName = fmt.Sprintf("%s-default-token", cr.Name)
		}
		if _, exists := existingSecrets[secretName]; exists {
			continue
		}

		// Calculate SecondsToLive if Expires != nil
		var ttlSeconds int64
		if token.Expires != nil {
			ttl := time.Until(token.Expires.Time)
			if ttl <= 0 {
				log.FromContext(ctx).Info("skipping token creation with negative or zero TTL", "tokenName", token.Name, "expires", token.Expires.Time)
				continue
			}
			ttlSeconds = int64(ttl.Seconds())
		}

		cmd := models.AddServiceAccountTokenCommand{
			Name:          token.Name,
			SecondsToLive: ttlSeconds,
		}
		resp, err := serviceAccountsClient.CreateToken(
			service_accounts.
				NewCreateTokenParamsWithContext(ctx).
				WithServiceAccountID(serviceAccountID).
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
		if err := controllerutil.SetControllerReference(cr, secret, r.Scheme); err != nil {
			log.FromContext(ctx).Error(err, "failed to set owner reference on token secret")
		}
		if err := r.Client.Create(ctx, secret); err != nil {
			return fmt.Errorf("creating token secret %s: %w", secretName, err)
		}

		cr.Status.Tokens = append(cr.Status.Tokens, v1beta1.GrafanaServiceAccountTokenStatus{
			Name:       token.Name,
			TokenID:    resp.Payload.ID,
			SecretName: secretName,
		})
		existingSecrets[secretName] = struct{}{}
	}

	// Update CR status with new tokens.
	if err := r.Client.Status().Update(ctx, cr); err != nil {
		log.FromContext(ctx).Error(err, "failed to update GrafanaServiceAccount status with new tokens")
		return err
	}
	return nil
}

// reconcilePermissions assigns or removes RBAC roles based on the CR spec.
func (r *GrafanaServiceAccountReconciler) reconcilePermissions(
	ctx context.Context,
	permissions []v1beta1.GrafanaServiceAccountPermission,
	serviceAccountID int64,
	accessControlClient access_control.ClientService,
	teamsClient teams.ClientService,
	usersClient users.ClientService,
) error {
	l := log.FromContext(ctx)

	const resource = "serviceaccounts"
	resourceID := strconv.FormatInt(serviceAccountID, 10)

	resp, err := accessControlClient.GetResourcePermissionsWithParams(
		access_control.NewGetResourcePermissionsParamsWithContext(ctx).
			WithResource(resource).
			WithResourceID(resourceID),
	)
	if err != nil {
		return fmt.Errorf("listing current resource permissions for service account %d: %w", serviceAccountID, err)
	}

	desiredTeams := map[int64]string{}
	desiredUsers := map[int64]string{}
	for _, permission := range permissions {
		switch {
		case permission.Team != "" && permission.User == "":
			id, err := r.findTeamID(ctx, teamsClient, permission.Team)
			if err != nil {
				return fmt.Errorf("finding team %q: %w", permission.Team, err)
			}
			desiredTeams[id] = permission.Permission
		case permission.Team == "" && permission.User != "":
			id, err := r.findUserID(ctx, usersClient, permission.User)
			if err != nil {
				return fmt.Errorf("finding user %q: %w", permission.User, err)
			}
			desiredUsers[id] = permission.Permission
		default:
			return fmt.Errorf("malfomed permission entry: team=%q user=%q", permission.Team, permission.User)
		}
	}

	for _, resourcePermission := range resp.Payload {
		l := l.WithValues("current_permission", resourcePermission.Permission)

		switch {
		case resourcePermission.TeamID != 0 && resourcePermission.UserID == 0:
			desiredPermission := desiredTeams[resourcePermission.TeamID]
			if desiredPermission == resourcePermission.Permission {
				continue
			}
			l := l.WithValues("team_id", resourcePermission.TeamID, "new_permission", desiredPermission)

			if _, err := accessControlClient.SetResourcePermissionsForTeam( // nolint:errcheck
				access_control.NewSetResourcePermissionsForTeamParamsWithContext(ctx).
					WithBody(&models.SetPermissionCommand{Permission: desiredPermission}).
					WithResource(resource).
					WithResourceID(resourceID).
					WithTeamID(resourcePermission.TeamID),
			); err != nil {
				l.Error(err, "failed to update permission for team")
			} else {
				l.Info("updated permission for team")
			}
		case resourcePermission.TeamID == 0 && resourcePermission.UserID != 0:
			desiredPerm := desiredUsers[resourcePermission.UserID]
			if desiredPerm == resourcePermission.Permission {
				continue
			}
			l := l.WithValues("user_id", resourcePermission.UserID, "new_permission", desiredPerm)

			if _, err := accessControlClient.SetResourcePermissionsForUser( // nolint:errcheck
				access_control.NewSetResourcePermissionsForUserParamsWithContext(ctx).
					WithBody(&models.SetPermissionCommand{Permission: desiredPerm}).
					WithResource(resource).
					WithResourceID(resourceID).
					WithUserID(resourcePermission.UserID),
			); err != nil {
				l.Error(err, "failed to update permission for user")
			} else {
				l.Info("updated permission for user", "userID")
			}
		default:
			return fmt.Errorf("malformed existing permission entry: team_id=%d user_id=%d", resourcePermission.TeamID, resourcePermission.UserID)
		}
	}

	existingTeams := map[int64]string{}
	existingUsers := map[int64]string{}
	for _, dto := range resp.Payload {
		switch {
		case dto.TeamID != 0 && dto.UserID == 0:
			existingTeams[dto.TeamID] = dto.Permission
		case dto.TeamID == 0 && dto.UserID != 0:
			existingUsers[dto.UserID] = dto.Permission
		default:
			return fmt.Errorf("malformed existing permission entry: team_id=%d user_id=%d", dto.TeamID, dto.UserID)
		}
	}
	for teamID, desiredPerm := range desiredTeams {
		if _, exists := existingTeams[teamID]; exists {
			continue
		}
		if _, err := accessControlClient.SetResourcePermissionsForTeam( // nolint:errcheck
			access_control.NewSetResourcePermissionsForTeamParamsWithContext(ctx).
				WithBody(&models.SetPermissionCommand{Permission: desiredPerm}).
				WithResource(resource).
				WithResourceID(resourceID).
				WithTeamID(teamID),
		); err != nil {
			l.Error(err, "failed to grant permission for team", "team_id", teamID, "permission", desiredPerm)
		} else {
			l.Info("granted permission for team", "team_id", teamID, "permission", desiredPerm)
		}
	}
	for userID, desiredPerm := range desiredUsers {
		if _, exists := existingUsers[userID]; exists {
			continue
		}
		if _, err := accessControlClient.SetResourcePermissionsForUser( // nolint:errcheck
			access_control.NewSetResourcePermissionsForUserParamsWithContext(ctx).
				WithBody(&models.SetPermissionCommand{Permission: desiredPerm}).
				WithResource(resource).
				WithResourceID(resourceID).
				WithUserID(userID),
		); err != nil {
			l.Error(err, "failed to grant permission for users", "user_id", userID, "permission", desiredPerm)
		} else {
			l.Info("granted permission for users", "user_id", userID, "permission", desiredPerm)
		}
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeToken(
	ctx context.Context,
	serviceAccountsClient service_accounts.ClientService,
	serviceAccountID int64,
	tokenName string,
	tokenID int64,
	namespace string,
	tokenSecretName string,
) error {
	l := log.FromContext(ctx)

	l.Info("removing token in Grafana", "tokenName", tokenName, "tokenID", tokenID)

	if _, err := serviceAccountsClient.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithTokenID(tokenID),
	); err != nil {
		var notFound *service_accounts.DeleteTokenInternalServerError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("deleting token %d: %w", tokenID, err)
		}
		l.Error(err, "service account token not found in Grafana", "tokenID", tokenID)
	}

	// Delete the Secret if it exists.
	var secret corev1.Secret
	if err := r.Client.Get(ctx, client.ObjectKey{Name: tokenSecretName, Namespace: namespace}, &secret); err == nil {
		if err := r.Client.Delete(ctx, &secret); err != nil {
			l.Error(err, "failed to delete token secret", "secretName", tokenSecretName)
		} else {
			l.Info("removed token secret", "secretName", tokenSecretName)
		}
	} else {
		if !kuberr.IsNotFound(err) {
			return fmt.Errorf("getting token secret %s: %w", tokenSecretName, err)
		}
		l.Info("token secret not found, skip", "secretName", tokenSecretName)
	}

	return nil
}

// findUserID looks up a Grafana user by login/email.
func (r *GrafanaServiceAccountReconciler) findUserID(
	ctx context.Context,
	usersClient users.ClientService,
	loginOrEmail string,
) (int64, error) {
	resp, err := usersClient.GetUserByLoginOrEmailWithParams(
		users.NewGetUserByLoginOrEmailParamsWithContext(ctx).
			WithLoginOrEmail(loginOrEmail),
	)
	if err != nil {
		return 0, fmt.Errorf("searching user %q: %w", loginOrEmail, err)
	}

	return resp.Payload.ID, nil
}

// findTeamID looks up a Grafana team by name.
func (r *GrafanaServiceAccountReconciler) findTeamID(
	ctx context.Context,
	teamsClient teams.ClientService,
	teamName string,
) (int64, error) {
	resp, err := teamsClient.SearchTeams(
		teams.NewSearchTeamsParamsWithContext(ctx).
			WithQuery(&teamName),
	)
	if err != nil {
		return 0, fmt.Errorf("searching teams for %q: %w", teamName, err)
	}
	if resp.Payload.TotalCount > 1 {
		return 0, fmt.Errorf("multiple teams found with name %q", teamName)
	}
	if resp.Payload.TotalCount != int64(len(resp.Payload.Teams)) {
		return 0, fmt.Errorf("inconsistent team count: %d != %d", resp.Payload.TotalCount, len(resp.Payload.Teams))
	}
	if resp.Payload.TotalCount == 0 {
		return 0, fmt.Errorf("team %q not found in Grafana", teamName)
	}

	return resp.Payload.Teams[0].ID, nil
}
