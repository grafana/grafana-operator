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
	"crypto/sha1"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/access_control"
	"github.com/grafana/grafana-openapi-client-go/client/health"
	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/client/teams"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	v1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"

	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

// Reconcile implements the main reconciliation loop.
func (r *GrafanaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler")
	ctx = logf.IntoContext(ctx, log)

	log.V(1).Info("Reconciling GrafanaServiceAccount started")
	defer log.V(1).Info("Reconciling GrafanaServiceAccount finished")

	cr := &v1beta1.GrafanaServiceAccount{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			log.V(1).Info("GrafanaServiceAccount not found, skipping")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting GrafanaServiceAccount %q: %w", req, err)
	}

	if cr.DeletionTimestamp != nil {
		log.V(1).Info("GrafanaServiceAccount is being deleted")
		err := r.finalize(ctx, cr)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("running finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// At the end, update status and set/unset finalizer if needed
	defer r.updateStatusAndFinalizer(ctx, cr)

	// Find matching Grafana instances
	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		return ctrl.Result{}, fmt.Errorf("fetching Grafana instances: %w", err)
	}
	if len(grafanas) == 0 {
		// No matching Grafana, set a condition and requeue
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, nil)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}
	removeNoMatchingInstance(&cr.Status.Conditions)
	removeInvalidSpec(&cr.Status.Conditions)
	log.V(1).Info("found matching Grafana instances", "count", len(grafanas))

	// Reconcile service accounts for each matching Grafana
	applyErrors := map[string]string{}
	for _, grafana := range grafanas {
		err := r.reconcileGrafana(ctx, grafana, cr)
		if err == nil {
			continue
		}

		switch {
		case errors.Is(err, &service_accounts.CreateTokenBadRequest{}):
			setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidSpec", err.Error())
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		case kuberr.IsAlreadyExists(err):
			setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidSpec", err.Error())
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		case kuberr.IsConflict(err):
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:               conditionServiceAccountSynchronized,
				Status:             metav1.ConditionFalse,
				Reason:             "Conflict",
				Message:            err.Error(),
				ObservedGeneration: cr.GetGeneration(),
			})
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
		}

		applyErrors[strings.Join([]string{grafana.Namespace, grafana.Name}, "/")] = fmt.Sprintf("reconciling Grafana: %v", err)
	}

	meta.SetStatusCondition(&cr.Status.Conditions, buildSynchronizedCondition(
		"ServiceAccount",
		conditionServiceAccountSynchronized,
		cr.Generation,
		applyErrors,
		len(grafanas),
	))

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("applying changes to some instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: cr.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaServiceAccountReconciler) reconcileGrafana(
	ctx context.Context,
	grafana v1beta1.Grafana,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
	if err != nil {
		return fmt.Errorf("creating Grafana client: %w", err)
	}

	resp, err := grafanaClient.Health.GetHealthWithParams(health.NewGetHealthParamsWithContext(ctx))
	if err != nil {
		return fmt.Errorf("getting Grafana health: %w", err)
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("Grafana is not healthy: %w", resp)
	}

	// Reconcile one instance record in cr.Status.Instances
	grafanaRef, err := r.reconcileGrafanaInstance(ctx, grafanaClient, cr, &grafana)
	if err != nil {
		return fmt.Errorf("reconciling Grafana instance: %w", err)
	}

	// Update the Grafana status.ServiceAccounts (for housekeeping)
	saUID := strconv.FormatInt(grafanaRef.ServiceAccountID, 10)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1beta1.Grafana{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: grafana.Namespace, Name: grafana.Name}, latest)
		if err != nil {
			return fmt.Errorf("getting Grafana %s/%s: %w", grafana.Namespace, grafana.Name, err)
		}
		latest.Status.ServiceAccounts = latest.Status.ServiceAccounts.Add(cr.Namespace, cr.Name, saUID)
		return r.Client.Status().Update(ctx, latest)
	})
	if err != nil {
		return fmt.Errorf("updating Grafana status: %w", err)
	}

	return nil
}

// reconcileGrafanaInstance ensures that for a given Grafana instance, we have a record in cr.Status.Instances,
// sets up (create/update) the service account and tokens, and returns the updated instance record.
func (r *GrafanaServiceAccountReconciler) reconcileGrafanaInstance(
	ctx context.Context,
	grafanaClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
	grafana *v1beta1.Grafana,
) (*v1beta1.GrafanaServiceAccountInstanceStatus, error) {
	idx := slices.IndexFunc(cr.Status.Instances, func(si v1beta1.GrafanaServiceAccountInstanceStatus) bool {
		return si.GrafanaNamespace == grafana.Namespace && si.GrafanaName == grafana.Name
	})
	if idx == -1 {
		cr.Status.Instances = append(cr.Status.Instances, v1beta1.GrafanaServiceAccountInstanceStatus{
			GrafanaNamespace: grafana.Namespace,
			GrafanaName:      grafana.Name,
			ServiceAccountID: 0,
			Tokens:           nil,
		})
		idx = len(cr.Status.Instances) - 1
	}
	instance := &cr.Status.Instances[idx]

	search, err := grafanaClient.ServiceAccounts.SearchOrgServiceAccountsWithPaging(
		service_accounts.NewSearchOrgServiceAccountsWithPagingParamsWithContext(ctx).
			WithQuery(&cr.Spec.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("searching service accounts: %w", err)
	}
	for _, sa := range search.Payload.ServiceAccounts {
		if sa.Name == cr.Spec.Name {
			if instance.ServiceAccountID == 0 || instance.ServiceAccountID == sa.ID {
				instance.ServiceAccountID = sa.ID
				break
			}

			return nil, kuberr.NewConflict(
				schema.GroupResource{
					Group:    v1beta1.GroupVersion.Group,
					Resource: "GrafanaServiceAccount",
				},
				"GrafanaServiceAccount",
				fmt.Errorf("Grafana has service account name=%q ID=%d, but instance record ID=%d", sa.Name, sa.ID, instance.ServiceAccountID),
			)
		}
	}

	if instance.ServiceAccountID == 0 {
		resp, err := grafanaClient.ServiceAccounts.CreateServiceAccount(
			service_accounts.NewCreateServiceAccountParamsWithContext(ctx).
				WithBody(&models.CreateServiceAccountForm{
					Name:       cr.Spec.Name,
					Role:       cr.Spec.Role,
					IsDisabled: cr.Spec.IsDisabled,
				}),
		)
		if err != nil {
			return instance, fmt.Errorf("creating service account: %w", err)
		}
		instance.ServiceAccountID = resp.Payload.ID
	} else {
		_, err := grafanaClient.ServiceAccounts.UpdateServiceAccount( // nolint:errcheck
			service_accounts.NewUpdateServiceAccountParamsWithContext(ctx).
				WithBody(&models.UpdateServiceAccountForm{
					Name:             cr.Spec.Name,
					Role:             cr.Spec.Role,
					IsDisabled:       cr.Spec.IsDisabled,
					ServiceAccountID: instance.ServiceAccountID,
				}).
				WithServiceAccountID(instance.ServiceAccountID),
		)
		if err != nil {
			return instance, fmt.Errorf("update service account: %w", err)
		}
	}

	err = r.reconcileTokens(ctx, grafanaClient, cr, instance)
	if err != nil {
		return instance, fmt.Errorf("reconcile tokens: %w", err)
	}
	err = r.reconcilePermissionsForInstance(ctx, grafanaClient, cr, instance)
	if err != nil {
		return instance, fmt.Errorf("reconcile permissions: %w", err)
	}

	return instance, nil
}

// finalize is called when the CR is being deleted. We remove service accounts and secrets
// from all instances, then remove finalizer.
func (r *GrafanaServiceAccountReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) error {
	if !controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
		return nil
	}
	log := logf.FromContext(ctx)

	// For each instance in .status.instances, remove tokens, remove the service account from Grafana,
	// and remove reference from grafana.Status.ServiceAccounts
	for _, instance := range cr.Status.Instances {
		log := log.WithValues(
			"grafanaNamespace", instance.GrafanaNamespace,
			"grafanaName", instance.GrafanaName,
		)

		// Check if the corresponding Grafana object is still around
		grafana := &v1beta1.Grafana{}
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: instance.GrafanaNamespace,
			Name:      instance.GrafanaName,
		}, grafana)
		if err != nil {
			log.Error(err, "unable to find Grafana instance for finalization")
			continue
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			log.Error(err, "unable to create Grafana client for finalization")
			continue
		}

		for _, token := range instance.Tokens {
			err := r.revokeToken(ctx, grafanaClient, instance.ServiceAccountID, token.TokenID)
			if err != nil {
				log.Error(err, "failed to revoke token", "tokenID", token.TokenID)
			}
			err = r.revokeSecret(ctx, cr.Namespace, token.SecretName)
			if err != nil {
				log.Error(err, "failed to remove token secret", "secretName", token.SecretName)
			}
		}

		if instance.ServiceAccountID != 0 {
			_, err := grafanaClient.ServiceAccounts.DeleteServiceAccountWithParams( // nolint:errcheck
				service_accounts.NewDeleteServiceAccountParamsWithContext(ctx).
					WithServiceAccountID(instance.ServiceAccountID),
			)
			if err != nil {
				logf.FromContext(ctx).Error(err, "failed to delete service account from Grafana", "serviceAccountID", instance.ServiceAccountID)
				return err
			}
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latest := &v1beta1.Grafana{}
			err := r.Client.Get(ctx, types.NamespacedName{
				Namespace: grafana.Namespace,
				Name:      grafana.Name,
			}, latest)
			if err != nil {
				return err
			}
			latest.Status.ServiceAccounts = latest.Status.ServiceAccounts.Remove(cr.Namespace, cr.Name)
			return r.Client.Status().Update(ctx, latest)
		})
		if err != nil {
			log.Error(err, "failed to update Grafana status during finalization", "grafana", grafana.Name)
		}
	}

	err := removeFinalizer(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("removing finalizer %s/%s: %w", cr.Namespace, cr.Name, err)
	}

	return nil
}

func generateTokenSecretName(crName, grafanaName, tokenName string) string {
	const maxSecretNameLength = 63

	sanitizeK8sName := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}

	base := fmt.Sprintf("%s-%s-%s", crName, grafanaName, tokenName)
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

// reconcileTokens cleans up expired or removed tokens, then creates missing ones.
func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	grafanaClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
	instanceStatus *v1beta1.GrafanaServiceAccountInstanceStatus,
) error {
	if cr.Spec.GenerateTokenSecret && len(cr.Spec.Tokens) == 0 {
		cr.Spec.Tokens = []v1beta1.GrafanaServiceAccountToken{{
			Name: fmt.Sprintf("%s-default-token", cr.Name),
		}}
	}

	// Remove tokens that are no longer valid or in spec
	now := time.Now()

	desiredTokens := make(map[string]*metav1.Time, len(cr.Spec.Tokens))
	for _, t := range cr.Spec.Tokens {
		desiredTokens[t.Name] = t.Expires
	}

	var kept []v1beta1.GrafanaServiceAccountTokenStatus
	for _, existing := range instanceStatus.Tokens {
		expires, found := desiredTokens[existing.Name]
		if found && (expires == nil || now.Before(expires.Time)) {
			kept = append(kept, existing)
			continue
		}

		err := r.revokeToken(ctx, grafanaClient, instanceStatus.ServiceAccountID, existing.TokenID)
		if err != nil {
			return fmt.Errorf("revoking token %d: %w", existing.TokenID, err)
		}
		err = r.revokeSecret(ctx, cr.Namespace, existing.SecretName)
		if err != nil {
			return fmt.Errorf("removing token secret %s: %w", existing.SecretName, err)
		}
	}
	instanceStatus.Tokens = kept

	// Create new ones if needed
	existingSecrets := make(map[string]struct{}, len(instanceStatus.Tokens))
	for _, token := range instanceStatus.Tokens {
		existingSecrets[token.SecretName] = struct{}{}
	}
	for _, token := range cr.Spec.Tokens {
		secretName := generateTokenSecretName(cr.Name, instanceStatus.GrafanaName, token.Name)
		if _, exists := existingSecrets[secretName]; exists {
			continue
		}

		cmd := models.AddServiceAccountTokenCommand{Name: token.Name}
		if token.Expires != nil {
			cmd.SecondsToLive = int64(time.Until(token.Expires.Time).Seconds())
		}
		resp, err := grafanaClient.ServiceAccounts.CreateToken(
			service_accounts.NewCreateTokenParamsWithContext(ctx).
				WithServiceAccountID(instanceStatus.ServiceAccountID).
				WithBody(&cmd),
		)
		if err != nil {
			return fmt.Errorf("creating token %q for service account %d: %w", token.Name, instanceStatus.ServiceAccountID, err)
		}
		keyResult := resp.Payload

		// Create a Kubernetes Secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Namespace,
				Labels:    map[string]string{"app": "grafana-serviceaccount-token"},
			},
			Data: map[string][]byte{
				"token": []byte(keyResult.Key),
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

		instanceStatus.Tokens = append(instanceStatus.Tokens, v1beta1.GrafanaServiceAccountTokenStatus{
			Name:       keyResult.Name,
			TokenID:    keyResult.ID,
			SecretName: secretName,
		})
		existingSecrets[secretName] = struct{}{}
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) reconcilePermissionsForInstance(
	ctx context.Context,
	gafanaClient *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
	instance *v1beta1.GrafanaServiceAccountInstanceStatus,
) error {
	if instance.ServiceAccountID == 0 {
		// No service account ID yet, so skip
		return nil
	}

	const resource = "serviceaccounts"
	resourceID := strconv.FormatInt(instance.ServiceAccountID, 10)

	// Check if the resource exists
	resp, err := gafanaClient.AccessControl.GetResourcePermissionsWithParams(
		access_control.NewGetResourcePermissionsParamsWithContext(ctx).
			WithResource(resource).
			WithResourceID(resourceID),
	)
	if err != nil {
		return fmt.Errorf("listing current resource permissions for service account %d: %w", instance.ServiceAccountID, err)
	}

	type subject struct {
		teamID int64
		userID int64
	}
	desired := map[subject]string{}

	// Collect desired
	for _, p := range cr.Spec.Permissions {
		switch {
		case p.Team != "" && p.User == "":
			resp, err := gafanaClient.Teams.SearchTeams(
				teams.NewSearchTeamsParamsWithContext(ctx).WithQuery(&p.Team),
			)
			if err != nil {
				return fmt.Errorf("searching Grafana team %q: %w", p.Team, err)
			}
			if resp.Payload.TotalCount == 0 {
				return fmt.Errorf("team %q not found in Grafana", p.Team)
			}
			if resp.Payload.TotalCount > 1 {
				return fmt.Errorf("multiple teams found with name %q", p.Team)
			}
			desired[subject{teamID: resp.Payload.Teams[0].ID}] = p.Permission

		case p.Team == "" && p.User != "":
			resp, err := gafanaClient.Users.GetUserByLoginOrEmailWithParams(
				users.NewGetUserByLoginOrEmailParamsWithContext(ctx).
					WithLoginOrEmail(p.User),
			)
			if err != nil {
				return fmt.Errorf("searching Grafana user %q: %w", p.User, err)
			}
			desired[subject{userID: resp.Payload.ID}] = p.Permission

		default:
			return fmt.Errorf("malformed permission entry: team=%q user=%q", p.Team, p.User)
		}
	}

	var cmds []*models.SetResourcePermissionCommand

	// Compare with existing permissions
	existing := map[subject]struct{}{}
	for _, curr := range resp.Payload {
		s := subject{teamID: curr.TeamID, userID: curr.UserID}
		desiredPerm, inDesired := desired[s]
		if inDesired {
			existing[s] = struct{}{}
		}
		if curr.Permission != desiredPerm {
			cmds = append(cmds, &models.SetResourcePermissionCommand{
				TeamID:     s.teamID,
				UserID:     s.userID,
				Permission: desiredPerm,
			})
		}
	}

	// Add new
	for s, desiredPerm := range desired {
		if _, exists := existing[s]; !exists {
			cmds = append(cmds, &models.SetResourcePermissionCommand{
				TeamID:     s.teamID,
				UserID:     s.userID,
				Permission: desiredPerm,
			})
		}
	}

	if len(cmds) == 0 {
		return nil
	}
	_, err = gafanaClient.AccessControl.SetResourcePermissions( // nolint:errcheck
		access_control.NewSetResourcePermissionsParamsWithContext(ctx).
			WithResource(resource).
			WithResourceID(resourceID).
			WithBody(&models.SetPermissionsCommand{Permissions: cmds}),
	)
	if err != nil {
		return fmt.Errorf("setting permissions for service account %d: %w", instance.ServiceAccountID, err)
	}
	return nil
}

func (r *GrafanaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager, _ context.Context) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&v1beta1.GrafanaServiceAccount{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Watches(
			&v1beta1.Grafana{},
			handler.EnqueueRequestsFromMapFunc(r.grafanaToServiceAccounts),
			builder.WithPredicates(
				predicate.Funcs{
					CreateFunc: func(e event.CreateEvent) bool { return true },
					DeleteFunc: func(e event.DeleteEvent) bool { return true },
					UpdateFunc: func(e event.UpdateEvent) bool {
						oldG, okOld := e.ObjectOld.(*v1beta1.Grafana)
						newG, okNew := e.ObjectNew.(*v1beta1.Grafana)
						if !okOld || !okNew {
							return false
						}
						return !reflect.DeepEqual(oldG.Labels, newG.Labels)
					},
					GenericFunc: func(e event.GenericEvent) bool { return false },
				},
			),
		).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

func (r *GrafanaServiceAccountReconciler) updateStatusAndFinalizer(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) {
	cr.Status.LastResync = metav1.Time{Time: time.Now()}

	slices.SortFunc(cr.Status.Instances, func(a, b v1beta1.GrafanaServiceAccountInstanceStatus) int {
		if a.GrafanaNamespace == b.GrafanaNamespace {
			return strings.Compare(a.GrafanaName, b.GrafanaName)
		}
		return strings.Compare(a.GrafanaNamespace, b.GrafanaNamespace)
	})

	err := r.Status().Update(ctx, cr)
	if err != nil {
		logf.FromContext(ctx).Error(err, "failed to update GrafanaServiceAccount status")
	}

	if meta.IsStatusConditionTrue(cr.Status.Conditions, conditionNoMatchingInstance) {
		err := removeFinalizer(ctx, r.Client, cr)
		if err != nil {
			logf.FromContext(ctx).Error(err, "failed to remove finalizer")
		}
	} else {
		err := addFinalizer(ctx, r.Client, cr)
		if err != nil {
			logf.FromContext(ctx).Error(err, "failed to set finalizer")
		}
	}
}

// grafanaToServiceAccounts is used to enqueue GSA reconcile requests when a Grafana changes.
func (r *GrafanaServiceAccountReconciler) grafanaToServiceAccounts(ctx context.Context, obj client.Object) []reconcile.Request {
	ctx = logf.IntoContext(ctx, logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler"))

	var list v1beta1.GrafanaServiceAccountList
	err := r.List(ctx, &list)
	if err != nil {
		return nil
	}

	grafana, ok := obj.(*v1beta1.Grafana)
	if !ok {
		return nil
	}

	var requests []reconcile.Request
	for _, serviceAccount := range list.Items {
		if serviceAccount.Spec.InstanceSelector == nil {
			continue
		}
		selector, err := metav1.LabelSelectorAsSelector(serviceAccount.Spec.InstanceSelector)
		if err != nil {
			logf.FromContext(ctx).Error(err, "invalid instanceSelector", "selector", serviceAccount.Spec.InstanceSelector)
			continue
		}
		if selector.Matches(labels.Set(grafana.Labels)) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: serviceAccount.Namespace,
					Name:      serviceAccount.Name,
				},
			})
		}
	}
	return requests
}

func (r *GrafanaServiceAccountReconciler) revokeToken(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
	tokenID int64,
) error {
	_, err := client.ServiceAccounts.DeleteTokenWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithTokenID(tokenID),
	)
	if err == nil {
		return nil
	}

	// TODO: check if this is the correct error type
	if !errors.Is(err, &service_accounts.DeleteTokenInternalServerError{}) {
		return fmt.Errorf("deleting token %d: %w", tokenID, err)
	}

	logf.FromContext(ctx).Info("token not found in Grafana, skip", "serviceAccountID", serviceAccountID, "tokenID", tokenID)
	return nil
}

func (r *GrafanaServiceAccountReconciler) revokeSecret(ctx context.Context, namespace string, secretName string) error {
	var secret corev1.Secret
	err := r.Client.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, &secret)
	if err != nil {
		if kuberr.IsNotFound(err) {
			logf.FromContext(ctx).Info("token secret not found, skip", "secretName", secretName)
			return nil
		}
		return fmt.Errorf("getting token secret %s: %w", secretName, err)
	}

	err = r.Client.Delete(ctx, &secret)
	if err != nil {
		return fmt.Errorf("deleting token secret %s: %w", secretName, err)
	}

	return nil
}
