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
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
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
	conditionLibraryPanelSynchronized = "LibraryPanelSynchronized"

	libraryElementTypePanel    libraryElementType = 1
	libraryElementTypeVariable libraryElementType = 2
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

	libraryPanel := &v1beta1.GrafanaLibraryPanel{}

	err := r.Get(ctx, req.NamespacedName, libraryPanel)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get GrafanaLibraryPanel: %w", err)
	}

	if libraryPanel.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(libraryPanel, grafanaFinalizer) {
			if err := r.finalize(ctx, libraryPanel); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaLibraryPanel: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, libraryPanel); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, libraryPanel)

	if libraryPanel.Spec.Suspend {
		setSuspended(&libraryPanel.Status.Conditions, libraryPanel.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&libraryPanel.Status.Conditions)

	resolver := content.NewContentResolver(libraryPanel, r.Client, content.WithDisabledSources([]content.ContentSourceType{
		// grafana.com does not currently support hosting library panels for distribution, but perhaps
		// this will change in the future.
		content.ContentSourceTypeGrafanaCom,
	}))

	// Retrieving the model before the loop ensures to exit early in case of failure and not fail once per matching instance
	contentModel, hash, err := resolver.Resolve(ctx)
	if err != nil {
		setInvalidSpec(&libraryPanel.Status.Conditions, libraryPanel.Generation, "InvalidModelResolution", err.Error())
		meta.RemoveStatusCondition(&libraryPanel.Status.Conditions, conditionLibraryPanelSynchronized)

		return ctrl.Result{}, fmt.Errorf("error resolving library panel contents: %w", err)
	}

	contentUID := fmt.Sprintf("%s", contentModel["uid"])
	// it can happen that the user does not utilize `.spec.uid` but updates
	// the UID within the content model itself. this will create a conflict b/c
	// we are effectively requesting a change to the uid, which is immutable.
	if content.IsUpdatedUID(libraryPanel, contentUID) {
		setInvalidSpec(&libraryPanel.Status.Conditions, libraryPanel.Generation, "InvalidModel", errLibraryPanelContentUIDImmutable.Error())
		meta.RemoveStatusCondition(&libraryPanel.Status.Conditions, conditionLibraryPanelSynchronized)

		return ctrl.Result{}, errLibraryPanelContentUIDImmutable
	}

	removeInvalidSpec(&libraryPanel.Status.Conditions)
	libraryPanel.Status.Hash = hash
	libraryPanel.Status.UID = content.CustomUIDOrUID(libraryPanel, contentUID)

	// begin instance selection and reconciliation

	instances, err := GetScopedMatchingInstances(ctx, r.Client, libraryPanel)
	if err != nil {
		setNoMatchingInstancesCondition(&libraryPanel.Status.Conditions, libraryPanel.Generation, err)
		meta.RemoveStatusCondition(&libraryPanel.Status.Conditions, conditionLibraryPanelSynchronized)

		return ctrl.Result{}, fmt.Errorf("could not find matching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&libraryPanel.Status.Conditions, libraryPanel.Generation, err)
		meta.RemoveStatusCondition(&libraryPanel.Status.Conditions, conditionLibraryPanelSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&libraryPanel.Status.Conditions)
	log.Info("found matching Grafana instances for library panel", "count", len(instances))

	folderUID, err := getFolderUID(ctx, r.Client, libraryPanel)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf(ErrFetchingFolder, err)
	}

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, libraryPanel, contentModel, hash, folderUID)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Library panel", conditionLibraryPanelSynchronized, libraryPanel.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&libraryPanel.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(libraryPanel.Spec.ResyncPeriod)}, nil
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

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return err
	}

	uid := content.CustomUIDOrUID(cr, fmt.Sprintf("%s", model["uid"]))
	name := fmt.Sprintf("%s", model["name"])

	resp, err := grafanaClient.LibraryElements.GetLibraryElementByUID(uid)

	var panelNotFound *library_elements.GetLibraryElementByUIDNotFound
	if err != nil {
		if !errors.As(err, &panelNotFound) {
			return err
		}

		// doesn't yet exist--should provision
		// nolint:errcheck
		_, err = grafanaClient.LibraryElements.CreateLibraryElement(&models.CreateLibraryElementCommand{
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
		_, err = grafanaClient.LibraryElements.UpdateLibraryElement(uid, &models.PatchLibraryElementCommand{ // nolint:errcheck
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

	uid := content.CustomUIDOrUID(cr, cr.Status.UID)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		isCleanupInGrafanaRequired := true

		resp, err := grafanaClient.LibraryElements.GetLibraryElementByUID(uid)
		if err != nil {
			var notFound *library_elements.GetLibraryElementByUIDNotFound
			if !errors.As(err, &notFound) {
				return fmt.Errorf("fetching library panel from instance %s/%s: %w", grafana.Namespace, grafana.Name, err)
			}

			isCleanupInGrafanaRequired = false
		}

		// Skip cleanup in instances
		if isCleanupInGrafanaRequired {
			if resp.Payload.Result.Meta.ConnectedDashboards > 0 {
				return fmt.Errorf("library panel %s/%s/%s on instance %s/%s has existing connections", cr.Namespace, cr.Name, uid, grafana.Namespace, grafana.Name) //nolint
			}

			_, err = grafanaClient.LibraryElements.DeleteLibraryElementByUID(uid) //nolint:errcheck
			if err != nil {
				var notFound *library_elements.DeleteLibraryElementByUIDNotFound
				if !errors.As(err, &notFound) {
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
		libraryPanel, ok := o.(*v1beta1.GrafanaLibraryPanel)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaLibraryPanel, got %T", o))
		}

		if libraryPanel.Spec.ConfigMapRef != nil {
			return []string{fmt.Sprintf("%s/%s", libraryPanel.Namespace, libraryPanel.Spec.ConfigMapRef.Name)}
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
		for _, libraryPanel := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: libraryPanel.Namespace,
				Name:      libraryPanel.Name,
			}})
		}

		return reqs
	}
}
