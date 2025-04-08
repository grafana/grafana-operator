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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels/finalizers,verbs=update

func (r *GrafanaLibraryPanelReconciler) syncStatuses(ctx context.Context) error {
	log := logf.FromContext(ctx)

	// get all grafana instances
	grafanas := &v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, grafanas, opts...)
	if err != nil {
		return err
	}
	// no instances, no need to sync
	if len(grafanas.Items) == 0 {
		return nil
	}

	// get all panels
	allPanels := &v1beta1.GrafanaLibraryPanelList{}
	err = r.Client.List(ctx, allPanels, opts...)
	if err != nil {
		return err
	}

	// delete panels from grafana statuses that no longer have a CR
	panelsSynced := 0
	for _, grafana := range grafanas.Items {
		statusUpdated := false
		for _, panel := range grafana.Status.LibraryPanels {
			namespace, name, _ := panel.Split()
			if allPanels.Find(namespace, name) == nil {
				grafana.Status.LibraryPanels = grafana.Status.LibraryPanels.Remove(namespace, name)
				panelsSynced += 1
				statusUpdated = true
			}
		}

		if statusUpdated {
			err = r.Client.Status().Update(ctx, &grafana)
			if err != nil {
				return err
			}
		}
	}

	if panelsSynced > 0 {
		log.Info("successfully synced library panels", "library panels", panelsSynced)
	}
	return nil
}

func (r *GrafanaLibraryPanelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaLibraryPanelReconciler")
	ctx = logf.IntoContext(ctx, log)

	libraryPanel := &v1beta1.GrafanaLibraryPanel{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, libraryPanel)
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

	defer func() {
		libraryPanel.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Client.Status().Update(ctx, libraryPanel); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(libraryPanel.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, libraryPanel); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, libraryPanel); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	// begin validation checks

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
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&libraryPanel.Status.Conditions)
	log.Info("found matching Grafana instances for library panel", "count", len(instances))

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, libraryPanel, contentModel, hash)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Library panel", conditionLibraryPanelSynchronized, libraryPanel.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&libraryPanel.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: libraryPanel.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaLibraryPanelReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaLibraryPanel, model map[string]any, hash string) error {
	if instance.IsInternal() {
		err := ReconcilePlugins(ctx, r.Client, r.Scheme, instance, cr.Spec.Plugins, fmt.Sprintf("%v-librarypanel", cr.Name))
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

	folderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		return err
	}

	uid := content.CustomUIDOrUID(cr, fmt.Sprintf("%s", model["uid"]))
	name := fmt.Sprintf("%s", model["name"])

	defer func() {
		instance.Status.LibraryPanels = instance.Status.LibraryPanels.Add(cr.Namespace, cr.Name, uid)
		//nolint:errcheck
		_ = r.Client.Status().Update(ctx, instance)
	}()

	resp, err := grafanaClient.LibraryElements.GetLibraryElementByUID(uid)

	var panelNotFound *library_elements.GetLibraryElementByUIDNotFound
	if err != nil {
		if !errors.As(err, &panelNotFound) {
			return err
		}

		// doesn't yet exist--should provision
		// nolint:errcheck
		if _, err = grafanaClient.LibraryElements.CreateLibraryElement(&models.CreateLibraryElementCommand{
			FolderUID: folderUID,
			Kind:      int64(libraryElementTypePanel),
			Model:     model,
			Name:      name,
			UID:       uid,
		}); err != nil {
			return err
		}
		return nil
	}

	// handle content caching
	if content.Unchanged(cr, hash) && !cr.ResyncPeriodHasElapsed() {
		return nil
	}

	// nolint:errcheck
	if _, err = grafanaClient.LibraryElements.UpdateLibraryElement(uid, &models.PatchLibraryElementCommand{
		FolderUID: folderUID,
		Kind:      int64(libraryElementTypePanel),
		Model:     model,
		Name:      name,
		Version:   resp.Payload.Result.Version,
	}); err != nil {
		return err
	}

	return nil
}

func (r *GrafanaLibraryPanelReconciler) finalize(ctx context.Context, libraryPanel *v1beta1.GrafanaLibraryPanel) error {
	log := logf.FromContext(ctx)
	log.Info("finalizing GrafanaLibraryPanel")

	uid := content.CustomUIDOrUID(libraryPanel, libraryPanel.Status.UID)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, libraryPanel)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &instance)
		if err != nil {
			return err
		}

		// Not in removeFromInstance to ensure that deleted library panels are removed on sync
		// Avoids repeated synchronization loops if a panel has leftover connections
		switch hasConnections, err := libraryElementHasConnections(grafanaClient, uid); {
		case err != nil:
			return fmt.Errorf("fetching library panel from instance %s/%s: %w", instance.Namespace, instance.Name, err)
		case hasConnections:
			return fmt.Errorf("library panel %s/%s/%s on instance %s/%s has existing connections", libraryPanel.Namespace, libraryPanel.Name, uid, instance.Namespace, instance.Name) //nolint
		}

		_, err = grafanaClient.LibraryElements.DeleteLibraryElementByUID(uid) //nolint:errcheck
		if err != nil {
			var notFound *library_elements.DeleteLibraryElementByUIDNotFound
			if !errors.As(err, &notFound) {
				return err
			}
		}

		instance.Status.LibraryPanels = instance.Status.LibraryPanels.Remove(libraryPanel.Namespace, libraryPanel.Name)
		if err = r.Client.Status().Update(ctx, &instance); err != nil {
			return fmt.Errorf("removing Folder from Grafana cr: %w", err)
		}
	}

	return nil
}

// libraryElementHasConnections returns whether a library panel is still connected to any dashboards
func libraryElementHasConnections(grafanaClient *genapi.GrafanaHTTPAPI, uid string) (bool, error) {
	resp, err := grafanaClient.LibraryElements.GetLibraryElementConnections(uid)
	if err != nil {
		var notFound *library_elements.GetLibraryElementByUIDNotFound
		if errors.Is(err, notFound) {
			// ensure we update the list of managed panels, otherwise we will have dangling references
			return false, nil
		}

		return false, err
	}

	return len(resp.Payload.Result) > 0, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaLibraryPanelReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaLibraryPanel{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
	if err != nil {
		return err
	}

	go func() {
		// periodic sync reconcile
		log := logf.FromContext(ctx).WithName("GrafanaLibraryPanelReconciler")

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(initialSyncDelay):
				start := time.Now()
				err := r.syncStatuses(ctx)
				elapsed := time.Since(start).Milliseconds()
				metrics.InitialLibraryPanelSyncDuration.Set(float64(elapsed))
				if err != nil {
					log.Error(err, "error synchronizing library panels")
					continue
				}

				log.Info("library panel sync complete")
				return
			}
		}
	}()

	return nil
}
