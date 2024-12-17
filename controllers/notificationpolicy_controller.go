/*
Copyright 2022.

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
	"fmt"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionNotificationPolicySynchronized = "NotificationPolicySynchronized"
)

// GrafanaNotificationPolicyReconciler reconciles a GrafanaNotificationPolicy object
type GrafanaNotificationPolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies/finalizers,verbs=update

// TODO listen for updates on GrafanaNotificationPolicyRoutes
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GrafanaNotifictionPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *GrafanaNotificationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
	r.Log = log.FromContext(ctx)

	notificationPolicy := &grafanav1beta1.GrafanaNotificationPolicy{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, notificationPolicy)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "Failed to get GrafanaNotificationPolicy")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if notificationPolicy.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(notificationPolicy, grafanaFinalizer) {
			err := r.finalize(ctx, notificationPolicy)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to finalize GrafanaNotificationPolicy: %w", err)
			}
			controllerutil.RemoveFinalizer(notificationPolicy, grafanaFinalizer)
			if err := r.Update(ctx, notificationPolicy); err != nil {
				r.Log.Error(err, "failed to remove finalizer")
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to update GrafanaNotificationPolicy: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	defer func() {
		if err := r.Client.Status().Update(ctx, notificationPolicy); err != nil {
			r.Log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(notificationPolicy.Status.Conditions, conditionNoMatchingInstance) {
			controllerutil.RemoveFinalizer(notificationPolicy, grafanaFinalizer)
		} else {
			controllerutil.AddFinalizer(notificationPolicy, grafanaFinalizer)
		}
		if err := r.Update(ctx, notificationPolicy); err != nil {
			r.Log.Error(err, "failed to set finalizer")
		}
	}()

	instances, err := GetMatchingInstances(ctx, r.Client, notificationPolicy.Spec.InstanceSelector)
	if err != nil {
		setNoMatchingInstance(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, "ErrFetchingInstances", fmt.Sprintf("error occurred during fetching of instances: %s", err.Error()))
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)
		r.Log.Error(err, "could not find matching instances")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if len(instances.Items) == 0 {
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)
		setNoMatchingInstance(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, "EmptyAPIReply", "Instances could not be fetched, reconciliation will be retried")
		return ctrl.Result{}, nil
	}

	removeNoMatchingInstance(&notificationPolicy.Status.Conditions)

	var matchingNotificationPolicyRoutes *v1beta1.GrafanaNotificationPolicyRouteList
	if notificationPolicy.Spec.RouteSelector != nil {
		var namespace *string
		if notificationPolicy.Spec.AllowCrossNamespaceImport != nil && !*notificationPolicy.Spec.AllowCrossNamespaceImport {
			ns := notificationPolicy.GetObjectMeta().GetNamespace()
			namespace = &ns
		}
		matchingNotificationPolicyRoutes, err = getMatchingNotificationPolicyRoutes(ctx, r.Client, notificationPolicy.Spec.RouteSelector, namespace)
		if err != nil {
			r.Log.Error(err, "failed to get matching GrafanaNotificationPolicyRoutes")
			return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to get matching GrafanaNotificationPolicyRoutes: %w", err)
		}
	}

	if matchingNotificationPolicyRoutes != nil {
		notificationPolicy = mergeNotificationPolicyRoutesWithRouteList(notificationPolicy, matchingNotificationPolicyRoutes)
	}

	applyErrors := make(map[string]string)
	appliedCount := 0
	for _, grafana := range instances.Items {
		// can be removed in go 1.22+
		grafana := grafana
		appliedPolicy := grafana.Annotations[annotationAppliedNotificationPolicy]
		if appliedPolicy != "" && appliedPolicy != notificationPolicy.NamespacedResource() {
			controllerLog.Info("instance already has a different notification policy applied - skipping", "grafana", grafana.Name)
			continue
		}
		appliedCount++

		if grafana.Status.Stage != grafanav1beta1.OperatorStageComplete || grafana.Status.StageStatus != grafanav1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		err := r.reconcileWithInstance(ctx, &grafana, notificationPolicy)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}
	condition := buildSynchronizedCondition("Notification Policy", conditionNotificationPolicySynchronized, notificationPolicy.Generation, applyErrors, appliedCount)
	meta.SetStatusCondition(&notificationPolicy.Status.Conditions, condition)
	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	discoveredRoutes := matchingNotificationPolicyRoutes.StatusDiscoveredRoutes()
	if len(discoveredRoutes) > 0 {
		notificationPolicy.Status.DiscoveredRoutes = &discoveredRoutes
	}

	return ctrl.Result{RequeueAfter: notificationPolicy.Spec.ResyncPeriod.Duration}, nil
}

// mergeNotificationPolicyRoutesWithRouteList merges a list of GrafanaNotificationPolicyRoutes into the
// spec.Route.Routes of a GrafanaNotificationPolicy following the specified priorities on the Routes
func mergeNotificationPolicyRoutesWithRouteList(notificationPolicy *grafanav1beta1.GrafanaNotificationPolicy, notificationPolicyRouteList *grafanav1beta1.GrafanaNotificationPolicyRouteList) *grafanav1beta1.GrafanaNotificationPolicy {
	if notificationPolicyRouteList == nil {
		return notificationPolicy
	}

	notificationPolicyRouteList.SortByPriority()

	for _, route := range notificationPolicyRouteList.Items {
		notificationPolicy.Spec.Route.Routes = append(notificationPolicy.Spec.Route.Routes, route.Spec.Route)
	}

	return notificationPolicy
}

func (r *GrafanaNotificationPolicyReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationPolicy *grafanav1beta1.GrafanaNotificationPolicy) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	trueRef := "true"
	editable := true
	if notificationPolicy.Spec.Editable != nil && !*notificationPolicy.Spec.Editable {
		editable = false
	}
	params := provisioning.NewPutPolicyTreeParams().WithBody(notificationPolicy.Spec.Route.ToModelRoute())
	if editable {
		params.SetXDisableProvenance(&trueRef)
	}
	if _, err := cl.Provisioning.PutPolicyTree(params); err != nil { //nolint:errcheck
		return fmt.Errorf("applying notification policy: %w", err)
	}
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[annotationAppliedNotificationPolicy] = notificationPolicy.NamespacedResource()
	if err := r.Client.Update(ctx, instance); err != nil {
		return fmt.Errorf("saving applied policy to instance CR: %w", err)
	}
	return nil
}

func (r *GrafanaNotificationPolicyReconciler) resetInstance(ctx context.Context, instance *grafanav1beta1.Grafana) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}
	if _, err := cl.Provisioning.ResetPolicyTree(); err != nil { //nolint:errcheck
		return fmt.Errorf("resetting policy tree")
	}
	delete(instance.Annotations, annotationAppliedNotificationPolicy)
	if err := r.Client.Update(ctx, instance); err != nil {
		return fmt.Errorf("removing applied policy from instance CR: %w", err)
	}

	return nil
}

func (r *GrafanaNotificationPolicyReconciler) finalize(ctx context.Context, notificationPolicy *grafanav1beta1.GrafanaNotificationPolicy) error {
	r.Log.Info("Finalizing GrafanaNotificationPolicy")

	instances, err := GetMatchingInstances(ctx, r.Client, notificationPolicy.Spec.InstanceSelector)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances.Items {
		instance := i
		appliedPolicy := i.Annotations[annotationAppliedNotificationPolicy]
		if appliedPolicy != "" && appliedPolicy != notificationPolicy.NamespacedResource() {
			r.Log.Info("instance already has a different notification policy applied - skipping", "grafana", instance.Name)
			continue
		}

		if err := r.resetInstance(ctx, &instance); err != nil {
			return fmt.Errorf("resetting instance notification policy: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaNotificationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaNotificationPolicy{}).
		Watches(&grafanav1beta1.GrafanaContactPoint{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			// resync all notification policies for now. Can be optimized by comparing instance selectors
			nps := &grafanav1beta1.GrafanaNotificationPolicyList{}
			if err := r.List(ctx, nps); err != nil {
				r.Log.Error(err, "failed to fetch notification policies for watch mapping")
				return nil
			}
			requests := make([]reconcile.Request, len(nps.Items))
			for i, np := range nps.Items {
				requests[i] = reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      np.Name,
						Namespace: np.Namespace,
					},
				}
			}
			return requests
		})).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

// getMatchingNotificationPolicyRoutes retrieves all GrafanaNotificationPolicyRoutes for the given labelSelector
// results will be limited to namespace when specified
func getMatchingNotificationPolicyRoutes(ctx context.Context, k8sClient client.Client, labelSelector *metav1.LabelSelector, namespace *string) (*v1beta1.GrafanaNotificationPolicyRouteList, error) {
	if labelSelector == nil {
		return nil, nil
	}

	var list v1beta1.GrafanaNotificationPolicyRouteList
	opts := []client.ListOption{
		client.MatchingLabels(labelSelector.MatchLabels),
	}

	if namespace != nil {
		opts = append(opts, client.InNamespace(*namespace))
	}

	err := k8sClient.List(ctx, &list, opts...)
	return &list, err
}
