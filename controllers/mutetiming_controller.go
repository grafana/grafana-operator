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
	conditionMuteTimingSynchronized = "MuteTimingSynchronized"
)

// GrafanaMuteTimingReconciler reconciles a GrafanaMuteTiming object
type GrafanaMuteTimingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaMuteTimingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaMuteTimingReconciler")
	ctx = logf.IntoContext(ctx, log)

	muteTiming := &grafanav1beta1.GrafanaMuteTiming{}

	err := r.Get(ctx, req.NamespacedName, muteTiming)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get GrafanaMuteTiming: %w", err)
	}

	if muteTiming.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(muteTiming, grafanaFinalizer) {
			if err := r.finalize(ctx, muteTiming); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaMuteTiming: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, muteTiming); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, muteTiming)

	if muteTiming.Spec.Suspend {
		setSuspended(&muteTiming.Status.Conditions, muteTiming.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&muteTiming.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, muteTiming)
	if err != nil {
		setNoMatchingInstancesCondition(&muteTiming.Status.Conditions, muteTiming.Generation, err)
		meta.RemoveStatusCondition(&muteTiming.Status.Conditions, conditionMuteTimingSynchronized)

		return ctrl.Result{}, fmt.Errorf("could not find matching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&muteTiming.Status.Conditions, muteTiming.Generation, err)
		meta.RemoveStatusCondition(&muteTiming.Status.Conditions, conditionMuteTimingSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&muteTiming.Status.Conditions)
	log.Info("found matching Grafana instances for mute timing", "count", len(instances))

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, muteTiming)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Mute timing", conditionMuteTimingSynchronized, muteTiming.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&muteTiming.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(muteTiming.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaMuteTimingReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, muteTiming *grafanav1beta1.GrafanaMuteTiming) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = r.getMuteTimingByName(ctx, muteTiming.Spec.Name, instance)

	shouldCreate := false
	if errors.Is(err, provisioning.NewGetMuteTimingNotFound()) {
		shouldCreate = true
	} else if err != nil {
		return fmt.Errorf("getting mute timing by name: %w", err)
	}

	trueRef := "true" //nolint:goconst

	var payload models.MuteTimeInterval

	payload.Name = muteTiming.Spec.Name

	payload.TimeIntervals = make([]*models.TimeIntervalItem, 0, len(muteTiming.Spec.TimeIntervals))
	for _, ti := range muteTiming.Spec.TimeIntervals {
		times := make([]*models.TimeIntervalTimeRange, 0, len(ti.Times))
		for _, tr := range ti.Times {
			times = append(times, &models.TimeIntervalTimeRange{
				StartTime: tr.StartTime,
				EndTime:   tr.EndTime,
			})
		}

		payload.TimeIntervals = append(payload.TimeIntervals, &models.TimeIntervalItem{
			DaysOfMonth: ti.DaysOfMonth,
			Location:    ti.Location,
			Months:      ti.Months,
			Weekdays:    ti.Weekdays,
			Years:       ti.Years,
			Times:       times,
		})
	}

	if shouldCreate {
		params := provisioning.NewPostMuteTimingParams().WithBody(&payload)
		if muteTiming.Spec.Editable {
			params.SetXDisableProvenance(&trueRef)
		}

		_, err = cl.Provisioning.PostMuteTiming(params) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("creating mute timing: %w", err)
		}
	} else {
		params := provisioning.NewPutMuteTimingParams().WithName(muteTiming.Spec.Name).WithBody(&payload)
		if muteTiming.Spec.Editable {
			params.SetXDisableProvenance(&trueRef)
		}

		_, err = cl.Provisioning.PutMuteTiming(params) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("updating mute timing: %w", err)
		}
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, muteTiming, muteTiming.NamespacedResource())
}

func (r *GrafanaMuteTimingReconciler) getMuteTimingByName(ctx context.Context, name string, instance *grafanav1beta1.Grafana) (*models.MuteTimeInterval, error) {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return nil, fmt.Errorf("building grafana client: %w", err)
	}

	muteTiming, err := cl.Provisioning.GetMuteTiming(name)
	if err != nil {
		return nil, fmt.Errorf("getting mute timing: %w", err)
	}

	return muteTiming.Payload, nil
}

func (r *GrafanaMuteTimingReconciler) finalize(ctx context.Context, muteTiming *grafanav1beta1.GrafanaMuteTiming) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaMuteTiming")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, muteTiming)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		if err := r.removeFromInstance(ctx, &instance, muteTiming); err != nil {
			return fmt.Errorf("removing mute timing from instance: %w", err)
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, muteTiming)
		if err != nil {
			return fmt.Errorf("removing mute timings from Grafana cr: %w", err)
		}
	}

	return nil
}

func (r *GrafanaMuteTimingReconciler) removeFromInstance(ctx context.Context, instance *grafanav1beta1.Grafana, muteTiming *grafanav1beta1.GrafanaMuteTiming) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = cl.Provisioning.DeleteMuteTiming(&provisioning.DeleteMuteTimingParams{Name: muteTiming.Spec.Name}) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("deleting mute timing: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaMuteTimingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaMuteTiming{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
