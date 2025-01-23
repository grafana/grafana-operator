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
	"time"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionNotificationPolicySynchronized = "NotificationPolicySynchronized"
	annotationAppliedNotificationPolicy     = "operator.grafana.com/applied-notificationpolicy"
)

// GrafanaNotificationPolicyReconciler reconciles a GrafanaNotificationPolicy object
type GrafanaNotificationPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationpolicies/finalizers,verbs=update

func (r *GrafanaNotificationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
	logf.IntoContext(ctx, log)

	notificationPolicy := &grafanav1beta1.GrafanaNotificationPolicy{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, notificationPolicy)
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

	defer func() {
		notificationPolicy.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Client.Status().Update(ctx, notificationPolicy); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(notificationPolicy.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, notificationPolicy); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, notificationPolicy); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationPolicy)
	if err != nil {
		setNoMatchingInstancesCondition(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, err)
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&notificationPolicy.Status.Conditions, notificationPolicy.Generation, err)
		meta.RemoveStatusCondition(&notificationPolicy.Status.Conditions, conditionNotificationPolicySynchronized)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&notificationPolicy.Status.Conditions)
	log.Info("found matching Grafana instances for notificationPolicy", "count", len(instances))

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		// can be removed in go 1.22+
		grafana := grafana

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
	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	condition := buildSynchronizedCondition("Notification Policy", conditionNotificationPolicySynchronized, notificationPolicy.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&notificationPolicy.Status.Conditions, condition)

	return ctrl.Result{RequeueAfter: notificationPolicy.Spec.ResyncPeriod.Duration}, nil
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

func (r *GrafanaNotificationPolicyReconciler) finalize(ctx context.Context, notificationPolicy *grafanav1beta1.GrafanaNotificationPolicy) error {
	log := logf.FromContext(ctx)
	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationPolicy)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, grafana := range instances {
		grafana := grafana

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

		delete(grafana.Annotations, annotationAppliedNotificationPolicy)
		if err := r.Client.Update(ctx, &grafana); err != nil {
			return fmt.Errorf("removing applied notification policy from Grafana cr: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaNotificationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaNotificationPolicy{}).
		Watches(&grafanav1beta1.GrafanaContactPoint{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			log := logf.FromContext(ctx).WithName("GrafanaNotificationPolicyReconciler")
			// resync all notification policies for now. Can be optimized by comparing instance selectors
			nps := &grafanav1beta1.GrafanaNotificationPolicyList{}
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
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
