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

// Package controllers implements Kubernetes controllers for Grafana Operator.
package controllers

// GrafanaServiceAccountReconciler manages the lifecycle of Grafana service accounts and their tokens.
//
// The controller ensures that service accounts in Grafana match the desired state defined in
// GrafanaServiceAccount custom resources. It handles:
//   - Service account creation, updates, and deletion
//   - Token lifecycle management with automatic recreation on expiration changes
//   - Secure token storage in Kubernetes Secrets
//   - Cleanup of orphaned resources
//
// Key architectural decisions:
//   - Tokens are immutable in Grafana - any change requires recreation
//   - Token names must be unique within a service account (enforced by CRD validation)
//   - Secrets use annotations to link them with their corresponding tokens
//   - The controller follows eventual consistency model, handling external modifications gracefully
//   - All resources are created in the same namespace as the CR for security
//
// The reconciliation process is idempotent and can recover from partial failures or
// external modifications to either Grafana or Kubernetes resources.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	model2 "github.com/grafana/grafana-operator/v5/controllers/model"
)

const (
	conditionServiceAccountSynchronized = "ServiceAccountSynchronized"
)

// GrafanaServiceAccountReconciler reconciles a GrafanaServiceAccount object.
type GrafanaServiceAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

// Reconcile synchronizes the actual state (Grafana service accounts and Kubernetes secrets)
// with the desired state defined in the GrafanaServiceAccount CR spec,
// taking into account Kubernetes' eventual consistency model.
//
// The reconciliation process:
// 1. Fetches the GrafanaServiceAccount resource from Kubernetes
// 2. Handles resource deletion (removes service account from Grafana and cleans up secrets)
// 3. Sets up status update handling (deferred)
// 4. Establishes connection to the target Grafana instance
// 5. For active resources - reconciles the actual state with the desired state (creates, updates, removes as needed)
// 6. Updates the resource status with current state and conditions
// 7. Schedules periodic reconciliation based on ResyncPeriod
func (r *GrafanaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler")
	ctx = logf.IntoContext(ctx, log)

	// 1. Fetch the GrafanaServiceAccount resource from Kubernetes
	cr := &v1beta1.GrafanaServiceAccount{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("getting GrafanaServiceAccount %q: %w", req, err)
	}

	// 2. Handle resource deletion (removes service account from Grafana and cleans up secrets)
	if cr.GetDeletionTimestamp() != nil {
		err := r.finalize(ctx, cr)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("finalizing GrafanaServiceAccount: %w", err)
		}

		err = removeFinalizer(ctx, r.Client, cr)
		if err != nil && !kuberr.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("removing finalizer: %w", err)
		}

		return ctrl.Result{}, nil
	}

	// 3. From here on, we're handling normal reconciliation (not deletion)
	defer UpdateStatus(ctx, r.Client, cr)

	// Check if reconciliation is suspended
	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	// 4. Establish connection to the target Grafana instance
	// First, get the Grafana CR
	meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)

	grafana, err := r.lookupGrafana(ctx, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionContactPointSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instance: %w", err)
	}

	if grafana == nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionContactPointSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)

	// 5. For active resources - reconcile the actual state with the desired state (creates, updates, removes as needed)
	err = r.reconcileWithInstance(ctx, cr, grafana)

	// 6. Update the resource status with current state and conditions
	applyErrors := map[string]string{}
	if err != nil {
		applyErrors[grafana.Name] = err.Error()
	}

	condition := buildSynchronizedCondition(
		"ServiceAccount",
		conditionServiceAccountSynchronized,
		cr.Generation,
		applyErrors,
		1, // We always synchronize with a single Grafana instance.
	)
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling service account: %w", err)
	}

	// 7. Schedule periodic reconciliation based on ResyncPeriod
	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

// finalize handles the cleanup logic when a GrafanaServiceAccount resource is being deleted.
// It attempts to remove the service account from Grafana and clean up associated secrets.
func (r *GrafanaServiceAccountReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) error {
	if !controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
		return nil
	}

	if cr.Status.Account == nil {
		return nil
	}

	// Get the Grafana CR for deletion
	grafana, err := r.lookupGrafana(ctx, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return nil
		}

		return err
	}

	gClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return fmt.Errorf("creating Grafana client: %w", err)
	}

	_, err = gClient.ServiceAccounts.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(cr.Status.Account.ID),
	)
	if err != nil {
		// ATM, service_accounts.DeleteServiceAccountNotFound doesn't have Is, Unwrap, Unwrap.
		// So, we cannot rely only on errors.Is().
		_, ok := err.(*service_accounts.DeleteServiceAccountNotFound) // nolint:errorlint
		if ok || errors.Is(err, service_accounts.NewDeleteServiceAccountNotFound()) {
			logf.FromContext(ctx).Info("service account not found, skipping removal",
				"serviceAccountID", cr.Status.Account.ID,
				"serviceAccountName", cr.Spec.Name,
			)

			return nil
		}

		// TODO: The operator now deploys Grafana 12.1.0 by default (see controllers/config/operator_constants.go#L6),
		// but it may still manage older Grafana instances.
		//
		// Before Grafana 12.0.2, there was no reliable way to detect a 404 when deleting a service account.
		// The API returned 500 instead (see https://github.com/grafana/grafana/issues/106618).
		//
		// Once we can guarantee all managed instances are >= 12.0.2 we can handle the real 404 explicitly.
		//
		// Until then, we treat any non-nil error from the delete call as "already removed" and just log it for visibility.
		logf.FromContext(ctx).Error(err, "failed to delete service account (may already be deleted)",
			"serviceAccountID", cr.Status.Account.ID,
			"serviceAccountName", cr.Spec.Name,
		)
		// return fmt.Errorf("deleting service account %q: %w", status.SpecID, err)
	}

	return nil
}

// lookupGrafana retrieves the Grafana instance referenced by the GrafanaServiceAccount
// and validates that it's in a ready state for accepting API requests.
func (r *GrafanaServiceAccountReconciler) lookupGrafana(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
) (*v1beta1.Grafana, error) {
	var grafana v1beta1.Grafana

	err := r.Get(ctx, client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      cr.Spec.InstanceName,
	}, &grafana)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	// Check if Grafana instance is ready
	if grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
		return nil, fmt.Errorf("Grafana instance %q is not ready (stage: %q, status: %q)", cr.Spec.InstanceName, grafana.Status.Stage, grafana.Status.StageStatus) // nolint:staticcheck
	}

	return &grafana, nil
}

// reconcileWithInstance performs the core reconciliation logic for active resources.
//
// It orchestrates the complete synchronization process:
// 1. Ensures the service account exists in Grafana (creates if missing)
// 2. Updates service account properties to match the spec
// 3. Manages the lifecycle of authentication tokens and their secrets
func (r *GrafanaServiceAccountReconciler) reconcileWithInstance(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
	grafana *v1beta1.Grafana,
) error {
	gClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	err = r.upsertAccount(ctx, gClient, cr)
	if err != nil {
		return fmt.Errorf("upserting service account: %w", err)
	}

	// Ensure tokens are always sorted for stable ordering
	defer func() {
		sort.Slice(cr.Status.Account.Tokens, func(i, j int) bool {
			return cr.Status.Account.Tokens[i].Name < cr.Status.Account.Tokens[j].Name
		})
	}()

	// Phase 1: Prune orphaned secrets and index valid ones
	secretsByTokenName, err := r.pruneAndIndexSecrets(ctx, cr)
	if err != nil {
		return err
	}

	// Phase 2: Remove outdated tokens (will be recreated with correct configuration)
	err = r.removeOutdatedTokens(ctx, gClient, cr)
	if err != nil {
		return err
	}

	// Phase 3: Validate existing tokens and restore their secret references
	err = r.validateAndRestoreTokenSecrets(ctx, gClient, cr, secretsByTokenName)
	if err != nil {
		return err
	}

	// Phase 4: Provision missing tokens
	tokensToCreate := r.determineMissingTokens(cr)

	err = r.provisionTokens(ctx, gClient, cr, tokensToCreate, secretsByTokenName)
	if err != nil {
		return err
	}

	if len(cr.Status.Account.Tokens) != 0 {
		// Grafana's create token API doesn't return expiration, requiring a separate fetch
		return r.populateTokenExpirations(ctx, gClient, cr)
	}

	return nil
}

// convertGrafanaExpiration converts Grafana's strfmt.DateTime to Kubernetes metav1.Time pointer.
// Returns nil if the expiration is zero.
func convertGrafanaExpiration(expiration strfmt.DateTime) *metav1.Time {
	if expiration.IsZero() {
		return nil
	}

	return ptr.To(metav1.NewTime(time.Time(expiration)))
}

// removeOutdatedTokens removes tokens that are not in the desired spec or have mismatched expiration times.
// These tokens will be recreated later with the correct configuration.
func (r *GrafanaServiceAccountReconciler) removeOutdatedTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	// Build map of desired tokens from spec
	desiredTokens := make(map[string]v1beta1.GrafanaServiceAccountTokenSpec, len(cr.Spec.Tokens))
	for _, token := range cr.Spec.Tokens {
		desiredTokens[token.Name] = token
	}

	for i := 0; i < len(cr.Status.Account.Tokens); i++ {
		tokenName := cr.Status.Account.Tokens[i].Name
		desiredToken, ok := desiredTokens[tokenName]

		needsRecreation := !ok ||
			!isEqualExpirationTime(desiredToken.Expires, cr.Status.Account.Tokens[i].Expires)

		if needsRecreation {
			err := r.removeAccountToken(ctx, gClient, cr.Status.Account.ID, &cr.Status.Account.Tokens[i])
			if err != nil {
				return fmt.Errorf("removing service account token %q: %w", tokenName, err)
			}

			cr.Status.Account.Tokens = slices.Delete(cr.Status.Account.Tokens, i, i+1)
			i--
		}
	}

	return nil
}

// provisionTokens creates the specified tokens in Grafana and ensures their secrets exist in Kubernetes.
// For each token, it creates the token in Grafana and either updates an existing secret or creates a new one.
func (r *GrafanaServiceAccountReconciler) provisionTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
	tokensToCreate []v1beta1.GrafanaServiceAccountTokenSpec,
	secretsByTokenName map[string]corev1.Secret,
) error {
	for _, tokenSpec := range tokensToCreate {
		tokenStatus, tokenKey, err := r.createToken(ctx, gClient, cr.Status.Account.ID, tokenSpec)
		if err != nil {
			return fmt.Errorf("creating token %q: %w", tokenSpec.Name, err)
		}

		// Check what we should do with the secret.
		secret, ok := secretsByTokenName[tokenSpec.Name]
		if ok {
			// The secret already exists, so we can just update it.
			renewSecret(&secret, tokenStatus, tokenKey)

			err := r.Update(ctx, &secret)
			if err != nil {
				return fmt.Errorf("updating token secret %q: %w", secret.Name, err)
			}
		} else {
			// The secret doesn't exist, so we need to create it.
			newSecret := buildTokenSecret(ctx, cr, tokenSpec, tokenStatus, tokenKey, r.Scheme)

			err := r.Create(ctx, newSecret)
			if err != nil {
				return fmt.Errorf("creating token secret %q: %w", newSecret.Name, err)
			}

			secret = *newSecret
		}

		tokenStatus.Secret = &v1beta1.GrafanaServiceAccountSecretStatus{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		}

		cr.Status.Account.Tokens = append(cr.Status.Account.Tokens, tokenStatus)
	}

	return nil
}

// determineMissingTokens returns a sorted list of tokens that are in the spec but not in the current status.
// These are the tokens that need to be created.
func (r *GrafanaServiceAccountReconciler) determineMissingTokens(cr *v1beta1.GrafanaServiceAccount) []v1beta1.GrafanaServiceAccountTokenSpec {
	// Build map of desired tokens from spec
	desiredTokens := make(map[string]v1beta1.GrafanaServiceAccountTokenSpec, len(cr.Spec.Tokens))
	for _, token := range cr.Spec.Tokens {
		desiredTokens[token.Name] = token
	}

	// Remove tokens that already exist in status
	for _, token := range cr.Status.Account.Tokens {
		delete(desiredTokens, token.Name)
	}

	// Convert map to sorted slice for stable ordering
	tokensToCreate := make([]v1beta1.GrafanaServiceAccountTokenSpec, 0, len(desiredTokens))
	for _, desiredToken := range desiredTokens {
		tokensToCreate = append(tokensToCreate, desiredToken)
	}

	// Sort for stable order - makes debugging and testing easier
	slices.SortFunc(tokensToCreate, func(a, b v1beta1.GrafanaServiceAccountTokenSpec) int {
		return strings.Compare(a.Name, b.Name)
	})

	return tokensToCreate
}

// validateAndRestoreTokenSecrets removes tokens without corresponding secrets and restores secret references for valid tokens.
// Tokens without secrets will be recreated with new secrets in the next phase.
func (r *GrafanaServiceAccountReconciler) validateAndRestoreTokenSecrets(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
	secretsByTokenName map[string]corev1.Secret,
) error {
	for i := 0; i < len(cr.Status.Account.Tokens); i++ {
		tokenName := cr.Status.Account.Tokens[i].Name
		secret, secretExists := secretsByTokenName[tokenName]

		if !secretExists {
			err := r.removeAccountToken(ctx, gClient, cr.Status.Account.ID, &cr.Status.Account.Tokens[i])
			if err != nil {
				return fmt.Errorf("removing service account token %q: %w", tokenName, err)
			}

			cr.Status.Account.Tokens = slices.Delete(cr.Status.Account.Tokens, i, i+1)
			i--

			continue
		}

		// Restore secret reference for valid token
		cr.Status.Account.Tokens[i].Secret = &v1beta1.GrafanaServiceAccountSecretStatus{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		}
	}

	return nil
}

// createToken creates a new token in Grafana for the service account.
// Returns the token status, the token key (secret value), and any error.
func (r *GrafanaServiceAccountReconciler) createToken(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
	tokenSpec v1beta1.GrafanaServiceAccountTokenSpec,
) (v1beta1.GrafanaServiceAccountTokenStatus, []byte, error) {
	cmd := models.AddServiceAccountTokenCommand{Name: tokenSpec.Name}
	if tokenSpec.Expires != nil {
		// Note: We pass potentially negative TTL to Grafana API and let it handle the validation.
		// This approach handles edge cases like clock drift, timezone differences, and API processing delays.
		// Grafana will reject tokens with invalid TTL values appropriately.
		cmd.SecondsToLive = int64(time.Until(tokenSpec.Expires.Time).Seconds())
	}

	createResp, err := gClient.ServiceAccounts.CreateToken(
		service_accounts.
			NewCreateTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithBody(&cmd),
	)
	if err != nil {
		return v1beta1.GrafanaServiceAccountTokenStatus{}, nil, err
	}

	tokenStatus := v1beta1.GrafanaServiceAccountTokenStatus{
		Name: createResp.Payload.Name,
		ID:   createResp.Payload.ID,
	}

	return tokenStatus, []byte(createResp.Payload.Key), nil
}

// populateTokenExpirations fetches token expiration times from Grafana and updates the status.
// This is a workaround for Grafana API limitation where the create token response doesn't include expiration.
func (r *GrafanaServiceAccountReconciler) populateTokenExpirations(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	listResp, err := gClient.ServiceAccounts.ListTokensWithParams(
		service_accounts.
			NewListTokensParamsWithContext(ctx).
			WithServiceAccountID(cr.Status.Account.ID),
	)
	if err != nil {
		return fmt.Errorf("listing tokens to get expirations: %w", err)
	}

	// Build a map of token ID to expiration time
	expirations := make(map[int64]*metav1.Time, len(listResp.Payload))
	for _, token := range listResp.Payload {
		expirations[token.ID] = convertGrafanaExpiration(token.Expiration)
	}

	// Update tokens in status with their expiration times
	for i := range cr.Status.Account.Tokens {
		cr.Status.Account.Tokens[i].Expires = expirations[cr.Status.Account.Tokens[i].ID]
	}

	return nil
}

// pruneAndIndexSecrets removes orphaned secrets (those whose tokens are no longer in the spec)
// and returns a map of remaining valid secrets indexed by token name for efficient lookup.
// It uses the token name annotation to match secrets with their corresponding tokens in the spec.
func (r *GrafanaServiceAccountReconciler) pruneAndIndexSecrets(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
) (map[string]corev1.Secret, error) {
	var secrets corev1.SecretList

	err := r.List(ctx, &secrets,
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(buildSecretLabels(cr)),
	)
	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}

	// Build map of desired tokens for efficient lookup
	desiredTokens := make(map[string]*v1beta1.GrafanaServiceAccountTokenSpec, len(cr.Spec.Tokens))
	for i, token := range cr.Spec.Tokens {
		desiredTokens[token.Name] = &cr.Spec.Tokens[i]
	}

	filtered := make(map[string]corev1.Secret)

	for _, secret := range secrets.Items {
		tokenName, ok := extractTokenNameFromSecret(&secret)
		if !ok {
			continue
		}

		token, ok := desiredTokens[tokenName]
		if ok {
			b, secretHasToken := secret.Data["token"]
			secretTokenNotEmpty := len(b) > 0
			secretNameIsValid := (token.SecretName == "" && secret.GenerateName != "") ||
				(token.SecretName != "" && secret.GenerateName == "" && token.SecretName == secret.Name)

			if secretHasToken && secretTokenNotEmpty && secretNameIsValid {
				// Keep this secret
				filtered[tokenName] = secret
				continue
			}
		}

		// Secret is not referenced or used by ServiceAccount
		logf.FromContext(ctx).Info("Deleting invalid or orphaned secret", "name", secret.Name, "namespace", secret.Namespace)

		err := r.Delete(ctx, &secret)
		if err != nil && !kuberr.IsNotFound(err) {
			return nil, fmt.Errorf("deleting invalid or orphaned secret %q: %w", secret.Name, err)
		}
	}

	return filtered, nil
}

// upsertAccount ensures a service account exists in Grafana.
func (r *GrafanaServiceAccountReconciler) upsertAccount(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	if cr.Status.Account != nil {
		update, err := gClient.ServiceAccounts.UpdateServiceAccount(
			service_accounts.
				NewUpdateServiceAccountParamsWithContext(ctx).
				WithServiceAccountID(cr.Status.Account.ID).
				WithBody(&models.UpdateServiceAccountForm{
					// The form contains a ServiceAccountID field which is unused in Grafana, so it's ignored here.
					Name:       cr.Spec.Name,
					Role:       cr.Spec.Role,
					IsDisabled: ptr.To(cr.Spec.IsDisabled),
				}),
		)
		if err == nil {
			cr.Status.Account = &v1beta1.GrafanaServiceAccountInfo{
				ID:         update.Payload.Serviceaccount.ID,
				Role:       update.Payload.Serviceaccount.Role,
				IsDisabled: update.Payload.Serviceaccount.IsDisabled,
				Name:       update.Payload.Serviceaccount.Name,
				Login:      update.Payload.Serviceaccount.Login,
			}

			// Load existing tokens from Grafana
			tokenList, err := gClient.ServiceAccounts.ListTokensWithParams(
				service_accounts.
					NewListTokensParamsWithContext(ctx).
					WithServiceAccountID(cr.Status.Account.ID),
			)
			if err != nil {
				return fmt.Errorf("listing tokens: %w", err)
			}

			cr.Status.Account.Tokens = make([]v1beta1.GrafanaServiceAccountTokenStatus, 0, len(tokenList.Payload))
			for _, token := range tokenList.Payload {
				if token != nil {
					cr.Status.Account.Tokens = append(cr.Status.Account.Tokens, v1beta1.GrafanaServiceAccountTokenStatus{
						ID:      token.ID,
						Name:    token.Name,
						Expires: convertGrafanaExpiration(token.Expiration),
					})
				}
			}

			return nil
		}

		// ATM, service_accounts.UpdateServiceAccountNotFound doesn't have Is, Unwrap, Unwrap.
		// So, we cannot rely only on errors.Is().
		_, ok := err.(*service_accounts.UpdateServiceAccountNotFound) // nolint:errorlint
		if !ok && !errors.Is(err, service_accounts.NewUpdateServiceAccountNotFound()) {
			return fmt.Errorf("updating service account: %w", err)
		}

		cr.Status.Account = nil
	}

	create, err := gClient.ServiceAccounts.CreateServiceAccount(
		service_accounts.
			NewCreateServiceAccountParamsWithContext(ctx).
			WithBody(&models.CreateServiceAccountForm{
				Name:       cr.Spec.Name,
				Role:       cr.Spec.Role,
				IsDisabled: cr.Spec.IsDisabled,
			}),
	)
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}

	cr.Status.Account = &v1beta1.GrafanaServiceAccountInfo{
		ID:         create.Payload.ID,
		Role:       create.Payload.Role,
		IsDisabled: create.Payload.IsDisabled,
		Name:       create.Payload.Name,
		Login:      create.Payload.Login,
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeAccountToken(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
	tokenStatus *v1beta1.GrafanaServiceAccountTokenStatus,
) error {
	if tokenStatus.Secret != nil {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
			Namespace: tokenStatus.Secret.Namespace,
			Name:      tokenStatus.Secret.Name,
		}}

		err := r.Delete(ctx, secret)
		if err != nil && !kuberr.IsNotFound(err) {
			return fmt.Errorf("deleting token secret %q: %w", secret.Name, err)
		}

		tokenStatus.Secret = nil
	}

	_, err := gClient.ServiceAccounts.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithTokenID(tokenStatus.ID),
	)
	if err != nil {
		// ATM, service_accounts.DeleteTokenNotFound doesn't have Is, Unwrap, Unwrap.
		// So, we cannot rely only on errors.Is().
		_, ok := err.(*service_accounts.DeleteTokenNotFound) // nolint:errorlint
		if ok || errors.Is(err, service_accounts.NewDeleteTokenNotFound()) {
			return nil
		}

		return fmt.Errorf("deleting token: %w", err)
	}

	return nil
}

func isEqualExpirationTime(a, b *metav1.Time) bool {
	// Token expiration drift tolerance
	const tokenExpirationDrift = 1 * time.Second

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Grafana API doesn't allow to set expiration time for tokens. Instead of it,
	// Grafana accepts TTL then calculates the expiration time against the current time.
	// So, we cannot just compare the expiration time with the spec' one.
	// Let's assume that two expiration times are equal if they are close enough.
	diff := a.Sub(b.Time)

	return diff.Abs() <= tokenExpirationDrift
}

func buildSecretLabels(cr *v1beta1.GrafanaServiceAccount) map[string]string {
	return map[string]string{
		"operator.grafana.com/service-account-instance": cr.Spec.InstanceName,
		"operator.grafana.com/service-account-name":     cr.Name,
		"operator.grafana.com/service-account-uid":      string(cr.UID),
	}
}

func generateSecretName(cr *v1beta1.GrafanaServiceAccount, tokenSpec v1beta1.GrafanaServiceAccountTokenSpec) string {
	return fmt.Sprintf("%s-%s-%s-", cr.Spec.InstanceName, cr.Spec.Name, tokenSpec.Name)
}

func extractTokenNameFromSecret(secret *corev1.Secret) (string, bool) {
	if secret.Annotations == nil {
		return "", false
	}

	tokenName, ok := secret.Annotations["operator.grafana.com/service-account-token-name"]

	return tokenName, ok
}

func buildTokenSecret(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
	tokenSpec v1beta1.GrafanaServiceAccountTokenSpec,
	tokenStatus v1beta1.GrafanaServiceAccountTokenStatus,
	tokenKey []byte,
	scheme *runtime.Scheme,
) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenSpec.SecretName,
			Namespace: cr.Namespace,
			Labels:    buildSecretLabels(cr),
			Annotations: map[string]string{
				"operator.grafana.com/service-account-spec-name":  cr.Spec.Name,
				"operator.grafana.com/service-account-uid":        string(cr.UID),
				"operator.grafana.com/service-account-token-name": tokenStatus.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
	}
	renewSecret(secret, tokenStatus, tokenKey)

	if secret.Name == "" {
		secret.GenerateName = generateSecretName(cr, tokenSpec)
	}

	if tokenStatus.Expires != nil {
		secret.Annotations["operator.grafana.com/service-account-token-expiry"] = tokenStatus.Expires.Format(time.RFC3339)
	}

	model2.SetInheritedLabels(secret, cr.Labels)

	if scheme != nil {
		err := controllerutil.SetControllerReference(cr, secret, scheme)
		if err != nil {
			logf.FromContext(ctx).Error(err, "Failed to set controller reference")
		}
	}

	return secret
}

func renewSecret(
	secret *corev1.Secret,
	tokenStatus v1beta1.GrafanaServiceAccountTokenStatus,
	tokenKey []byte,
) {
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	secret.Data["token"] = tokenKey

	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}

	secret.Annotations["operator.grafana.com/service-account-token-id"] = strconv.FormatInt(tokenStatus.ID, 10)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaServiceAccount{}, builder.WithPredicates(ignoreStatusUpdates())).
		Owns(&corev1.Secret{}).
		WithOptions(controller.Options{RateLimiter: defaultRateLimiter()}).
		Complete(r)
}
