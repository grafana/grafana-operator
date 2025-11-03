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
	"errors"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

var ErrLoopDetected = errors.New("loop detected")

const (
	conditionNotificationPolicySynchronized  = "NotificationPolicySynchronized"
	conditionRoutesIgnoredDueToRouteSelector = "RoutesIgnoredDueToRouteSelector"
	annotationAppliedNotificationPolicy      = "operator.grafana.com/applied-notificationpolicy"

	conditionReasonFieldsMutuallyExclusive = "FieldsMutuallyExclusive"
	conditionReasonLoopDetected            = "LoopDetected"
)

// GrafanaNotificationPolicyReconciler reconciles a GrafanaNotificationPolicy object
type GrafanaNotificationPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Cfg      *Config
}

func (r *GrafanaNotificationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
	ctx = logf.IntoContext(ctx, log)

	notificationPolicy := &v1beta1.GrafanaNotificationPolicy{}

	err := r.Get(ctx, req.NamespacedName, notificationPolicy)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting GrafanaNotificationPolicy cr: %w", err)
	}

	if notificationPolicy.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(notificationPolicy, grafanaFinalizer) {
			if err := r.finalize(ctx, notificationPolicy); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaNotificationPolicy: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, notificationPolicy); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, notificationPolicy)

	if notificationPolicy.Spec.Suspend {
		setSuspended(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&notificationPolicy.Status.Conditions)

	// check if spec is valid
	if !notificationPolicy.Spec.Route.IsRouteSelectorMutuallyExclusive() {
		setInvalidSpecMutuallyExclusive(&notificationPolicy.Status.Conditions, notificationPolicy.Generation)
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)

		return ctrl.Result{}, fmt.Errorf("invalid route spec discovered: routeSelector is mutually exclusive with routes")
	}

	removeInvalidSpec(&notificationPolicy.Status.Conditions)

	// Assemble routes and check for loops
	var mergedRoutes []*v1beta1.GrafanaNotificationPolicyRoute
	if notificationPolicy.Spec.Route.HasRouteSelector() {
		mergedRoutes, err = assembleNotificationPolicyRoutes(ctx, r.Client, notificationPolicy)
		if errors.Is(err, ErrLoopDetected) {
			meta.SetStatusCondition(&notificationPolicy.Status.Conditions, metav1.Condition{
				Type:               conditionNotificationPolicyLoopDetected,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: notificationPolicy.Generation,
				Reason:             conditionReasonLoopDetected,
				Message:            fmt.Sprintf("Loop detected in notification policy routes: %s", err.Error()),
			})
			meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)

			return ctrl.Result{}, fmt.Errorf("failed to assemble notification policy routes: %w", err)
		}

		if err != nil {
			r.Recorder.Event(notificationPolicy, corev1.EventTypeWarning, "AssemblyFailed", fmt.Sprintf("Failed to assemble GrafanaNotificationPolicy using routeSelectors: %v", err))
			return ctrl.Result{}, fmt.Errorf("failed to assemble GrafanaNotificationPolicy using routeSelectors: %w", err)
		}
	}

	meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicyLoopDetected)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationPolicy)
	if err != nil {
		setNoMatchingInstancesCondition(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, err)
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, err)
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&notificationPolicy.Status.Conditions)
	log.Info("found matching Grafana instances for notificationPolicy", "count", len(instances))

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		appliedPolicy := grafana.Annotations[annotationAppliedNotificationPolicy]
		if appliedPolicy != "" && appliedPolicy != notificationPolicy.NamespacedResource() {
			log.Info("instance already has a different notification policy applied - skipping", "grafana", grafana.Name)
			continue
		}

		err := r.reconcileWithInstance(ctx, &grafana, notificationPolicy)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Notification Policy", conditionNotificationPolicySynchronized, notificationPolicy.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&notificationPolicy.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	if len(mergedRoutes) > 0 {
		status := statusDiscoveredRoutes(mergedRoutes)
		notificationPolicy.Status.DiscoveredRoutes = &status
	}

	if err := r.updateNotificationPolicyRoutesStatus(ctx, notificationPolicy, mergedRoutes); err != nil {
		log.Error(err, "failed to add merged events to routes")
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(notificationPolicy.Spec.ResyncPeriod)}, nil
}

// assembleNotificationPolicyRoutes iterates over all routeSelectors transitively.
// returns an assembled GrafanaNotificationPolicy as well as a list of all merged routes.
// it ensures that there are no reference loops when discovering routes via labelSelectors

func assembleNotificationPolicyRoutes(ctx context.Context, k8sClient client.Client, notificationPolicy *v1beta1.GrafanaNotificationPolicy) ([]*v1beta1.GrafanaNotificationPolicyRoute, error) {
	var namespace *string

	if !notificationPolicy.AllowCrossNamespace() {
		ns := notificationPolicy.GetObjectMeta().GetNamespace()
		namespace = &ns
	}

	mergedRoutes := []*v1beta1.GrafanaNotificationPolicyRoute{}

	// visitedGlobal keeps track of all routes that have been appended to mergedRoutes
	// so we can record a status update for them later
	visitedGlobal := make(map[string]bool)

	// visitedChilds keeps track of all routes that have been visited on the current path
	// so we can detect loops
	visitedChilds := make(map[string]bool)

	var assembleRoute func(*v1beta1.Route) error

	assembleRoute = func(route *v1beta1.Route) error {
		if route.RouteSelector != nil {
			routes, err := getMatchingNotificationPolicyRoutes(ctx, k8sClient, route.RouteSelector, namespace)
			if err != nil {
				return fmt.Errorf("failed to get matching routes: %w", err)
			}

			// Replace the RouteSelector with matched routes
			route.RouteSelector = nil

			for i := range routes {
				matchedRoute := &routes[i]
				key := matchedRoute.NamespacedResource()

				if _, exists := visitedGlobal[key]; !exists {
					mergedRoutes = append(mergedRoutes, matchedRoute)
					visitedGlobal[key] = true
				}

				if _, exists := visitedChilds[key]; exists {
					return fmt.Errorf("%w: %s exists", ErrLoopDetected, key)
				}

				visitedChilds[key] = true

				// Recursively assemble the matched route
				if err := assembleRoute(&matchedRoute.Spec.Route); err != nil {
					return err
				}

				delete(visitedChilds, key)

				route.Routes = append(route.Routes, &matchedRoute.Spec.Route)
			}
		} else {
			// if no RouteSelector is specified, process inline routes, as they are mutually exclusive
			for i, inlineRoute := range route.Routes {
				if err := assembleRoute(inlineRoute); err != nil {
					return err
				}

				route.Routes[i] = inlineRoute
			}
		}

		return nil
	}

	// Start with Spec.Route
	if err := assembleRoute(notificationPolicy.Spec.Route); err != nil {
		return nil, err
	}

	return mergedRoutes, nil
}

func (r *GrafanaNotificationPolicyReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, notificationPolicy *v1beta1.GrafanaNotificationPolicy) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	trueRef := "true" //nolint:goconst

	editable := true //nolint:staticcheck
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

	err = addAnnotation(ctx, r.Client, instance, annotationAppliedNotificationPolicy, notificationPolicy.NamespacedResource())
	if err != nil {
		return fmt.Errorf("saving applied notification policy to Grafana CR: %w", err)
	}

	return nil
}

func (r *GrafanaNotificationPolicyReconciler) finalize(ctx context.Context, notificationPolicy *v1beta1.GrafanaNotificationPolicy) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaNotificationPolicy")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationPolicy)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		appliedPolicy := grafana.Annotations[annotationAppliedNotificationPolicy]
		if appliedPolicy != "" && appliedPolicy != notificationPolicy.NamespacedResource() {
			log.Info("instance already has a different notification policy applied - skipping", "grafana", grafana.Name)
			continue
		}

		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return fmt.Errorf("building grafana client: %w", err)
		}

		if _, err := grafanaClient.Provisioning.ResetPolicyTree(); err != nil { //nolint:errcheck
			return fmt.Errorf("resetting policy tree")
		}

		err = removeAnnotation(ctx, r.Client, &grafana, annotationAppliedNotificationPolicy)
		if err != nil {
			return fmt.Errorf("removing applied notification policy from Grafana CR: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaNotificationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaNotificationPolicy{}).
		Watches(&v1beta1.GrafanaContactPoint{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			log := logf.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
			// resync all notification policies for now. Can be optimized by comparing instance selectors
			nps := &v1beta1.GrafanaNotificationPolicyList{}
			if err := r.List(ctx, nps); err != nil {
				log.Error(err, "failed to fetch notification policies for watch mapping")
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
		Watches(&v1beta1.GrafanaNotificationPolicyRoute{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			log := logf.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
			npr, ok := o.(*v1beta1.GrafanaNotificationPolicyRoute)
			if !ok {
				log.Error(fmt.Errorf("expected object to be NotificationPolicyRoute"), "skipping resource")
			}

			defer func() {
				// update the status
				if err := r.Client.Status().Update(ctx, npr); err != nil {
					log.Error(err, "updating NotificationPolicyRoute status")
				}
			}()

			// check if notification policy route is valid
			if !npr.Spec.Route.IsRouteSelectorMutuallyExclusive() {
				setInvalidSpecMutuallyExclusive(&npr.Status.Conditions, npr.Generation)
				return nil
			}
			removeInvalidSpec(&npr.Status.Conditions)

			// resync all notification policies that have a routeSelector that matches the routes labels
			npList := &v1beta1.GrafanaNotificationPolicyList{}
			if err := r.List(ctx, npList); err != nil {
				log.Error(err, "failed to fetch notification policies for watch mapping")
				return nil
			}
			requests := []reconcile.Request{}
			for _, np := range npList.Items {
				if !np.Spec.Route.HasRouteSelector() {
					continue
				}

				if np.GetNamespace() != npr.GetNamespace() && !np.AllowCrossNamespace() {
					continue
				}

				requests = append(requests,
					reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      np.Name,
							Namespace: np.Namespace,
						},
					})
			}
			return requests
		})).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

// getMatchingNotificationPolicyRoutes retrieves all valid GrafanaNotificationPolicyRoutes for the given labelSelector
// results will be limited to namespace when specified and excludes routes with invalidSpec status condition
func getMatchingNotificationPolicyRoutes(ctx context.Context, k8sClient client.Client, labelSelector *metav1.LabelSelector, namespace *string) ([]v1beta1.GrafanaNotificationPolicyRoute, error) {
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
	if err != nil {
		return nil, err
	}

	// Filter out routes with invalidSpec status condition
	validRoutes := make([]v1beta1.GrafanaNotificationPolicyRoute, 0, len(list.Items))
	for _, route := range list.Items {
		if !meta.IsStatusConditionTrue(route.Status.Conditions, conditionInvalidSpec) {
			validRoutes = append(validRoutes, route)
		}
	}

	return validRoutes, nil
}

// updateNotificationPolicyRoutesStatus sets status conditions and emits a merged event to all matched notification policy routes
func (r *GrafanaNotificationPolicyReconciler) updateNotificationPolicyRoutesStatus(ctx context.Context, notificationPolicy *v1beta1.GrafanaNotificationPolicy, routes []*v1beta1.GrafanaNotificationPolicyRoute) error {
	if notificationPolicy == nil || routes == nil {
		return nil
	}

	for _, route := range routes {
		r.Recorder.Event(route, corev1.EventTypeNormal, "Merged", fmt.Sprintf("Route merged into NotificationPolicy %s/%s", notificationPolicy.GetNamespace(), notificationPolicy.GetName()))

		// Update the status of the route in case conditions have been set
		if err := r.Status().Update(ctx, route); err != nil {
			return fmt.Errorf("failed to update status for route %s/%s: %w", route.Namespace, route.Name, err)
		}
	}

	return nil
}

// statusDiscoveredRoutes returns the list of discovered routes using the namespace and name
// Used to display all discovered routes in the GrafanaNotificationPolicy status
func statusDiscoveredRoutes(routes []*v1beta1.GrafanaNotificationPolicyRoute) []string {
	discoveredRoutes := make([]string, len(routes))
	for i, route := range routes {
		discoveredRoutes[i] = fmt.Sprintf("%s/%s", route.Namespace, route.Name)
	}
	// Reduce status updates by ensuring order of routes
	slices.Sort(discoveredRoutes)

	return discoveredRoutes
}

// setInvalidSpecMutuallyExclusive sets the invalid spec condition due to the routeSelector being mutually exclusive with routes
func setInvalidSpecMutuallyExclusive(conditions *[]metav1.Condition, generation int64) {
	setInvalidSpec(conditions, generation, conditionReasonFieldsMutuallyExclusive, "RouteSelector and Routes are mutually exclusive")
}
