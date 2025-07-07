package controllers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	model2 "github.com/grafana/grafana-operator/v5/controllers/model"
)

const conditionServiceAccountsSynchronized = "ServiceAccountsSynchronized"

func setFailedServiceAccountsCondition(cr *v1beta1.Grafana, message string) {
	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:               conditionServiceAccountsSynchronized,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Status:             metav1.ConditionFalse,
		Reason:             conditionApplyFailed,
		Message:            message,
	})
}

func setSuccessfulServiceAccountsCondition(cr *v1beta1.Grafana, message string) {
	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:               conditionServiceAccountsSynchronized,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Status:             metav1.ConditionTrue,
		Reason:             conditionApplySuccessful,
		Message:            message,
	})
}

type GrafanaServiceAccountReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func newGrafanaServiceAccountReconciler(client client.Client, scheme *runtime.Scheme) *GrafanaServiceAccountReconciler {
	return &GrafanaServiceAccountReconciler{
		Client: client,
		scheme: scheme,
	}
}

func (r *GrafanaServiceAccountReconciler) reconcile(ctx context.Context, cr *v1beta1.Grafana) error {
	if cr.Spec.GrafanaServiceAccounts == nil && cr.Status.GrafanaServiceAccounts == nil {
		return nil
	}

	gClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, cr)
	if err != nil {
		setFailedServiceAccountsCondition(cr, err.Error())
		return fmt.Errorf("building grafana client: %w", err)
	}

	err = r.reconcileAccounts(ctx, cr, gClient)
	if err != nil {
		setFailedServiceAccountsCondition(cr, err.Error())
		return fmt.Errorf("reconciling service accounts: %w", err)
	}
	setSuccessfulServiceAccountsCondition(cr, "service accounts reconciled")

	if cr.Spec.GrafanaServiceAccounts == nil {
		// Spec is empty, so we don't need to check periodically the service accounts status.
		return nil
	}

	return nil
}

// syncAccounts checks if the service accounts status in the Grafana CR is up to date
// with the actual service accounts in Grafana. If there are any discrepancies, it updates the status
// accordingly. If a service account was removed from Grafana, it removes it from the status.
func (r *GrafanaServiceAccountReconciler) syncAccounts(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
) error {
	ctx = logf.IntoContext(ctx, logf.FromContext(ctx).WithName("GrafanaServiceAccountController"))

	if len(cr.Status.GrafanaServiceAccounts) == 0 {
		return nil
	}

	existingAccounts, err := listExistingServiceAccounts(ctx, gClient)
	if err != nil {
		return fmt.Errorf("listing service accounts: %w", err)
	}

	removed := 0
	for i := 0; i < len(cr.Status.GrafanaServiceAccounts); i++ {
		existingAccount, ok := existingAccounts[cr.Status.GrafanaServiceAccounts[i].ServiceAccountID]
		if !ok {
			// It seems that the service account was removed from Grafana. Let's remove it from the status.
			cr.Status.GrafanaServiceAccounts = removeFromSlice(cr.Status.GrafanaServiceAccounts, i)
			removed++
			i--
			continue
		}

		actualizeAccountStatus(&cr.Status.GrafanaServiceAccounts[i], existingAccount)
		err := r.syncTokens(ctx, gClient, &cr.Status.GrafanaServiceAccounts[i])
		if err != nil {
			return fmt.Errorf("syncing tokens for service account %q: %w", cr.Status.GrafanaServiceAccounts[i].SpecID, err)
		}
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) syncTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	accountStatus *v1beta1.GrafanaServiceAccountStatus,
) error {
	tokens, err := listExistingTokens(ctx, gClient, accountStatus.ServiceAccountID)
	if err != nil {
		return fmt.Errorf("listing tokens for service account %q: %w", accountStatus.ServiceAccountID, err)
	}

	for i := 0; i < len(accountStatus.Tokens); i++ {
		existingToken, ok := tokens[accountStatus.Tokens[i].ID]
		if !ok {
			// It seems that the service account token was removed from Grafana. Let's remove it from the status.
			accountStatus.Tokens = removeFromSlice(accountStatus.Tokens, i)
			i--
			continue
		}

		actualizeTokenStatus(&accountStatus.Tokens[i], existingToken)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) reconcileAccounts(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
) error {
	// Sort the accounts to ensure a stable order.
	defer func(cr *v1beta1.Grafana) {
		sort.Slice(cr.Status.GrafanaServiceAccounts, func(i, j int) bool {
			return cr.Status.GrafanaServiceAccounts[i].SpecID < cr.Status.GrafanaServiceAccounts[j].SpecID
		})
	}(cr)

	err := r.syncAccounts(ctx, cr, gClient)
	if err != nil {
		return fmt.Errorf("syncing GrafanaServiceAccounts status: %w", err)
	}

	specMap := map[string]v1beta1.GrafanaServiceAccountSpec{}
	if cr.Spec.GrafanaServiceAccounts != nil {
		for _, spec := range cr.Spec.GrafanaServiceAccounts.Accounts {
			specMap[spec.ID] = spec
		}
	}

	// What we want to do is:
	// 1. Iterate over the existing service accounts in the status.
	// 2. If a service account is not in the spec anymore, remove it.
	// 3. If a service account is still in the spec, reconcile it.
	// 4. Create new service accounts for any remaining specs that aren't in the status.

	// Let's iterate over the existing service accounts in the status.
	for i := 0; i < len(cr.Status.GrafanaServiceAccounts); i++ {
		spec, ok := specMap[cr.Status.GrafanaServiceAccounts[i].SpecID]
		if !ok {
			// It's not in the spec anymore, so we need to remove it.
			err := r.removeAccount(ctx, gClient, &cr.Status.GrafanaServiceAccounts[i])
			if err != nil {
				return fmt.Errorf("removing service account %q: %w", cr.Status.GrafanaServiceAccounts[i].SpecID, err)
			}
			cr.Status.GrafanaServiceAccounts = removeFromSlice(cr.Status.GrafanaServiceAccounts, i)
			i--
			continue
		}

		// The service account is still in the spec, so we need to reconcile it.
		delete(specMap, cr.Status.GrafanaServiceAccounts[i].SpecID)
		err := r.reconcileAccount(ctx, gClient, cr, spec, &cr.Status.GrafanaServiceAccounts[i])
		if err != nil {
			return fmt.Errorf("reconciling service account %q: %w", cr.Status.GrafanaServiceAccounts[i].SpecID, err)
		}
	}

	// We assume that specMap now contains only the service accounts that need to be created.
	for _, spec := range specMap {
		newAccount, err := r.createAccount(ctx, gClient, cr, spec)
		if newAccount != nil {
			cr.Status.GrafanaServiceAccounts = append(cr.Status.GrafanaServiceAccounts, *newAccount)
		}
		if err != nil {
			return fmt.Errorf("creating service account %q: %w", spec.ID, err)
		}
	}

	return nil
}

// createAccount creates a new service account in Grafana and all related resources such as tokens and secrets.
// This operation isn't atomic, so can succeed partially. Always check not only the error, but also the returned status.
// If the service account was created successfully, it will be returned in the status.
func (r *GrafanaServiceAccountReconciler) createAccount(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.Grafana,
	spec v1beta1.GrafanaServiceAccountSpec,
) (*v1beta1.GrafanaServiceAccountStatus, error) {
	create, err := gClient.ServiceAccounts.CreateServiceAccount(
		service_accounts.
			NewCreateServiceAccountParamsWithContext(ctx).
			WithBody(&models.CreateServiceAccountForm{
				Name:       spec.Name,
				Role:       spec.Role,
				IsDisabled: spec.IsDisabled,
			}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating service account %q: %w", spec.ID, err)
	}

	newAccount := v1beta1.GrafanaServiceAccountStatus{
		SpecID:           spec.ID,
		ServiceAccountID: create.Payload.ID,
		Name:             create.Payload.Name,
		Role:             create.Payload.Role,
		IsDisabled:       create.Payload.IsDisabled,
	}

	err = r.reconcileTokens(ctx, gClient, cr, spec, &newAccount)
	if err != nil {
		return &newAccount, fmt.Errorf("reconciling service account tokens for %q: %w", newAccount.SpecID, err)
	}

	return &newAccount, nil
}

func (r *GrafanaServiceAccountReconciler) removeAccount(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	status *v1beta1.GrafanaServiceAccountStatus,
) error {
	// We don't need to remove tokens, because they will be removed automatically.
	// The only thing we need to do is to remove the secrets.
	{
		i := len(status.Tokens) - 1
		for ; i >= 0; i-- {
			err := r.removeTokenSecret(ctx, &status.Tokens[i])
			if err != nil {
				status.Tokens = status.Tokens[:i+1]
				return fmt.Errorf("removing token secret %q for service account %q: %w", status.Tokens[i].Name, status.SpecID, err)
			}
		}
		status.Tokens = nil
	}

	_, err := gClient.ServiceAccounts.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(status.ServiceAccountID),
	)
	if err != nil {
		// ATM, service_accounts.DeleteServiceAccountNotFound doesn't have Is, Unwrap, Unwrap.
		// So, we cannot rely only on errors.Is().
		_, ok := err.(*service_accounts.DeleteServiceAccountNotFound) // nolint:errorlint
		if ok || errors.Is(err, service_accounts.NewDeleteServiceAccountNotFound()) {
			logf.FromContext(ctx).Info("service account not found, skipping removal",
				"serviceAccountID", status.ServiceAccountID,
				"specID", status.SpecID,
			)
			return nil
		}

		// TODO: Grafana Operator currently deployes Grafana 11.3.0 (see controllers/config/operator_constants.go#L6).
		// Till Grafana 12.0.2 there was no reliable way to detect a 404 when deleting a service account.
		// The API still returns 500 (see grafana/grafana#106618).
		//
		// Once we upgrade to Grafana > 12.0.2 and bump github.com/grafana/grafana-openapi-client-go,
		// we can handle the real 404 explicitly.
		//
		// In the meantime we treat any non-nil error from the delete call as "already removed" and just log it for visibility.
		logf.FromContext(ctx).Error(err, "failed to delete service account",
			"serviceAccountID", status.ServiceAccountID,
			"specID", status.SpecID,
		)
		// return fmt.Errorf("deleting service account %q: %w", status.SpecID, err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) reconcileAccount(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.Grafana,
	spec v1beta1.GrafanaServiceAccountSpec,
	status *v1beta1.GrafanaServiceAccountStatus,
) error {
	err := r.reconcileTokens(ctx, gClient, cr, spec, status)
	if err != nil {
		return fmt.Errorf("reconciling service account tokens for %q: %w", status.SpecID, err)
	}

	form := patchAccount(spec, *status)
	if form == nil {
		return nil
	}

	update, err := gClient.ServiceAccounts.UpdateServiceAccount(
		service_accounts.
			NewUpdateServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(status.ServiceAccountID).
			WithBody(form),
	)
	if err != nil {
		return fmt.Errorf("updating service account %q: %w", status.SpecID, err)
	}
	status.IsDisabled = update.Payload.Serviceaccount.IsDisabled
	status.Role = update.Payload.Serviceaccount.Role
	status.Name = update.Payload.Serviceaccount.Name

	return nil
}

func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.Grafana,
	accountSpec v1beta1.GrafanaServiceAccountSpec,
	accountStatus *v1beta1.GrafanaServiceAccountStatus,
) error {
	defer func() {
		sort.Slice(accountStatus.Tokens, func(i, j int) bool {
			return accountStatus.Tokens[i].Name < accountStatus.Tokens[j].Name
		})
	}()

	tokenSpecs := accountSpec.Tokens
	if len(tokenSpecs) == 0 && cr.Spec.GrafanaServiceAccounts.GenerateTokenSecret {
		// If there are no tokens in the spec, we create a default token.
		tokenSpecs = []v1beta1.GrafanaServiceAccountTokenSpec{
			{Name: fmt.Sprintf("%s-%s-default-token", cr.Name, accountStatus.SpecID)},
		}
	}

	specMap := map[string]v1beta1.GrafanaServiceAccountTokenSpec{}
	for _, tokenSpec := range tokenSpecs {
		specMap[tokenSpec.Name] = tokenSpec
	}

	for i := 0; i < len(accountStatus.Tokens); i++ {
		tokenSpec, ok := specMap[accountStatus.Tokens[i].Name]
		if !ok ||
			(tokenSpec.Expires != nil && accountStatus.Tokens[i].Expires == nil) ||
			(tokenSpec.Expires == nil && accountStatus.Tokens[i].Expires != nil) ||
			(tokenSpec.Expires != nil && accountStatus.Tokens[i].Expires != nil &&
				!isEqualExpirationTime(tokenSpec.Expires, accountStatus.Tokens[i].Expires)) {
			err := r.removeAccountToken(ctx, gClient, accountStatus, &accountStatus.Tokens[i])
			if err != nil {
				return fmt.Errorf("removing service account token %q: %w", accountStatus.Tokens[i].Name, err)
			}
			accountStatus.Tokens = removeFromSlice(accountStatus.Tokens, i)
			i--
			continue
		}

		delete(specMap, accountStatus.Tokens[i].Name)
	}

	for _, tokenSpec := range specMap {
		newToken, err := r.createAccountToken(ctx, gClient, cr, *accountStatus, tokenSpec)
		if newToken != nil {
			accountStatus.Tokens = append(accountStatus.Tokens, *newToken)
		}
		if err != nil {
			return fmt.Errorf("creating service account token %q: %w", tokenSpec.Name, err)
		}
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) createAccountToken(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.Grafana,
	accountStatus v1beta1.GrafanaServiceAccountStatus,
	tokenSpec v1beta1.GrafanaServiceAccountTokenSpec,
) (*v1beta1.GrafanaServiceAccountTokenStatus, error) {
	cmd := models.AddServiceAccountTokenCommand{Name: tokenSpec.Name}
	if tokenSpec.Expires != nil {
		cmd.SecondsToLive = int64(time.Until(tokenSpec.Expires.Time).Seconds())
	}
	createResp, err := gClient.ServiceAccounts.CreateToken(
		service_accounts.
			NewCreateTokenParamsWithContext(ctx).
			WithServiceAccountID(accountStatus.ServiceAccountID).
			WithBody(&cmd),
	)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}
	status := &v1beta1.GrafanaServiceAccountTokenStatus{
		Name: createResp.Payload.Name,
		ID:   createResp.Payload.ID,
	}

	// Grafana doesn't return the expiration time in the response.
	// So, we need to do another request to get it.
	listResp, err := gClient.ServiceAccounts.ListTokensWithParams(
		service_accounts.
			NewListTokensParamsWithContext(ctx).
			WithServiceAccountID(accountStatus.ServiceAccountID),
	)
	if err != nil {
		return status, fmt.Errorf("listing tokens: %w", err)
	}
	var found bool
	for _, token := range listResp.Payload {
		if token.ID == createResp.Payload.ID {
			if !token.Expiration.IsZero() {
				status.Expires = ptr(metav1.NewTime(time.Time(token.Expiration)))
			}
			found = true
			break
		}
	}
	if !found {
		return status, fmt.Errorf("token %q not found in the list", createResp.Payload.ID)
	}

	// The token was created, let's create a secret for it.
	tokenKey := []byte(createResp.Payload.Key)
	secret := model2.GetInternalServiceAccountSecret(cr, accountStatus, *status, tokenKey, r.Scheme())
	err = r.Create(ctx, secret)
	if err != nil {
		return status, fmt.Errorf("creating token secret %q: %w", secret.Name, err)
	}
	status.Secret = &v1beta1.GrafanaServiceAccountSecretStatus{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	return status, nil
}

func (r *GrafanaServiceAccountReconciler) removeTokenSecret(
	ctx context.Context,
	tokenStatus *v1beta1.GrafanaServiceAccountTokenStatus,
) error {
	if tokenStatus.Secret == nil {
		// Nothing to remove.
		return nil
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Namespace: tokenStatus.Secret.Namespace,
		Name:      tokenStatus.Secret.Name,
	}}
	err := r.Delete(ctx, secret)
	if err != nil {
		if kuberr.IsNotFound(err) {
			tokenStatus.Secret = nil
			return nil
		}
		return err
	}
	tokenStatus.Secret = nil

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeAccountToken(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	accountStatus *v1beta1.GrafanaServiceAccountStatus,
	tokenStatus *v1beta1.GrafanaServiceAccountTokenStatus,
) error {
	err := r.removeTokenSecret(ctx, tokenStatus)
	if err != nil {
		return err
	}

	_, err = gClient.ServiceAccounts.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(accountStatus.ServiceAccountID).
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

func actualizeAccountStatus(
	status *v1beta1.GrafanaServiceAccountStatus,
	actual models.ServiceAccountDTO,
) {
	status.Name = actual.Name
	status.Role = actual.Role
	status.IsDisabled = actual.IsDisabled
}

func actualizeTokenStatus(
	status *v1beta1.GrafanaServiceAccountTokenStatus,
	actual models.TokenDTO,
) {
	status.Name = actual.Name
	if actual.Expiration.IsZero() {
		status.Expires = nil
	} else {
		status.Expires = ptr(metav1.NewTime(time.Time(actual.Expiration)))
	}
}

func patchAccount(
	spec v1beta1.GrafanaServiceAccountSpec,
	status v1beta1.GrafanaServiceAccountStatus,
) *models.UpdateServiceAccountForm {
	var hasDiscrepancy bool
	form := models.UpdateServiceAccountForm{
		// The form contains a ServiceAccountID field which is unused in Grafana, so it's ignored here.
		// ServiceAccountID: status.ServiceAccountID,
	}
	if status.Name != spec.Name {
		hasDiscrepancy = true
		form.Name = spec.Name
	}
	if status.Role != spec.Role {
		hasDiscrepancy = true
		form.Role = spec.Role
	}
	if status.IsDisabled != spec.IsDisabled {
		hasDiscrepancy = true
		form.IsDisabled = ptr(spec.IsDisabled)
	}

	if hasDiscrepancy {
		return &form
	}
	return nil
}

func listExistingServiceAccounts(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
) (map[int64]models.ServiceAccountDTO, error) {
	serviceAccounts := map[int64]models.ServiceAccountDTO{}

	var page int64 = 1
	for {
		resp, err := gClient.ServiceAccounts.SearchOrgServiceAccountsWithPaging(
			service_accounts.
				NewSearchOrgServiceAccountsWithPagingParamsWithContext(ctx).
				WithPage(&page),
		)
		if err != nil {
			return nil, fmt.Errorf("searching service accounts: %w", err)
		}

		for _, sa := range resp.Payload.ServiceAccounts {
			if sa == nil {
				continue
			}
			serviceAccounts[sa.ID] = *sa
		}

		if resp.Payload.TotalCount <= int64(len(serviceAccounts)) {
			return serviceAccounts, nil
		}
		page++
	}
}

func listExistingTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
) (map[int64]models.TokenDTO, error) {
	resp, err := gClient.ServiceAccounts.ListTokensWithParams(
		service_accounts.
			NewListTokensParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID),
	)
	if err != nil {
		return nil, fmt.Errorf("listing tokens: %w", err)
	}

	tokens := map[int64]models.TokenDTO{}
	for _, token := range resp.Payload {
		if token == nil {
			continue
		}
		tokens[token.ID] = *token
	}

	return tokens, nil
}

func ptr[T any](v T) *T { return &v }

func removeFromSlice[T any](slice []T, idx int) []T {
	// Keep order stable by using slices.Delete which shifts remaining elements left.
	// The alternative would be swapping with the last element for O(1) removal.
	return slices.Delete(slice, idx, idx+1)
}

func isEqualExpirationTime(a, b *metav1.Time) bool {
	// Grafana API doesn't allow to set expiration time for tokens. Instead of it,
	// Grafana accepts TTL then calculates the expiration time against the current time.
	// So, we cannot just compare the expiration time with the spec' one.
	// Let's assume that two expiration times are equal if they are close enough.
	const expiresDrift = 1 * time.Second

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	diff := a.Sub(b.Time)
	return diff.Abs() <= expiresDrift
}
