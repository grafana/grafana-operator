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
	"strings"
	"time"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
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
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafananotificationtemplates/finalizers,verbs=update

func (r *GrafanaNotificationTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx).WithName("GrafanaNotificationTemplateReconciler")
	r.Log = log.FromContext(ctx)

	notificationTemplate := &grafanav1beta1.GrafanaNotificationTemplate{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, notificationTemplate)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "Failed to get GrafanaNotificationTemplate")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if notificationTemplate.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(notificationTemplate, grafanaFinalizer) {
			err := r.finalize(ctx, notificationTemplate)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to finalize GrafanaNotificationTemplate: %w", err)
			}
			controllerutil.RemoveFinalizer(notificationTemplate, grafanaFinalizer)
			if err := r.Update(ctx, notificationTemplate); err != nil {
				r.Log.Error(err, "failed to remove finalizer")
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to update GrafanaNotificationTemplate: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	defer func() {
		if err := r.Client.Status().Update(ctx, notificationTemplate); err != nil {
			r.Log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(notificationTemplate.Status.Conditions, conditionNoMatchingInstance) {
			controllerutil.RemoveFinalizer(notificationTemplate, grafanaFinalizer)
		} else {
			controllerutil.AddFinalizer(notificationTemplate, grafanaFinalizer)
		}
		if err := r.Update(ctx, notificationTemplate); err != nil {
			r.Log.Error(err, "failed to set finalizer")
		}
	}()

	instances, err := r.GetMatchingInstances(ctx, notificationTemplate, r.Client)
	if err != nil {
		setNoMatchingInstance(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, "ErrFetchingInstances", fmt.Sprintf("error occurred during fetching of instances: %s", err.Error()))
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)
		r.Log.Error(err, "could not find matching instances")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if len(instances) == 0 {
		meta.RemoveStatusCondition(&notificationTemplate.Status.Conditions, conditionNotificationTemplateSynchronized)
		setNoMatchingInstance(&notificationTemplate.Status.Conditions, notificationTemplate.Generation, "EmptyAPIReply", "Instances could not be fetched, reconciliation will be retried")
		return ctrl.Result{}, nil
	}

	removeNoMatchingInstance(&notificationTemplate.Status.Conditions)

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		// can be removed in go 1.22+
		grafana := grafana
		if grafana.Status.Stage != grafanav1beta1.OperatorStageComplete || grafana.Status.StageStatus != grafanav1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		err := r.reconcileWithInstance(ctx, &grafana, notificationTemplate)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}
	condition := metav1.Condition{
		Type:               conditionNotificationTemplateSynchronized,
		ObservedGeneration: notificationTemplate.Generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	}

	if len(applyErrors) == 0 {
		condition.Status = metav1.ConditionTrue
		condition.Reason = conditionApplySuccessful
		condition.Message = fmt.Sprintf("Notification template was successfully applied to %d instances", len(instances))
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Reason = conditionApplyFailed

		var sb strings.Builder
		for i, err := range applyErrors {
			sb.WriteString(fmt.Sprintf("\n- %s: %s", i, err))
		}

		condition.Message = fmt.Sprintf("Notification template failed to be applied for %d out of %d instances. Errors:%s", len(applyErrors), len(instances), sb.String())
	}
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

func (r *GrafanaNotificationTemplateReconciler) getNotificationTemplateByName(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) (models.NotificationTemplate, error) {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return models.NotificationTemplate{}, fmt.Errorf("building grafana client: %w", err)
	}

	params := provisioning.NewGetTemplateParams().WithName(notificationTemplate.Spec.Name)
	remote, err := cl.Provisioning.GetTemplateWithParams(params)
	if err != nil {
		return models.NotificationTemplate{}, fmt.Errorf("getting notification template: %w", err)
	}
	return *remote.Payload, nil
}

func (r *GrafanaNotificationTemplateReconciler) finalize(ctx context.Context, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	r.Log.Info("Finalizing GrafanaNotificationTemplate")

	instances, err := r.GetMatchingInstances(ctx, notificationTemplate, r.Client)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances {
		instance := i
		if err := r.removeFromInstance(ctx, &instance, notificationTemplate); err != nil {
			return fmt.Errorf("removing contact point from instance: %w", err)
		}
	}

	return nil
}

func (r *GrafanaNotificationTemplateReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = r.getNotificationTemplateByName(ctx, instance, notificationTemplate)
	if err != nil {
		return fmt.Errorf("getting notification template by name: %w", err)
	}
	_, err = cl.Provisioning.DeleteTemplate(notificationTemplate.Spec.Name) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("deleting notification template: %w", err)
	}

	return nil
}

func (r *GrafanaNotificationTemplateReconciler) GetMatchingInstances(ctx context.Context, notificationTemplate *grafanav1beta1.GrafanaNotificationTemplate, k8sClient client.Client) ([]grafanav1beta1.Grafana, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, notificationTemplate.Spec.InstanceSelector)
	if err != nil || len(instances.Items) == 0 {
		return nil, err
	}
	if notificationTemplate.Spec.AllowCrossNamespaceImport != nil && *notificationTemplate.Spec.AllowCrossNamespaceImport {
		return instances.Items, nil
	}
	items := []grafanav1beta1.Grafana{}
	for _, i := range instances.Items {
		if i.Namespace == notificationTemplate.Namespace {
			items = append(items, i)
		}
	}

	return items, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaNotificationTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaNotificationTemplate{}).
		Complete(r)
}
