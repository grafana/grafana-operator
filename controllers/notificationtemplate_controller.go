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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionNotificationTemplateSynchronized = "NotificationTemplateSynchronized"
)

// GrafanaNotificationTemplateReconciler reconciles a GrafanaNotificationTemplate object
type GrafanaNotificationTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaNotificationTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaNotificationTemplateReconciler")
	ctx = logf.IntoContext(ctx, log)

	notificationTemplate := &grafanav1beta1.GrafanaNotificationTemplate{}

	err := r.Get(ctx, req.NamespacedName, notificationTemplate)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get GrafanaNotificationTemplate: %w", err)
	}

	if notificationTemplate.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(notificationTemplate, grafanaFinalizer) {
			if err := r.finalize(ctx, notificationTemplate); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaNotificationTemplate: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, notificationTemplate); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, notificationTemplate)

	if notificationTemplate.Spec.Suspend {
		setSuspended(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&notificationTemplate.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationTemplate)
	if err != nil {
		setNoMatchingInstancesCondition(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, err)
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, err)
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&notificationTemplate.Status.Conditions)
	log.Info("found matching Grafana instances for notification template", "count", len(instances))

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, notificationTemplate)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Notification template", conditionNotificationTemplateSynchronized, notificationTemplate.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&notificationTemplate.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(notificationTemplate.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaNotificationTemplateReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	trueRef := "true" //nolint:goconst

	editable := true //nolint:staticcheck
	if notificationTemplate.Spec.Editable != nil && !*notificationTemplate.Spec.Editable {
		editable = false
	}

	var updatedNT models.NotificationTemplateContent

	updatedNT.Template = notificationTemplate.Spec.Template

	params := provisioning.NewPutTemplateParams().WithName(notificationTemplate.Spec.Name).WithBody(&updatedNT)
	if editable {
		params.SetXDisableProvenance(&trueRef)
	}

	_, err = cl.Provisioning.PutTemplate(params) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("creating or updating notification template: %w", err)
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, notificationTemplate, notificationTemplate.NamespacedResource())
}

func (r *GrafanaNotificationTemplateReconciler) finalize(ctx context.Context, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaNotificationTemplate")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationTemplate)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		if err := r.removeFromInstance(ctx, &instance, notificationTemplate); err != nil {
			return fmt.Errorf("removing notification template from instance: %w", err)
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, notificationTemplate)
		if err != nil {
			return fmt.Errorf("removing notification template from Grafana cr: %w", err)
		}
	}

	return nil
}

func (r *GrafanaNotificationTemplateReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = cl.Provisioning.DeleteTemplate(&provisioning.DeleteTemplateParams{Name: notificationTemplate.Spec.Name}) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("deleting notification template: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaNotificationTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaNotificationTemplate{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
