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

	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

type ValidationError struct {
	Err    error
	Reason string
}

func (v *ValidationError) Error() string {
	return v.Err.Error()
}

// GenericReconciler reconciles an arbitrary resource with the corresponding
// api-server resource
type GenericReconciler[T any, PT interface {
	*T
	v1beta1.CommonResource
	v1beta1.CommonSpecResource
}] struct {
	client.Client
	ResourceName          string
	Cfg                   *Config
	GVR                   schema.GroupVersionResource
	Convert               func(context.Context, client.Client, PT) (runtime.Object, error)
	Validate              func(PT) *ValidationError
	PostApplyHook         func(context.Context, client.Client, *v1beta1.Grafana, PT) error
	SynchronizedCondition string
}

func (r *GenericReconciler[T, PT]) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName(fmt.Sprintf("Generic%sReconciler", r.ResourceName))
	ctx = logf.IntoContext(ctx, log)

	cr := PT(new(T))

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, LogMsgGettingCR)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingCR, err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				log.Error(err, LogMsgRunningFinalizer)
				return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgRunningFinalizer, err)
			}

			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				log.Error(err, LogMsgRemoveFinalizer)
				return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgRemoveFinalizer, err)
			}
		}

		return ctrl.Result{}, nil
	}
	defer UpdateStatus(ctx, r.Client, cr, snapshotStatus(cr))

	if cr.CommonSpec().Suspend {
		setSuspended(cr.Conditions(), cr.GetGeneration(), conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(cr.Conditions())

	if err := r.Validate(cr); err != nil {
		setInvalidSpec(cr.Conditions(), cr.GetGeneration(), err.Reason, err.Error())
		meta.RemoveStatusCondition(cr.Conditions(), r.SynchronizedCondition)

		return ctrl.Result{}, err
	}

	removeInvalidSpec(cr.Conditions())

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		if n, ok := any(cr).(v1beta1.NoMatchingInstancesResource); ok {
			n.SetNoMatchingInstances(true)
		}

		setNoMatchingInstancesCondition(cr.Conditions(), cr.GetGeneration(), err)
		meta.RemoveStatusCondition(cr.Conditions(), r.SynchronizedCondition)
		log.Error(err, LogMsgGettingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	if len(instances) == 0 {
		if n, ok := any(cr).(v1beta1.NoMatchingInstancesResource); ok {
			n.SetNoMatchingInstances(true)
		}

		setNoMatchingInstancesCondition(cr.Conditions(), cr.GetGeneration(), err)
		meta.RemoveStatusCondition(cr.Conditions(), r.SynchronizedCondition)
		log.Error(ErrNoMatchingInstances, LogMsgNoMatchingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgNoMatchingInstances, ErrNoMatchingInstances)
	}

	if n, ok := any(cr).(v1beta1.NoMatchingInstancesResource); ok {
		n.SetNoMatchingInstances(false)
	}

	removeNoMatchingInstance(cr.Conditions())
	log.V(1).Info(DbgMsgFoundMatchingInstances, "count", len(instances))

	resource, err := r.Convert(ctx, r, cr)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr, resource)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			continue
		}

		if r.PostApplyHook != nil {
			if err := r.PostApplyHook(ctx, r.Client, &grafana, cr); err != nil {
				applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			}
		}
	}

	condition := buildSynchronizedCondition(r.ResourceName, r.SynchronizedCondition, cr.GetGeneration(), applyErrors, len(instances))
	meta.SetStatusCondition(cr.Conditions(), condition)

	if len(applyErrors) > 0 {
		err = fmt.Errorf(FmtStrApplyErrors, applyErrors)
		log.Error(err, LogMsgApplyErrors)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgApplyErrors, err)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.CommonSpec().ResyncPeriod)}, nil
}

func (r *GenericReconciler[T, PT]) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr PT, resource runtime.Object) error {
	dc, err := grafanaclient.NewDynamicClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	if err := dc.ApplyObject(ctx, resource); err != nil {
		return fmt.Errorf("applying folder resource: %w", err)
	}

	return instance.AddNamespacedResource(ctx, r.Client, cr, v1beta1.NewNamespacedResource(cr.GetNamespace(), cr.GetName(), string(cr.GetUID())))
}

func (r *GenericReconciler[T, PT]) finalize(ctx context.Context, cr PT) error {
	log := logf.FromContext(ctx)

	uid := cr.GetUID()

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgGettingInstances)
		return fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	for _, grafana := range instances {
		gClient, err := grafanaclient.NewDynamicClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		err = gClient.DeleteInDefaultNamespace(ctx, r.GVR, string(uid))
		if err != nil {
			return err
		}

		// Update grafana instance Status
		err = grafana.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return fmt.Errorf("removing resource from Grafana cr: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GenericReconciler[T, PT]) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(PT(new(T))).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
