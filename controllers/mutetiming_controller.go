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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/pkg/ptr"
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

	cr := &v1beta1.GrafanaMuteTiming{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get GrafanaMuteTiming: %w", err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaMuteTiming: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, cr)

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionMuteTimingSynchronized)

		return ctrl.Result{}, fmt.Errorf("could not find matching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionMuteTimingSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.Info("found matching Grafana instances for mute timing", "count", len(instances))

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Mute timing", conditionMuteTimingSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaMuteTimingReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaMuteTiming) error {
	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = r.getMuteTimingByName(ctx, cr.Spec.Name, instance)

	shouldCreate := false
	if errors.Is(err, provisioning.NewGetMuteTimingNotFound()) {
		shouldCreate = true
	} else if err != nil {
		return fmt.Errorf("getting mute timing by name: %w", err)
	}

	refTrue := ptr.To("true")

	var payload models.MuteTimeInterval

	payload.Name = cr.Spec.Name

	payload.TimeIntervals = make([]*models.TimeIntervalItem, 0, len(cr.Spec.TimeIntervals))
	for _, ti := range cr.Spec.TimeIntervals {
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
		if cr.Spec.Editable {
			params.SetXDisableProvenance(refTrue)
		}

		_, err = gClient.Provisioning.PostMuteTiming(params) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("creating mute timing: %w", err)
		}
	} else {
		params := provisioning.NewPutMuteTimingParams().WithName(cr.Spec.Name).WithBody(&payload)
		if cr.Spec.Editable {
			params.SetXDisableProvenance(refTrue)
		}

		_, err = gClient.Provisioning.PutMuteTiming(params) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("updating mute timing: %w", err)
		}
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaMuteTimingReconciler) getMuteTimingByName(ctx context.Context, name string, instance *v1beta1.Grafana) (*models.MuteTimeInterval, error) {
	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return nil, fmt.Errorf("building grafana client: %w", err)
	}

	remoteMuteTiming, err := gClient.Provisioning.GetMuteTiming(name)
	if err != nil {
		return nil, fmt.Errorf("getting mute timing: %w", err)
	}

	return remoteMuteTiming.Payload, nil
}

func (r *GrafanaMuteTimingReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaMuteTiming) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaMuteTiming")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		if err := r.removeFromInstance(ctx, &instance, cr); err != nil {
			return fmt.Errorf("removing mute timing from instance: %w", err)
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return fmt.Errorf("removing mute timings from Grafana cr: %w", err)
		}
	}

	return nil
}

func (r *GrafanaMuteTimingReconciler) removeFromInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaMuteTiming) error {
	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	_, err = gClient.Provisioning.DeleteMuteTiming(&provisioning.DeleteMuteTimingParams{Name: cr.Spec.Name}) //nolint:errcheck
	if err != nil {
		return fmt.Errorf("deleting mute timing: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaMuteTimingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaMuteTiming{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
