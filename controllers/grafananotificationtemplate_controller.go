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
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates/finalizers,verbs=update

func (r *GrafanaNotificationTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaNotificationTemplateReconciler")
	logf.IntoContext(ctx, log)

	notificationTemplate := &grafanav1beta1.GrafanaNotificationTemplate{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, notificationTemplate)
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

	defer func() {
		notificationTemplate.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Client.Status().Update(ctx, notificationTemplate); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(notificationTemplate.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, notificationTemplate); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, notificationTemplate); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationTemplate)
	if err != nil {
		setNoMatchingInstancesCondition(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, err)
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, err)
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&notificationTemplate.Status.Conditions)
	log.Info("found matching Grafana instances for notification template", "count", len(instances))

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		// can be removed in go 1.22+
		grafana := grafana

		err := r.reconcileWithInstance(ctx, &grafana, notificationTemplate)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}
	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	condition := buildSynchronizedCondition("Notification template", conditionNotificationTemplateSynchronized, notificationTemplate.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&notificationTemplate.Status.Conditions, condition)
	return ctrl.Result{RequeueAfter: notificationTemplate.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaNotificationTemplateReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	trueRef := "true" //nolint:goconst
	editable := true
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
	return nil
}

func (r *GrafanaNotificationTemplateReconciler) finalize(ctx context.Context, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaNotificationTemplate")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, notificationTemplate)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances {
		instance := i
		if err := r.removeFromInstance(ctx, &instance, notificationTemplate); err != nil {
			return fmt.Errorf("removing notification template from instance: %w", err)
		}
	}

	return nil
}

func (r *GrafanaNotificationTemplateReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = cl.Provisioning.DeleteTemplate(notificationTemplate.Spec.Name) //nolint:errcheck
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
