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
	"reflect"
	"strconv"
	"time"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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

		var toRemove v1beta1.NamespacedResourceList
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
			cr := crs.Find(namespace, name)
			if cr != nil && cr.Status.ID == serviceAccountID {
				remained++
				continue
			}

			r.removeServiceAccount(ctx, grafanaClient, serviceAccountID)

			toRemove = append(toRemove, serviceAccountResource)
			removed++
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latest := &v1beta1.Grafana{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: grafana.Namespace, Name: grafana.Name}, latest)
			if err != nil {
				return err
			}

			for _, resource := range toRemove {
				latest.Status.ServiceAccounts = latest.Status.ServiceAccounts.Remove(resource.Namespace(), resource.Name())
			}

			return r.Client.Status().Update(ctx, latest)
		})
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
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	for _, tokenStatus := range cr.Status.Tokens {
		err := r.revokeSecret(ctx, cr.Namespace, tokenStatus.SecretName)
		if err != nil {
			return fmt.Errorf("removing token resources %s: %w", tokenStatus.Name, err)
		}
	}
	cr.Status.Tokens = nil

	r.removeServiceAccount(ctx, client, cr.Status.ID)

	return nil
}

func (r *GrafanaServiceAccountReconciler) removeServiceAccount(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
) {
	log := logf.FromContext(ctx)

	if _, err := client.ServiceAccounts.DeleteServiceAccountWithParams( // nolint:errcheck
		service_accounts.
			NewDeleteServiceAccountParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID),
	); err != nil {
		log.Error(err, "failed to delete service account from Grafana", "serviceAccountID", serviceAccountID)
	}
}

// Reconcile contains the main reconciliation logic for GrafanaServiceAccount.
func (r *GrafanaServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaServiceAccount{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, cr)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil // CR no longer exists
		}
		return ctrl.Result{}, fmt.Errorf("getting GrafanaServiceAccount %q: %w", req, err)
	}

	// If CR is marked for deletion, run finalization logic.
	if cr.GetDeletionTimestamp() != nil {
		err := r.finalize(ctx, cr)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("executing finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	defer r.syncStatusAndFinalizer(ctx, cr)

	// Get Grafana grafanas matching the CR.
	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, cr)
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
	log.V(1).Info("found matching Grafana instances", "count", len(grafanas))

	// Apply changes to each matching Grafana instance.
	applyErrors := map[string]string{}
	for _, grafana := range grafanas {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating Grafana client: %w", err)
		}

		err = r.setupServiceAccount(ctx, grafanaClient, cr)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			continue
		}

		saUID := strconv.FormatInt(cr.Status.ID, 10)
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latest := &v1beta1.Grafana{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: grafana.Namespace, Name: grafana.Name}, latest)
			if err != nil {
				return err
			}

			latest.Status.ServiceAccounts = latest.Status.ServiceAccounts.Add(cr.Namespace, cr.Name, saUID)

			return r.Client.Status().Update(ctx, latest)
		})
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

func (r *GrafanaServiceAccountReconciler) syncStatusAndFinalizer(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) {
	cr.Status.LastResync = metav1.Time{Time: time.Now()}
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

func (r *GrafanaServiceAccountReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaServiceAccount) error {
	if !controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
		return nil
	}

	grafanas, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances for finalization: %w", err)
	}

	for _, grafana := range grafanas {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("creating Grafana client: %w", err)
		}

		err = r.cleanupServiceAccount(ctx, grafanaClient, cr)
		if err != nil {
			return fmt.Errorf("cleaning up service account: %w", err)
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latest := &v1beta1.Grafana{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: grafana.Namespace, Name: grafana.Name}, latest)
			if err != nil {
				return err
			}

			latest.Status.ServiceAccounts = latest.Status.ServiceAccounts.Remove(cr.Namespace, cr.Name)

			return r.Client.Status().Update(ctx, latest)
		})
		if err != nil {
			return fmt.Errorf("updating Grafana status after service account removal: %w", err)
		}
	}

	err = removeFinalizer(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("removing finalizer %s/%s: %w", cr.Namespace, cr.Name, err)
	}

	return nil
}

// setupServiceAccount creates or updates the service account in Grafana.
func (r *GrafanaServiceAccountReconciler) setupServiceAccount(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	err := r.lookupServiceAccount(ctx, client, cr)
	if err != nil {
		return fmt.Errorf("looking up service account: %w", err)
	}

	if cr.Status.ID == 0 {
		err := r.createServiceAccount(ctx, client, cr)
		if err != nil {
			return fmt.Errorf("creating service account: %w", err)
		}
	} else {
		err := r.updateServiceAccount(ctx, client, cr)
		if err != nil {
			return fmt.Errorf("updating service account: %w", err)
		}
	}

	err = r.reconcileTokens(ctx, client, cr)
	if err != nil {
		return fmt.Errorf("reconciling tokens: %w", err)
	}
	err = r.reconcilePermissions(ctx, client, cr)
	if err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) lookupServiceAccount(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	search, err := client.ServiceAccounts.SearchOrgServiceAccountsWithPaging(
		service_accounts.
			NewSearchOrgServiceAccountsWithPagingParamsWithContext(ctx).
			WithQuery(&cr.Spec.Name),
	)
	if err != nil {
		return fmt.Errorf("searching service accounts: %w", err)
	}
	for _, serviceAccount := range search.Payload.ServiceAccounts {
		if serviceAccount.Name == cr.Spec.Name {
			if cr.Status.ID == 0 || serviceAccount.ID == cr.Status.ID {
				cr.Status.ID = serviceAccount.ID
				return nil
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

	return nil
}

func (r *GrafanaServiceAccountReconciler) createServiceAccount(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	resp, err := client.ServiceAccounts.CreateServiceAccount(
		service_accounts.
			NewCreateServiceAccountParamsWithContext(ctx).
			WithBody(&models.CreateServiceAccountForm{
				IsDisabled: cr.Spec.IsDisabled,
				Name:       cr.Spec.Name,
				Role:       cr.Spec.Role,
			}),
	)
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}
	cr.Status.ID = resp.Payload.ID

	return nil
}

func (r *GrafanaServiceAccountReconciler) updateServiceAccount(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	if _, err := client.ServiceAccounts.UpdateServiceAccount( // nolint:errcheck
		service_accounts.
			NewUpdateServiceAccountParamsWithContext(ctx).
			WithBody(&models.UpdateServiceAccountForm{
				IsDisabled:       cr.Spec.IsDisabled,
				Name:             cr.Spec.Name,
				Role:             cr.Spec.Role,
				ServiceAccountID: cr.Status.ID,
			}).
			WithServiceAccountID(cr.Status.ID),
	); err != nil {
		return err
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) grafanaToServiceAccounts(ctx context.Context, obj client.Object) []reconcile.Request {
	ctx = logf.IntoContext(ctx, logf.FromContext(ctx).WithName("GrafanaServiceAccountReconciler"))

	var gsaList v1beta1.GrafanaServiceAccountList
	if err := r.List(ctx, &gsaList); err != nil {
		return nil
	}

	grafana := obj.(*v1beta1.Grafana) //nolint:errcheck
	var requests []reconcile.Request
	for _, gsa := range gsaList.Items {
		if gsa.Spec.InstanceSelector == nil {
			continue
		}
		sel, err := metav1.LabelSelectorAsSelector(gsa.Spec.InstanceSelector)
		if err != nil {
			logf.FromContext(ctx).Error(err, "failed to convert label selector", "selector", gsa.Spec.InstanceSelector)
			continue
		}
		if sel.Matches(labels.Set(grafana.Labels)) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: gsa.Namespace,
					Name:      gsa.Name,
				},
			})
		}
	}
	return requests
}

// SetupWithManager registers the reconciler with the manager.
func (r *GrafanaServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
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
				// res, err := r.syncServiceAccounts(ctx)
				elapsed := time.Since(start).Milliseconds()
				metrics.InitialServiceAccountSyncDuration.Set(float64(elapsed))

				// if err != nil {
				// 	log.Error(err, "synchronizing service accounts")
				// 	continue
				// }
				// if res.Requeue {
				// 	log.Info("more service accounts to synchronize")
				// 	continue
				// }
				log.V(1).Info("service account sync complete")
				return
			}
		}
	}()

	return nil
}

// reconcileTokens creates or updates tokens in Grafana based on the CR spec and removes expired or stale tokens.
func (r *GrafanaServiceAccountReconciler) reconcileTokens(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	if cr.Spec.GenerateTokenSecret && len(cr.Spec.Tokens) == 0 {
		cr.Spec.Tokens = []v1beta1.GrafanaServiceAccountToken{{Name: fmt.Sprintf("%s-default-token", cr.Name)}}
	}

	err := r.cleanupTokens(ctx, client, cr)
	if err != nil {
		return fmt.Errorf("cleaning up tokens: %w", err)
	}
	err = r.createMissingTokens(ctx, client, cr)
	if err != nil {
		return fmt.Errorf("creating missing tokens: %w", err)
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) createMissingTokens(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	existingSecrets := map[string]struct{}{}
	for _, tokenStatus := range cr.Status.Tokens {
		existingSecrets[tokenStatus.SecretName] = struct{}{}
	}
	for _, token := range cr.Spec.Tokens {
		secretName := token.Name
		if _, exists := existingSecrets[secretName]; exists {
			continue
		}

		apiKey, err := r.createToken(ctx, client, cr.Status.ID, token.Name, token.Expires)
		if err != nil {
			return fmt.Errorf("creating token for service account: %w", err)
		}
		err = r.storeTokenSecret(ctx, cr, secretName, []byte(apiKey.Key), token.Expires)
		if err != nil {
			return fmt.Errorf("storing token secret %s/%s: %w", cr.Namespace, secretName, err)
		}

		cr.Status.Tokens = append(cr.Status.Tokens, v1beta1.GrafanaServiceAccountTokenStatus{
			Name:       apiKey.Name,
			TokenID:    apiKey.ID,
			SecretName: secretName,
		})
		existingSecrets[secretName] = struct{}{}
	}

	return nil
}

func (r *GrafanaServiceAccountReconciler) createToken(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	serviceAccountID int64,
	tokenName string,
	expires *metav1.Time,
) (*models.NewAPIKeyResult, error) {
	cmd := models.AddServiceAccountTokenCommand{
		Name: tokenName,
	}
	if expires != nil {
		cmd.SecondsToLive = int64(time.Until(expires.Time).Seconds())
	}
	resp, err := client.ServiceAccounts.CreateToken(
		service_accounts.
			NewCreateTokenParamsWithContext(ctx).
			WithServiceAccountID(serviceAccountID).
			WithBody(&cmd),
	)
	if err != nil {
		//TODO: handle already exists error
		return nil, fmt.Errorf("creating token for service account: %w", err)
	}

	return resp.Payload, nil
}

func (r *GrafanaServiceAccountReconciler) cleanupTokens(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	newStatusTokens := make([]v1beta1.GrafanaServiceAccountTokenStatus, 0, len(cr.Spec.Tokens))

	now := time.Now()

	tokensMap := make(map[string]*metav1.Time, len(cr.Spec.Tokens))
	for _, token := range cr.Spec.Tokens {
		tokensMap[token.Name] = token.Expires
	}

	for _, tokenStatus := range cr.Status.Tokens {
		if expires, exists := tokensMap[tokenStatus.Name]; exists && (expires == nil || now.Before(expires.Time)) {
			newStatusTokens = append(newStatusTokens, tokenStatus)
			continue
		}

		err := r.revokeToken(ctx, client, cr.Status.ID, tokenStatus.TokenID)
		if err != nil {
			return fmt.Errorf("removing token %s: %w", tokenStatus.Name, err)
		}
		err = r.revokeSecret(ctx, cr.Namespace, tokenStatus.SecretName)
		if err != nil {
			return fmt.Errorf("removing token secret %s/%s: %w", cr.Namespace, tokenStatus.SecretName, err)
		}
	}
	cr.Status.Tokens = newStatusTokens

	return nil
}

func (r *GrafanaServiceAccountReconciler) storeTokenSecret(
	ctx context.Context,
	cr *v1beta1.GrafanaServiceAccount,
	secretName string,
	token []byte,
	expires *metav1.Time,
) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: cr.Namespace,
			Labels:    map[string]string{"app": "grafana-serviceaccount-token"},
		},
		Data: map[string][]byte{
			"token": token,
		},
	}
	if expires != nil {
		secret.Annotations = map[string]string{
			"grafana.integreatly.org/token-expiry": expires.Format(time.RFC3339),
		}
	}
	err := controllerutil.SetControllerReference(cr, secret, r.Scheme)
	if err != nil {
		logf.FromContext(ctx).Error(err, "failed to set owner reference on token secret")
	}
	err = r.Client.Create(ctx, secret)
	if err != nil {
		//TODO: handle already exists error
		return fmt.Errorf("creating token secret %s: %w", secretName, err)
	}

	return nil
}

// reconcilePermissions assigns or removes RBAC roles based on the CR spec.
func (r *GrafanaServiceAccountReconciler) reconcilePermissions(
	ctx context.Context,
	client *genapi.GrafanaHTTPAPI,
	cr *v1beta1.GrafanaServiceAccount,
) error {
	const resource = "serviceaccounts"
	resourceID := strconv.FormatInt(cr.Status.ID, 10)

	resp, err := client.AccessControl.GetResourcePermissionsWithParams(
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
			resp, err := client.Teams.SearchTeams(
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
			resp, err := client.Users.GetUserByLoginOrEmailWithParams(
				users.NewGetUserByLoginOrEmailParamsWithContext(ctx).
					WithLoginOrEmail(p.User),
			)
			if err != nil {
				return fmt.Errorf("searching user %q: %w", p.User, err)
			}
			desired[subject{userID: resp.Payload.ID}] = p.Permission
		default:
			err := fmt.Errorf("malformed permission entry: team=%q user=%q", p.Team, p.User)
			setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidSpec", err.Error())
			meta.RemoveStatusCondition(&cr.Status.Conditions, conditionServiceAccountSynchronized)
			return err
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
		return nil
	}

	_, err = client.AccessControl.SetResourcePermissions( // nolint:errcheck
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

	var notFound *service_accounts.DeleteTokenInternalServerError // TODO: check if this is the correct error type
	if !errors.As(err, &notFound) {
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
