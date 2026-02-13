/*
Copyright 2025.

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

	"github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type libraryElementType int

const (
	conditionLibraryPanelSynchronized                    = "LibraryPanelSynchronized"
	libraryElementTypePanel           libraryElementType = 1

	LogMsgInvalidPanelSpec       = "invalid Library Panel spec"
	LogMsgResolvingPanelContents = "error resolving library panel contents"
)

var errLibraryPanelContentUIDImmutable = errors.New("library panel uid is immutable, but was updated on the content model")

// GrafanaLibraryPanelReconciler reconciles a GrafanaLibraryPanel object
type GrafanaLibraryPanelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaLibraryPanelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaLibraryPanelReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaLibraryPanel{}

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

	resolver := content.NewResolver(cr, r.Client, content.WithDisabledSources([]content.SourceType{
		// grafana.com does not currently support hosting library panels for distribution, but perhaps
		// this will change in the future.
		content.SourceTypeGrafanaCom,
	}))

	// Retrieving the model before the loop ensures to exit early in case of failure and not fail once per matching instance
	contentModel, hash, err := resolver.Resolve(ctx)
	if err != nil {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidModelResolution", err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionLibraryPanelSynchronized)
		log.Error(err, LogMsgResolvingPanelContents)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgResolvingPanelContents, err)
	}

	contentUID := fmt.Sprintf("%s", contentModel["uid"])
	// it can happen that the user does not utilize `.spec.uid` but updates
	// the UID within the content model itself. this will create a conflict b/c
	// we are effectively requesting a change to the uid, which is immutable.
	if content.IsUpdatedUID(cr, contentUID) {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, "InvalidModel", errLibraryPanelContentUIDImmutable.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionLibraryPanelSynchronized)
		log.Error(errLibraryPanelContentUIDImmutable, LogMsgInvalidPanelSpec)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgInvalidPanelSpec, errLibraryPanelContentUIDImmutable)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	// begin instance selection and reconciliation

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionLibraryPanelSynchronized)
		log.Error(err, LogMsgGettingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionLibraryPanelSynchronized)
		log.Error(ErrNoMatchingInstances, LogMsgNoMatchingInstances)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgNoMatchingInstances, ErrNoMatchingInstances)
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	log.V(1).Info(DbgMsgFoundMatchingInstances, "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgResolvingFolderUID)
		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgResolvingFolderUID, err)
	}

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr, contentModel, hash, folderUID)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Library panel", conditionLibraryPanelSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		err = fmt.Errorf(FmtStrApplyErrors, applyErrors)
		log.Error(err, LogMsgApplyErrors)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgApplyErrors, err)
	}

	cr.Status.Hash = hash
	cr.Status.UID = content.GetGrafanaUID(cr, contentUID)

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaLibraryPanelReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaLibraryPanel, model map[string]any, hash, folderUID string) error {
	if instance.IsInternal() {
		err := ReconcilePlugins(ctx, r.Client, r.Scheme, instance, cr.Spec.Plugins, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
		if err != nil {
			return err
		}
	} else if instance.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your library panel cr")
	}

	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return err
	}

	uid := content.GetGrafanaUID(cr, fmt.Sprintf("%s", model["uid"]))
	name := fmt.Sprintf("%s", model["name"])

	resp, err := gClient.LibraryElements.GetLibraryElementByUID(uid)
	if err != nil {
		if IsNotErrorType[*library_elements.GetLibraryElementByUIDNotFound](err) {
			return err
		}

		// doesn't yet exist--should provision
		//nolint:errcheck
		_, err = gClient.LibraryElements.CreateLibraryElement(&models.CreateLibraryElementCommand{
			FolderUID: folderUID,
			Kind:      int64(libraryElementTypePanel),
			Model:     model,
			Name:      name,
			UID:       uid,
		})
		if err != nil {
			return err
		}

		return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource(uid))
	}

	// handle content caching
	if content.HasChanged(cr, hash) {
		_, err = gClient.LibraryElements.UpdateLibraryElement(uid, &models.PatchLibraryElementCommand{ //nolint:errcheck
			FolderUID: folderUID,
			Kind:      int64(libraryElementTypePanel),
			Model:     model,
			Name:      name,
			Version:   resp.Payload.Result.Version,
		})
		if err != nil {
			return err
		}
	}

	// Update grafana instance Status
	return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource(uid))
}

func (r *GrafanaLibraryPanelReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaLibraryPanel) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaLibraryPanel")

	uid := content.GetGrafanaUID(cr, cr.Status.UID)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		log.Error(err, LogMsgGettingInstances)
		return fmt.Errorf("%s: %w", LogMsgGettingInstances, err)
	}

	for _, grafana := range instances {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		isCleanupInGrafanaRequired := true

		resp, err := gClient.LibraryElements.GetLibraryElementByUID(uid)
		if err != nil {
			if IsNotErrorType[*library_elements.GetLibraryElementByUIDNotFound](err) {
				return fmt.Errorf("fetching library panel from instance %s/%s: %w", grafana.Namespace, grafana.Name, err)
			}

			isCleanupInGrafanaRequired = false
		}

		// Skip cleanup in instances
		if isCleanupInGrafanaRequired {
			if resp.Payload.Result.Meta.ConnectedDashboards > 0 {
				return fmt.Errorf("library panel %s/%s/%s on instance %s/%s has existing connections", cr.Namespace, cr.Name, uid, grafana.Namespace, grafana.Name)
			}

			_, err = gClient.LibraryElements.DeleteLibraryElementByUID(uid) //nolint:errcheck
			if err != nil {
				if IsNotErrorType[*library_elements.DeleteLibraryElementByUIDNotFound](err) {
					return err
				}
			}
		}

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
			if err != nil {
				return fmt.Errorf("reconciling plugins: %w", err)
			}
		}

		// Update grafana instance Status
		err = grafana.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return fmt.Errorf("removing Folder from Grafana cr: %w", err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaLibraryPanelReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	const (
		configMapIndexKey string = ".metadata.configMap"
	)

	// Index the library panels by the ConfigMap references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &v1beta1.GrafanaLibraryPanel{}, configMapIndexKey,
		r.indexConfigMapSource()); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaLibraryPanel{}, builder.WithPredicates(
			ignoreStatusUpdates(),
		)).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(configMapIndexKey)),
		).
		Complete(r)
}

func (r *GrafanaLibraryPanelReconciler) indexConfigMapSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		cr, ok := o.(*v1beta1.GrafanaLibraryPanel)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaLibraryPanel, got %T", o))
		}

		if cr.Spec.ConfigMapRef != nil {
			return []string{fmt.Sprintf("%s/%s", cr.Namespace, cr.Spec.ConfigMapRef.Name)}
		}

		return nil
	}
}

func (r *GrafanaLibraryPanelReconciler) requestsForChangeByField(indexKey string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		var list v1beta1.GrafanaLibraryPanelList
		if err := r.List(ctx, &list, client.MatchingFields{
			indexKey: fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
		}); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for _, cr := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: cr.Namespace,
				Name:      cr.Name,
			}})
		}

		return reqs
	}
}
