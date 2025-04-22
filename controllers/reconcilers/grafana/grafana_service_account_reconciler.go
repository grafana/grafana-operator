package grafana

import (
	"context"
	"errors"

	// sha1 is used to generate a hash for the secret name.
	"crypto/sha1" // nolint:gosec
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/models"
	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	model2 "github.com/grafana/grafana-operator/v5/controllers/model"
)

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

func ptr[T any](v T) *T { return &v }

type GrafanaServiceAccountReconciler struct{ client client.Client }

func NewGrafanaServiceAccountReconciler(c client.Client) *GrafanaServiceAccountReconciler {
	return &GrafanaServiceAccountReconciler{client: c}
}

// Reconcile ensures service‑accounts/tokens match the spec and updates status.ServiceAccounts.
func (r *GrafanaServiceAccountReconciler) Reconcile(
	ctx context.Context,
	cr *v1beta1.Grafana,
	_ *v1beta1.OperatorReconcileVars,
	scheme *runtime.Scheme,
) (v1beta1.OperatorStageStatus, error) {
	ctx = logf.IntoContext(ctx, logf.FromContext(ctx).WithName("serviceAccountReconciler"))

	if cr.Spec.GrafanaServiceAccounts == nil {
		// No service accounts to reconcile.
		return v1beta1.OperatorStageResultSuccess, nil
	}

	gClient, err := client2.NewGeneratedGrafanaClient(ctx, r.client, cr)
	if err != nil {
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("building grafana client: %w", err)
	}

	err = r.reconcileAccounts(ctx, cr, gClient, scheme)
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *GrafanaServiceAccountReconciler) reconcileAccounts(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
	scheme *runtime.Scheme,
) error {
	allSecrets, err := r.listAllTokenSecrets(ctx, cr)
	if err != nil {
		return fmt.Errorf("listing all token secrets for instance %q: %w", cr.Name, err)
	}

	existingSAs, err := listServiceAccounts(ctx, gClient)
	if err != nil {
		return fmt.Errorf("listing service accounts: %w", err)
	}
	groups := r.classifyServiceAccounts(cr, existingSAs)

	for i, account := range groups.toDelete {
		err := r.removeAccount(ctx, cr, gClient, &groups.toDelete[i])
		if err != nil {
			// Let's actualize the status throwing away successfully removed accounts.
			groups.toDelete = groups.toDelete[i:]
			groups.consolidateTo(cr)
			return fmt.Errorf("removing service account %q: %w", account.SpecID, err)
		}
	}
	groups.toDelete = nil

	for _, spec := range groups.toCreate {
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
			groups.consolidateTo(cr)
			return fmt.Errorf("creating service account %q: %w", spec.ID, err)
		}
		groups.toSync = append(groups.toSync, accountSync{
			spec: spec,
			status: v1beta1.GrafanaServiceAccountStatus{
				SpecID:           spec.ID,
				ServiceAccountID: create.Payload.ID,
				Name:             create.Payload.Name,
				Role:             create.Payload.Role,
				IsDisabled:       create.Payload.IsDisabled,
			},
		})
	}

	for i := range groups.toSync {
		spec := groups.toSync[i].spec
		status := &groups.toSync[i].status
		existingSecrets := allSecrets[spec.ID]
		err := r.reconcileAccount(ctx, cr, gClient, spec, status, existingSecrets, scheme)
		if err != nil {
			groups.consolidateTo(cr)
			return fmt.Errorf("reconciling service account %q: %w", groups.toSync[i].status.SpecID, err)
		}
	}

	groups.consolidateTo(cr)
	return nil
}

func (r *GrafanaServiceAccountReconciler) removeAccount(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
	status *v1beta1.GrafanaServiceAccountStatus,
) error {
	// We don't need to remove tokens, because they will be removed automatically.
	// The only thing we need to do is to remove the secrets.
	err := r.wipeAccountSecrets(ctx, cr.Namespace, status)
	if err != nil {
		return fmt.Errorf("removing secrets for service account %q: %w", status.SpecID, err)
	}

	// Now we can remove the service account.
	_, err = gClient.ServiceAccounts.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(status.ServiceAccountID),
	)
	if err != nil {
		return fmt.Errorf("deleting service account %q: %w", status.SpecID, err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) reconcileAccount(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
	spec v1beta1.GrafanaServiceAccountSpec,
	status *v1beta1.GrafanaServiceAccountStatus,
	existingSecrets map[string]corev1.Secret,
	scheme *runtime.Scheme,
) error {
	var hasDiscrepancy bool
	form := models.UpdateServiceAccountForm{
		// I have no clue why a form has 'ServiceAccountID' field. We aleady have it as a parameter.
		// At the moment in Grafana it's not used at all.
		// So, let's just ignore it.
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
		update, err := gClient.ServiceAccounts.UpdateServiceAccount(
			service_accounts.
				NewUpdateServiceAccountParamsWithContext(ctx).
				WithServiceAccountID(status.ServiceAccountID).
				WithBody(&form),
		)
		if err != nil {
			return fmt.Errorf("updating service account %q: %w", status.SpecID, err)
		}
		status.IsDisabled = update.Payload.Serviceaccount.IsDisabled
		status.Role = update.Payload.Serviceaccount.Role
		status.Name = update.Payload.Serviceaccount.Name
	}

	err := r.reconcileTokens(ctx, cr, gClient, spec, status, existingSecrets, scheme)
	if err != nil {
		return fmt.Errorf("reconciling tokens for service account %s: %w", status.Name, err)
	}

	return nil
}

// reconcileTokens ensures tokens match the spec.
func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
	spec v1beta1.GrafanaServiceAccountSpec,
	sa *v1beta1.GrafanaServiceAccountStatus,
	existingSecrets map[string]corev1.Secret,
	scheme *runtime.Scheme,
) error {
	existingTokens, err := listServiceAccountTokens(ctx, gClient, sa.ServiceAccountID)
	if err != nil {
		return fmt.Errorf("listing tokens for service account: %w", err)
	}
	groups := r.classifyTokens(cr, existingTokens, existingSecrets, spec, *sa)

	for i, tokenStatus := range groups.toDelete {
		err := r.wipeToken(ctx, gClient, cr.Namespace, sa.ServiceAccountID, &groups.toDelete[i])
		if err != nil {
			groups.toDelete = groups.toDelete[i:]
			groups.consolidateTo(sa)
			return fmt.Errorf("removing token %q: %w", tokenStatus.Name, err)
		}
	}
	groups.toDelete = nil

	for _, tokenSpec := range groups.toCreate {
		// We need to remove the existing secret before creating a new token.
		secretName := generateTokenSecretName(cr.Name, sa.SpecID, tokenSpec.Name)
		secret, secretExists := existingSecrets[secretName]
		if secretExists {
			logf.FromContext(ctx).V(1).Info("stale secret was found, need to clean it up before creating a new token", "secret", secret.Name)
			err := r.wipeTokenSecret(ctx, cr.Namespace, &v1beta1.GrafanaServiceAccountTokenStatus{
				SecretName: secret.Name,
			})
			if err != nil {
				groups.consolidateTo(sa)
				return fmt.Errorf("removing token secret %q: %w", secret.Name, err)
			}
		}

		newToken, err := r.createToken(ctx, cr, gClient, *sa, tokenSpec, scheme)
		if err != nil {
			groups.consolidateTo(sa)
			return err
		}
		groups.toKeep = append(groups.toKeep, *newToken)
	}

	groups.consolidateTo(sa)

	return nil
}

// createToken creates a new token in Grafana and a secret for it.
// On error it returns a status reflecting all successful changes before the failure.
// In case of a success it returns the status of the created token.
func (r *GrafanaServiceAccountReconciler) createToken(
	ctx context.Context,
	cr *v1beta1.Grafana,
	gClient *genapi.GrafanaHTTPAPI,
	sa v1beta1.GrafanaServiceAccountStatus,
	spec v1beta1.GrafanaServiceAccountTokenSpec,
	scheme *runtime.Scheme,
) (*v1beta1.GrafanaServiceAccountTokenStatus, error) {
	cmd := models.AddServiceAccountTokenCommand{Name: spec.Name}
	if spec.Expires != nil {
		cmd.SecondsToLive = int64(time.Until(spec.Expires.Time).Seconds())
	}
	createResp, err := gClient.ServiceAccounts.CreateToken(
		service_accounts.
			NewCreateTokenParamsWithContext(ctx).
			WithServiceAccountID(sa.ServiceAccountID).
			WithBody(&cmd),
	)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}
	status := &v1beta1.GrafanaServiceAccountTokenStatus{
		Name:       createResp.Payload.Name,
		ID:         createResp.Payload.ID,
		SecretName: generateTokenSecretName(cr.Name, sa.SpecID, createResp.Payload.Name),
	}
	// Grafana doesn't return the expiration time in the response.
	// So, we need to do another request to get it.
	listResp, err := gClient.ServiceAccounts.ListTokensWithParams(
		service_accounts.
			NewListTokensParamsWithContext(ctx).
			WithServiceAccountID(sa.ServiceAccountID),
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
	secret := model2.GetInternalServiceAccountSecret(cr, sa, *status, tokenKey, scheme)
	err = r.client.Create(ctx, secret)
	if err != nil {
		return status, fmt.Errorf("creating token secret %q: %w", secret.Name, err)
	}
	status.SecretName = secret.Name

	return status, nil
}

// wipeToken removes the token from Grafana.
// Returns:
//   - nil if the token and related stuff was removed successfully
//   - error if removal has done partially and the status should be updated
func (r *GrafanaServiceAccountReconciler) wipeToken(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	namespace string,
	serviceAccountID int64,
	status *v1beta1.GrafanaServiceAccountTokenStatus,
) error {
	err := r.wipeTokenSecret(ctx, namespace, status)
	if err != nil {
		return fmt.Errorf("deleting token secret %q: %w", status.SecretName, err)
	}

	_, err = gClient.ServiceAccounts.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithTokenID(status.ID),
	)
	if err != nil {
		// ATM, service_accounts.DeleteTokenNotFound doesn't have Is, Unwrap, Unwrap.
		// So, we cannot rely only on errors.Is().
		_, ok := err.(*service_accounts.DeleteTokenNotFound) // nolint:errorlint
		if ok || errors.Is(err, service_accounts.NewDeleteTokenNotFound()) {
			logf.FromContext(ctx).V(1).Info("token not found, skipping", "token", status.Name)
			return nil
		}
		return fmt.Errorf("deleting token: %w", err)
	}

	return nil
}

// wipeAccountSecrets removes all secrets for specified tokens.
// It's not atomic. If an error occurs, it'll keep remain tokens in the status.
func (r *GrafanaServiceAccountReconciler) wipeAccountSecrets(
	ctx context.Context,
	namespace string,
	status *v1beta1.GrafanaServiceAccountStatus,
) error {
	i := len(status.Tokens) - 1
	for ; i >= 0; i-- {
		token := status.Tokens[i]
		err := r.wipeTokenSecret(ctx, namespace, &token)
		if err != nil {
			status.Tokens = status.Tokens[:i]
			return fmt.Errorf("removing token secret %q for service account: %w", token.SecretName, err)
		}
	}
	status.Tokens = nil
	return nil
}

func (r *GrafanaServiceAccountReconciler) wipeTokenSecret(
	ctx context.Context,
	namespace string,
	status *v1beta1.GrafanaServiceAccountTokenStatus,
) error {
	if status.SecretName == "" {
		// Nothing to remove.
		return nil
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: status.SecretName}}
	err := r.client.Delete(ctx, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			status.SecretName = ""
			logf.FromContext(ctx).V(1).Info("token secret not found, skipping", "secret", secret.Name)
			return nil
		}
		return err
	}
	status.SecretName = ""

	return nil
}

func generateTokenSecretName(grafanaName, serviceAccountSpecID, tokenName string) string {
	const maxSecretNameLength = 63

	sanitizeK8sName := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}

	base := strings.Join([]string{grafanaName, serviceAccountSpecID, tokenName}, "-")
	if len(base) <= maxSecretNameLength {
		return sanitizeK8sName(base)
	}

	prefixLen := maxSecretNameLength - 7
	if prefixLen < 1 {
		prefixLen = 1
	}
	prefix := base[:prefixLen]

	sum := sha1.Sum([]byte(base)) // nolint:gosec
	shortHash := fmt.Sprintf("%x", sum[:3])

	return sanitizeK8sName(fmt.Sprintf("%s-%s", prefix, shortHash))
}

type accountSync struct {
	spec   v1beta1.GrafanaServiceAccountSpec
	status v1beta1.GrafanaServiceAccountStatus
}

type saGroups struct {
	toDelete []v1beta1.GrafanaServiceAccountStatus
	toCreate []v1beta1.GrafanaServiceAccountSpec
	toSync   []accountSync
}

func (g *saGroups) consolidateTo(cr *v1beta1.Grafana) {
	cr.Status.GrafanaServiceAccounts = g.toDelete
	for _, r := range g.toSync {
		cr.Status.GrafanaServiceAccounts = append(cr.Status.GrafanaServiceAccounts, r.status)
	}
	// Sort the accounts to ensure a stable order.
	sort.Slice(cr.Status.GrafanaServiceAccounts, func(i, j int) bool {
		return cr.Status.GrafanaServiceAccounts[i].Name < cr.Status.GrafanaServiceAccounts[j].Name
	})
}

func (r *GrafanaServiceAccountReconciler) classifyServiceAccounts(
	cr *v1beta1.Grafana,
	existingSAs map[int64]models.ServiceAccountDTO,
) *saGroups {
	var g saGroups
	specMap := map[string]v1beta1.GrafanaServiceAccountSpec{}
	for _, spec := range cr.Spec.GrafanaServiceAccounts.Accounts {
		specMap[spec.ID] = spec
	}

	for _, sa := range cr.Status.GrafanaServiceAccounts {
		spec, ok := specMap[sa.SpecID]
		if !ok {
			g.toDelete = append(g.toDelete, sa)
			continue
		}
		delete(specMap, sa.SpecID)

		existingSA, ok := existingSAs[sa.ServiceAccountID]
		if !ok {
			g.toCreate = append(g.toCreate, spec)
			continue
		}
		sa = v1beta1.GrafanaServiceAccountStatus{
			SpecID:           spec.ID,
			ServiceAccountID: existingSA.ID,
			Name:             existingSA.Name,
			Role:             existingSA.Role,
			IsDisabled:       existingSA.IsDisabled,
			Tokens:           sa.Tokens,
		}

		g.toSync = append(g.toSync, accountSync{spec: spec, status: sa})
	}
	for _, spec := range specMap {
		g.toCreate = append(g.toCreate, spec)
	}
	// Sort the accounts to ensure a stable order.
	sort.Slice(g.toDelete, func(i, j int) bool { return g.toDelete[i].SpecID < g.toDelete[j].SpecID })
	sort.Slice(g.toCreate, func(i, j int) bool { return g.toCreate[i].ID < g.toCreate[j].ID })
	sort.Slice(g.toSync, func(i, j int) bool { return g.toSync[i].status.SpecID < g.toSync[j].status.SpecID })

	return &g
}

type tokenGroups struct {
	toDelete []v1beta1.GrafanaServiceAccountTokenStatus
	toCreate []v1beta1.GrafanaServiceAccountTokenSpec
	toKeep   []v1beta1.GrafanaServiceAccountTokenStatus
}

func (g *tokenGroups) consolidateTo(status *v1beta1.GrafanaServiceAccountStatus) {
	status.Tokens = g.toKeep
	status.Tokens = append(status.Tokens, g.toDelete...)
	sort.Slice(status.Tokens, func(i, j int) bool { return status.Tokens[i].Name < status.Tokens[j].Name })
}

func (r *GrafanaServiceAccountReconciler) classifyTokens(
	cr *v1beta1.Grafana,
	existingToken map[int64]models.TokenDTO,
	existingSecrets map[string]corev1.Secret,
	accountSpec v1beta1.GrafanaServiceAccountSpec,
	accountStatus v1beta1.GrafanaServiceAccountStatus,
) *tokenGroups {
	tokenSpecs := accountSpec.Tokens
	if len(tokenSpecs) == 0 && cr.Spec.GrafanaServiceAccounts.GenerateTokenSecret {
		tokenSpecs = []v1beta1.GrafanaServiceAccountTokenSpec{
			{Name: fmt.Sprintf("%s-%s-default-token", cr.Name, accountStatus.SpecID)},
		}
	}
	desired := map[string]v1beta1.GrafanaServiceAccountTokenSpec{}
	for _, tokenSpec := range tokenSpecs {
		desired[tokenSpec.Name] = tokenSpec
	}

	var g tokenGroups
	for _, tokenStatus := range accountStatus.Tokens {
		desiredToken, ok := desired[tokenStatus.Name]
		if !ok {
			g.toDelete = append(g.toDelete, tokenStatus)
			continue
		}
		delete(desired, tokenStatus.Name)

		existingToken, ok := existingToken[tokenStatus.ID]
		if !ok {
			g.toCreate = append(g.toCreate, desiredToken)
			continue
		}
		tokenStatus = v1beta1.GrafanaServiceAccountTokenStatus{
			Name:       existingToken.Name,
			ID:         existingToken.ID,
			SecretName: tokenStatus.SecretName,
		}
		if !existingToken.Expiration.IsZero() {
			tokenStatus.Expires = ptr(metav1.NewTime(time.Time(existingToken.Expiration)))
		}

		if (desiredToken.Expires != nil && tokenStatus.Expires == nil) ||
			(desiredToken.Expires == nil && tokenStatus.Expires != nil) ||
			!isEqualExpirationTime(desiredToken.Expires, tokenStatus.Expires) {
			g.toDelete = append(g.toDelete, tokenStatus)
			g.toCreate = append(g.toCreate, desiredToken)
			continue
		}

		_, secretExists := existingSecrets[tokenStatus.SecretName]
		if !secretExists {
			g.toDelete = append(g.toDelete, tokenStatus)
			g.toCreate = append(g.toCreate, desiredToken)
			continue
		}

		g.toKeep = append(g.toKeep, tokenStatus)
	}

	for _, tokenSpec := range desired {
		g.toCreate = append(g.toCreate, tokenSpec)
	}

	return &g
}

func listServiceAccounts(
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
				logf.FromContext(ctx).V(1).Info("service account is nil, skipping")
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

func listServiceAccountTokens(
	ctx context.Context,
	gClient *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
) (map[int64]models.TokenDTO, error) {
	list, err := gClient.ServiceAccounts.ListTokensWithParams(
		service_accounts.
			NewListTokensParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID),
	)
	if err != nil {
		return nil, fmt.Errorf("listing tokens for service account: %w", err)
	}
	existingTokens := map[int64]models.TokenDTO{}
	for _, token := range list.Payload {
		if token == nil {
			logf.FromContext(ctx).V(1).Info("token is nil, skipping")
			continue
		}
		existingTokens[token.ID] = *token
	}
	return existingTokens, nil
}

func (r *GrafanaServiceAccountReconciler) listAllTokenSecrets(
	ctx context.Context,
	cr *v1beta1.Grafana,
) (map[string]map[string]corev1.Secret, error) {
	var list corev1.SecretList
	err := r.client.List(ctx, &list,
		client.InNamespace(cr.Namespace),
		client.MatchingLabels{
			"app":                              "grafana-serviceaccount-token",
			"grafana.integreatly.org/instance": cr.Name,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("listing token secrets: %w", err)
	}

	res := map[string]map[string]corev1.Secret{}
	for _, s := range list.Items {
		specID := model2.GetInternalServiceAccountSpecIDFromSecret(s)
		if specID == "" {
			logf.FromContext(ctx).V(1).Info("secret doesn't have a spec ID, skipping", "secret", s.Name)
			continue
		}
		if res[specID] == nil {
			res[specID] = map[string]corev1.Secret{}
		}
		res[specID][s.Name] = s
	}
	return res, nil
}
