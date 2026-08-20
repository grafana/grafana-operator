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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
)

const (
	conditionSilenceSynchronized  = "SilenceSynchronized"
	conditionReasonExpiredSilence = "ExpiredSilence"
)

// GrafanaSilenceReconciler reconciles a GrafanaSilence object
type GrafanaSilenceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaSilenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaSilenceReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaSilence{}

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

	defer UpdateStatus(ctx, r.Client, cr)

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	// Kubernetes CEL validation cannot reference the current time, so enforce that the
	// silence has not already expired here rather than in the CRD schema.
	if !cr.Spec.EndsAt.After(time.Now()) {
		err := fmt.Errorf("spec.endsAt %s must be in the future", cr.Spec.EndsAt.UTC().Format(time.RFC3339))
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, conditionReasonExpiredSilence, err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionSilenceSynchronized)
		log.Error(err, "refusing to apply expired silence")

		return ctrl.Result{}, err
	}

	removeInvalidSpec(&cr.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionSilenceSynchronized)
		log.Error(err, LogMsgGettingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionSilenceSynchronized)
		log.Error(ErrNoMatchingInstances, LogMsgNoMatchingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgNoMatchingInstances, ErrNoMatchingInstances)
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.V(1).Info(DbgMsgFoundMatchingInstances, "count", len(instances))

	// The silence ID assigned by Grafana differs per instance and is tracked in an
	// annotation as a JSON map of "<namespace>/<name>" instance -> silence ID.
	silenceIDs, err := getSilenceIDs(cr)
	if err != nil {
		log.Error(err, "failed to read silence ID annotation, treating as empty")

		silenceIDs = map[string]string{}
	}

	originalIDs := maps.Clone(silenceIDs)

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr, silenceIDs)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	// Persist any newly assigned silence IDs back to the annotation
	if !maps.Equal(originalIDs, silenceIDs) {
		if err := r.updateSilenceIDs(ctx, cr, silenceIDs); err != nil {
			log.Error(err, "failed to persist silence ID annotation")
			applyErrors[fmt.Sprintf("%s/%s", cr.Namespace, cr.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Silence", conditionSilenceSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		err = fmt.Errorf(FmtStrApplyErrors, applyErrors)
		log.Error(err, LogMsgApplyErrors)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgApplyErrors, err)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaSilenceReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaSilence, silenceIDs map[string]string) error {
	key := fmt.Sprintf("%s/%s", instance.Namespace, instance.Name)

	payload := grafanaclient.PostableSilence{
		StartsAt:  cr.Spec.StartsAt.UTC().Format(time.RFC3339),
		EndsAt:    cr.Spec.EndsAt.UTC().Format(time.RFC3339),
		Comment:   cr.Spec.Comment,
		CreatedBy: cr.Spec.CreatedBy,
		Matchers:  make([]grafanaclient.SilenceMatcher, 0, len(cr.Spec.Matchers)),
	}

	for _, m := range cr.Spec.Matchers {
		payload.Matchers = append(payload.Matchers, grafanaclient.SilenceMatcher{
			Name:    m.Name,
			Value:   m.Value,
			IsRegex: m.IsRegex,
			IsEqual: m.IsEqual,
		})
	}

	// Update the existing silence in place when we already track an active one for this
	// instance. An expired or missing silence is recreated.
	if id := silenceIDs[key]; id != "" {
		existing, err := grafanaclient.GetSilence(ctx, r.Client, instance, id)
		if err != nil {
			return fmt.Errorf("getting silence: %w", err)
		}

		if existing != nil && existing.Status.State != "expired" {
			payload.ID = id
		}
	}

	id, err := grafanaclient.CreateOrUpdateSilence(ctx, r.Client, instance, payload)
	if err != nil {
		return fmt.Errorf("applying silence: %w", err)
	}

	silenceIDs[key] = id

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaSilenceReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaSilence) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaSilence")

	silenceIDs, err := getSilenceIDs(cr)
	if err != nil {
		log.Error(err, "failed to read silence ID annotation during finalize")

		silenceIDs = map[string]string{}
	}

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgGettingInstances)
		return fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	for _, instance := range instances {
		key := fmt.Sprintf("%s/%s", instance.Namespace, instance.Name)

		if id := silenceIDs[key]; id != "" {
			if err := grafanaclient.DeleteSilence(ctx, r.Client, &instance, id); err != nil {
				return fmt.Errorf("deleting silence from instance: %w", err)
			}
		}

		// Update grafana instance Status
		if err := instance.RemoveNamespacedResource(ctx, r.Client, cr); err != nil {
			return fmt.Errorf("removing silence from Grafana cr: %w", err)
		}
	}

	return nil
}

// getSilenceIDs decodes the silence ID annotation into a map of instance key -> silence ID.
func getSilenceIDs(cr *v1beta1.GrafanaSilence) (map[string]string, error) {
	ids := map[string]string{}

	raw, ok := cr.Annotations[v1beta1.SilenceIDAnnotation]
	if !ok || raw == "" {
		return ids, nil
	}

	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return map[string]string{}, fmt.Errorf("parsing %s annotation: %w", v1beta1.SilenceIDAnnotation, err)
	}

	return ids, nil
}

// updateSilenceIDs writes the silence ID map back to the annotation. Annotation changes do
// not bump metadata.generation, so this does not trigger an additional reconcile.
func (r *GrafanaSilenceReconciler) updateSilenceIDs(ctx context.Context, cr *v1beta1.GrafanaSilence, silenceIDs map[string]string) error {
	encoded, err := json.Marshal(silenceIDs)
	if err != nil {
		return fmt.Errorf("encoding silence IDs: %w", err)
	}

	patchBase := client.MergeFrom(cr.DeepCopy())

	if cr.Annotations == nil {
		cr.Annotations = map[string]string{}
	}

	cr.Annotations[v1beta1.SilenceIDAnnotation] = string(encoded)

	return r.Patch(ctx, cr, patchBase)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaSilenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaSilence{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
