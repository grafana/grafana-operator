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
	conditionContactPointSynchronized = "ContactPointSynchronized"
)

// GrafanaContactPointReconciler reconciles a GrafanaContactPoint object
type GrafanaContactPointReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanacontactpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanacontactpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanacontactpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GrafanaContactPoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *GrafanaContactPointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx).WithName("GrafanaContactPointReconciler")
	r.Log = log.FromContext(ctx)

	contactPoint := &grafanav1beta1.GrafanaContactPoint{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, contactPoint)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "Failed to get GrafanaContactPoint")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if contactPoint.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(contactPoint, grafanaFinalizer) {
			err := r.finalize(ctx, contactPoint)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to finalize GrafanaContactPoint: %w", err)
			}
			controllerutil.RemoveFinalizer(contactPoint, grafanaFinalizer)
			if err := r.Update(ctx, contactPoint); err != nil {
				r.Log.Error(err, "failed to remove finalizer")
				return ctrl.Result{RequeueAfter: RequeueDelay}, fmt.Errorf("failed to update GrafanaContactPoint: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	defer func() {
		if err := r.Client.Status().Update(ctx, contactPoint); err != nil {
			r.Log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(contactPoint.Status.Conditions, conditionNoMatchingInstance) {
			controllerutil.RemoveFinalizer(contactPoint, grafanaFinalizer)
		} else {
			controllerutil.AddFinalizer(contactPoint, grafanaFinalizer)
		}
		if err := r.Update(ctx, contactPoint); err != nil {
			r.Log.Error(err, "failed to set finalizer")
		}
	}()

	instances, err := r.GetMatchingInstances(ctx, contactPoint, r.Client)
	if err != nil {
		setNoMatchingInstance(&contactPoint.Status.Conditions, contactPoint.Generation, "ErrFetchingInstances", fmt.Sprintf("error occurred during fetching of instances: %s", err.Error()))
		meta.RemoveStatusCondition(&contactPoint.Status.Conditions, conditionContactPointSynchronized)
		r.Log.Error(err, "could not find matching instances")
		return ctrl.Result{RequeueAfter: RequeueDelay}, err
	}

	if len(instances) == 0 {
		meta.RemoveStatusCondition(&contactPoint.Status.Conditions, conditionContactPointSynchronized)
		setNoMatchingInstance(&contactPoint.Status.Conditions, contactPoint.Generation, "EmptyAPIReply", "Instances could not be fetched, reconciliation will be retried")
		return ctrl.Result{}, nil
	}

	removeNoMatchingInstance(&contactPoint.Status.Conditions)

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		// can be removed in go 1.22+
		grafana := grafana
		if grafana.Status.Stage != grafanav1beta1.OperatorStageComplete || grafana.Status.StageStatus != grafanav1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		err := r.reconcileWithInstance(ctx, &grafana, contactPoint)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}
	condition := metav1.Condition{
		Type:               conditionContactPointSynchronized,
		ObservedGeneration: contactPoint.Generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	}

	if len(applyErrors) == 0 {
		condition.Status = "True"
		condition.Reason = "ApplySuccesfull"
		condition.Message = fmt.Sprintf("Contact point was successfully applied to %d instances", len(instances))
	} else {
		condition.Status = "False"
		condition.Reason = "ApplyFailed"

		var sb strings.Builder
		for i, err := range applyErrors {
			sb.WriteString(fmt.Sprintf("\n- %s: %s", i, err))
		}

		condition.Message = fmt.Sprintf("Contact point failed to be applied for %d out of %d instances. Errors:%s", len(applyErrors), len(instances), sb.String())
	}
	meta.SetStatusCondition(&contactPoint.Status.Conditions, condition)

	return ctrl.Result{RequeueAfter: contactPoint.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaContactPointReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, contactPoint *grafanav1beta1.GrafanaContactPoint) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	var applied models.EmbeddedContactPoint

	applied, err = r.getContactPointFromUID(ctx, instance, contactPoint)
	if err != nil {
		return fmt.Errorf("getting contact point by UID: %w", err)
	}

	if applied.UID == "" {
		// create
		cp := &models.EmbeddedContactPoint{
			DisableResolveMessage: contactPoint.Spec.DisableResolveMessage,
			Name:                  contactPoint.Spec.Name,
			Type:                  &contactPoint.Spec.Type,
			Settings:              contactPoint.Spec.Settings,
			UID:                   string(contactPoint.UID),
		}
		_, err := cl.Provisioning.PostContactpoints(provisioning.NewPostContactpointsParams().WithBody(cp)) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("creating contact point: %w", err)
		}
	} else {
		// update
		var updatedCP models.EmbeddedContactPoint
		updatedCP.Name = contactPoint.Spec.Name
		updatedCP.Type = &contactPoint.Spec.Type
		updatedCP.Settings = contactPoint.Spec.Settings
		_, err := cl.Provisioning.PutContactpoint(provisioning.NewPutContactpointParams().WithUID(applied.UID).WithBody(&updatedCP)) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("updating contact point: %w", err)
		}
	}
	return nil
}

func (r *GrafanaContactPointReconciler) getContactPointFromUID(ctx context.Context, instance *grafanav1beta1.Grafana, contactPoint *grafanav1beta1.GrafanaContactPoint) (models.EmbeddedContactPoint, error) {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return models.EmbeddedContactPoint{}, fmt.Errorf("building grafana client: %w", err)
	}

	params := provisioning.NewGetContactpointsParams()
	remote, err := cl.Provisioning.GetContactpoints(params)
	if err != nil {
		return models.EmbeddedContactPoint{}, fmt.Errorf("getting contact points: %w", err)
	}
	for _, cp := range remote.Payload {
		if cp.UID == string(contactPoint.UID) {
			return *cp, nil
		}
	}
	return models.EmbeddedContactPoint{}, nil
}

func (r *GrafanaContactPointReconciler) finalize(ctx context.Context, contactPoint *grafanav1beta1.GrafanaContactPoint) error {
	r.Log.Info("Finalizing GrafanaContactPoint")

	instances, err := r.GetMatchingInstances(ctx, contactPoint, r.Client)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances {
		instance := i
		if err := r.removeFromInstance(ctx, &instance, contactPoint); err != nil {
			return fmt.Errorf("removing contact point from instance: %w", err)
		}
	}

	return nil
}

func (r *GrafanaContactPointReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, contactPoint *grafanav1beta1.GrafanaContactPoint) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = r.getContactPointFromUID(ctx, instance, contactPoint)
	if err != nil {
		return fmt.Errorf("getting contact point by UID: %w", err)
	}
	_, err = cl.Provisioning.DeleteContactpoints(string(contactPoint.UID)) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("deleting contact point: %w", err)
	}

	return nil
}

func (r *GrafanaContactPointReconciler) GetMatchingInstances(ctx context.Context, contactPoint *grafanav1beta1.GrafanaContactPoint, k8sClient client.Client) ([]grafanav1beta1.Grafana, error) {
	instances, err := GetMatchingInstances(ctx, k8sClient, contactPoint.Spec.InstanceSelector)
	if err != nil || len(instances.Items) == 0 {
		return nil, err
	}
	if contactPoint.Spec.AllowCrossNamespaceImport != nil && *contactPoint.Spec.AllowCrossNamespaceImport {
		return instances.Items, nil
	}
	items := []grafanav1beta1.Grafana{}
	for _, i := range instances.Items {
		if i.Namespace == contactPoint.Namespace {
			items = append(items, i)
		}
	}

	return items, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaContactPointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaContactPoint{}).
		Complete(r)
}
