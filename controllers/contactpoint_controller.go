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
	"strings"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	simplejson "github.com/bitly/go-simplejson"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	corev1 "k8s.io/api/core/v1"
)

const (
	conditionContactPointSynchronized = "ContactPointSynchronized"
	conditionReasonInvalidSettings    = "InvalidSettings"
)

// GrafanaContactPointReconciler reconciles a GrafanaContactPoint object
type GrafanaContactPointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaContactPointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaContactPointReconciler")
	ctx = logf.IntoContext(ctx, log)

	contactPoint := &grafanav1beta1.GrafanaContactPoint{}

	err := r.Get(ctx, req.NamespacedName, contactPoint)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting grafana Contact point cr: %w", err)
	}

	if contactPoint.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(contactPoint, grafanaFinalizer) {
			if err := r.finalize(ctx, contactPoint); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaContactPoint: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, contactPoint); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, contactPoint)

	if contactPoint.Spec.Suspend {
		setSuspended(&contactPoint.Status.Conditions, contactPoint.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&contactPoint.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, contactPoint)
	if err != nil {
		setNoMatchingInstancesCondition(&contactPoint.Status.Conditions, contactPoint.Generation, err)
		meta.RemoveStatusCondition(&contactPoint.Status.Conditions, conditionContactPointSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&contactPoint.Status.Conditions, contactPoint.Generation, err)
		meta.RemoveStatusCondition(&contactPoint.Status.Conditions, conditionContactPointSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&contactPoint.Status.Conditions)
	log.Info("found matching Grafana instances for Contact point", "count", len(instances))

	settings, err := r.buildContactPointSettings(ctx, contactPoint)
	if err != nil {
		setInvalidSpec(&contactPoint.Status.Conditions, contactPoint.Generation, conditionReasonInvalidSettings, err.Error())
		meta.RemoveStatusCondition(&contactPoint.Status.Conditions, conditionContactPointSynchronized)

		return ctrl.Result{}, fmt.Errorf("building contactpoint settings: %w", err)
	}

	removeInvalidSpec(&contactPoint.Status.Conditions)

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, contactPoint, &settings)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Contact point", conditionContactPointSynchronized, contactPoint.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&contactPoint.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(contactPoint.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaContactPointReconciler) reconcileWithInstance(ctx context.Context, instance *grafanav1beta1.Grafana, contactPoint *grafanav1beta1.GrafanaContactPoint, settings *models.JSON) error {
	cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana client: %w", err)
	}

	var applied models.EmbeddedContactPoint

	applied, err = r.getContactPointFromUID(cl, contactPoint)
	if err != nil {
		return fmt.Errorf("getting contact point by UID: %w", err)
	}

	if applied.UID == "" {
		// create
		cp := &models.EmbeddedContactPoint{
			DisableResolveMessage: contactPoint.Spec.DisableResolveMessage,
			Name:                  contactPoint.Spec.Name,
			Type:                  &contactPoint.Spec.Type,
			Settings:              settings,
			UID:                   contactPoint.CustomUIDOrUID(),
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
		updatedCP.Settings = settings
		updatedCP.DisableResolveMessage = contactPoint.Spec.DisableResolveMessage

		_, err := cl.Provisioning.PutContactpoint(provisioning.NewPutContactpointParams().WithUID(applied.UID).WithBody(&updatedCP)) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("updating contact point: %w", err)
		}
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, contactPoint, contactPoint.NamespacedResource())
}

func (r *GrafanaContactPointReconciler) buildContactPointSettings(ctx context.Context, contactPoint *grafanav1beta1.GrafanaContactPoint) (models.JSON, error) {
	log := logf.FromContext(ctx)

	marshaled, err := json.Marshal(contactPoint.Spec.Settings)
	if err != nil {
		return nil, fmt.Errorf("encoding existing settings as json: %w", err)
	}

	simpleContent, err := simplejson.NewJson(marshaled)
	if err != nil {
		return nil, fmt.Errorf("parsing marshaled json as simplejson")
	}

	for _, override := range contactPoint.Spec.ValuesFrom {
		val, _, err := getReferencedValue(ctx, r.Client, contactPoint, override.ValueFrom)
		if err != nil {
			return nil, fmt.Errorf("getting referenced value: %w", err)
		}

		log.V(1).Info("overriding value", "key", override.TargetPath, "value", val)

		simpleContent.SetPath(strings.Split(override.TargetPath, "."), val)
	}

	return simpleContent.Interface(), nil
}

func (r *GrafanaContactPointReconciler) getContactPointFromUID(cl *genapi.GrafanaHTTPAPI, contactPoint *grafanav1beta1.GrafanaContactPoint) (models.EmbeddedContactPoint, error) {
	params := provisioning.NewGetContactpointsParams()

	remote, err := cl.Provisioning.GetContactpoints(params)
	if err != nil {
		return models.EmbeddedContactPoint{}, fmt.Errorf("getting contact points: %w", err)
	}

	for _, cp := range remote.Payload {
		if cp.UID == contactPoint.CustomUIDOrUID() {
			return *cp, nil
		}
	}

	return models.EmbeddedContactPoint{}, nil
}

func (r *GrafanaContactPointReconciler) finalize(ctx context.Context, contactPoint *grafanav1beta1.GrafanaContactPoint) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaContactPoint")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, contactPoint)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		cl, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &instance)
		if err != nil {
			return fmt.Errorf("building grafana client: %w", err)
		}

		_, err = cl.Provisioning.DeleteContactpoints(contactPoint.CustomUIDOrUID()) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("deleting contact point: %w", err)
		}

		// Update grafana instance Status
		err = instance.RemoveNamespacedResource(ctx, r.Client, contactPoint)
		if err != nil {
			return fmt.Errorf("removing contact point from Grafana cr: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaContactPointReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	const (
		secretIndexKey    string = ".metadata.secret"
		configMapIndexKey string = ".metadata.configMap"
	)

	// Index the contact points by the Secret references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &grafanav1beta1.GrafanaContactPoint{}, secretIndexKey,
		r.indexSecretSource()); err != nil {
		return fmt.Errorf("failed setting secret index fields: %w", err)
	}

	// Index the contact points by the ConfigMap references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &grafanav1beta1.GrafanaContactPoint{}, configMapIndexKey,
		r.indexConfigMapSource()); err != nil {
		return fmt.Errorf("failed setting configmap index fields: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaContactPoint{}, builder.WithPredicates(
			ignoreStatusUpdates(),
		)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(secretIndexKey)),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(configMapIndexKey)),
		).
		Complete(r)
}

func (r *GrafanaContactPointReconciler) indexSecretSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		contactPoint, ok := o.(*grafanav1beta1.GrafanaContactPoint)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaContactPoint, got %T", o))
		}

		var secretRefs []string

		for _, valueFrom := range contactPoint.Spec.ValuesFrom {
			if valueFrom.ValueFrom.SecretKeyRef != nil {
				secretRefs = append(secretRefs, fmt.Sprintf("%s/%s", contactPoint.Namespace, valueFrom.ValueFrom.SecretKeyRef.Name))
			}
		}

		return secretRefs
	}
}

func (r *GrafanaContactPointReconciler) indexConfigMapSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		contactPoint, ok := o.(*grafanav1beta1.GrafanaContactPoint)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaContactPoint, got %T", o))
		}

		var configMapRefs []string

		for _, valueFrom := range contactPoint.Spec.ValuesFrom {
			if valueFrom.ValueFrom.ConfigMapKeyRef != nil {
				configMapRefs = append(configMapRefs, fmt.Sprintf("%s/%s", contactPoint.Namespace, valueFrom.ValueFrom.ConfigMapKeyRef.Name))
			}
		}

		return configMapRefs
	}
}

func (r *GrafanaContactPointReconciler) requestsForChangeByField(indexKey string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		var list grafanav1beta1.GrafanaContactPointList
		if err := r.List(ctx, &list, client.MatchingFields{
			indexKey: fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
		}); err != nil {
			logf.FromContext(ctx).Error(err, "failed to list contact points for watch mapping")
			return nil
		}

		var reqs []reconcile.Request
		for _, contactPoint := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: contactPoint.Namespace,
				Name:      contactPoint.Name,
			}})
		}

		return reqs
	}
}
