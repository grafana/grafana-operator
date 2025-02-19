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
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type libraryElementType int

type libraryElementOperation int

const (
	conditionLibraryPanelSynchronized = "LibraryPanelSynchronized"

	libraryElementTypePanel    libraryElementType = 1
	libraryElementTypeVariable libraryElementType = 2

	libraryElementOperationNoop   libraryElementOperation = 0
	libraryElementOperationCreate libraryElementOperation = 1
	libraryElementOperationUpdate libraryElementOperation = 2
)

type libraryPanelModelWithHash struct {
	Model map[string]interface{}
	Hash  string
}

type libraryPanelToReconcile struct {
	FolderUID     string
	Kind          int64
	Name          string
	UID           string
	ModelWithHash *libraryPanelModelWithHash
	Version       int64
}

// GrafanaLibraryPanelReconciler reconciles a GrafanaLibraryPanel object
type GrafanaLibraryPanelReconciler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	Discovery discovery.DiscoveryInterface
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanalibrarypanels/finalizers,verbs=update

// libraryElementHasConnections returns whether a library panel is still connected to any dashboards
func libraryElementHasConnections(grafanaClient *genapi.GrafanaHTTPAPI, uid string) (bool, error) {
	resp, err := grafanaClient.LibraryElements.GetLibraryElementConnections(uid)
	if err != nil {
		return false, err
	}
	return len(resp.Payload.Result) > 0, nil
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

	resolver, err := content.NewContentResolver(libraryPanel, r.Client, content.WithDisabledSources([]content.ContentSourceType{
		// grafana.com does not currently support hosting library panels for distribution, but perhaps
		// this will change in the future.
		content.ContentSourceTypeGrafanaCom,
	}))
	if err != nil {
		log.Error(err, "error creating library panel content resolver, this indicates an implementation bug", "libraryPanel", libraryPanel.Name)
		// Failing to create a resolver is an unrecoverable error
		return ctrl.Result{Requeue: false}, nil
	}

	// Retrieving the model before the loop ensures to exit early in case of failure and not fail once per matching instance
	contentModel, hash, err := resolver.Resolve(ctx)
	if err != nil {
		log.Error(err, "error resolving library panel contents", "libraryPanel", libraryPanel.Name)
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, libraryPanel, &libraryPanelModelWithHash{
			Model: contentModel,
			Hash:  hash,
		})
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Library panel", conditionLibraryPanelSynchronized, libraryPanel.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&libraryPanel.Status.Conditions, condition)

	contentUID := fmt.Sprintf("%s", contentModel["uid"])
	libraryPanel.Status.Hash = hash
	libraryPanel.Status.UID = content.CustomUIDOrUID(libraryPanel, contentUID)
	if err := r.Client.Status().Update(ctx, libraryPanel); err != nil {
		log.Error(err, "failed to update status")
	}

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: libraryPanel.Spec.ResyncPeriod.Duration}, nil
}

func (r *GrafanaLibraryPanelReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaLibraryPanel, modelWithHash *libraryPanelModelWithHash) error {
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
	} else if folderUID == "" {
		// this indicates an implementation error; we should be returning errors
		// if the folder isn't found, but an empty UID is still technically possible to return.
		return fmt.Errorf("found folder but could not determine its UID")
	}

	uid := content.CustomUIDOrUID(cr, fmt.Sprintf("%s", modelWithHash.Model["uid"]))
	name := fmt.Sprintf("%s", modelWithHash.Model["name"])

	desired := &libraryPanelToReconcile{
		Name:          name,
		FolderUID:     folderUID,
		Kind:          int64(libraryElementTypePanel),
		UID:           uid,
		ModelWithHash: modelWithHash,
		Version:       0, // this will be filled in by computeOperation in case of update
	}

	op, err := r.computeOperation(grafanaClient, cr, desired)
	if err != nil {
		return err
	}

	switch op {
	case libraryElementOperationNoop: // do nothing
	case libraryElementOperationCreate:
		// nolint:errcheck
		_, err = grafanaClient.LibraryElements.CreateLibraryElement(&models.CreateLibraryElementCommand{
			FolderUID: desired.FolderUID,
			Kind:      desired.Kind,
			Model:     desired.ModelWithHash.Model,
			Name:      desired.Name,
			UID:       desired.UID,
		})
		if err != nil {
			return err
		}
	case libraryElementOperationUpdate:
		// nolint:errcheck
		_, err = grafanaClient.LibraryElements.UpdateLibraryElement(desired.UID, &models.PatchLibraryElementCommand{
			FolderUID: desired.FolderUID,
			Kind:      desired.Kind,
			Model:     desired.ModelWithHash.Model,
			Name:      desired.Name,
			Version:   desired.Version,
		})
		if err != nil {
			return err
		}
	}

	instance.Status.LibraryPanels = instance.Status.LibraryPanels.Add(cr.Namespace, cr.Name, uid)
	return r.Client.Status().Update(ctx, instance)
}

// computeOperation looks at the current state of Grafana versus the CR and determines
// whether a create or update operation should take place (or neither.)
func (r *GrafanaLibraryPanelReconciler) computeOperation(client *genapi.GrafanaHTTPAPI, cr *v1beta1.GrafanaLibraryPanel, desired *libraryPanelToReconcile) (libraryElementOperation, error) {
	resp, err := client.LibraryElements.GetLibraryElementByUID(desired.UID)

	var panelNotFound *library_elements.GetLibraryElementByUIDNotFound
	if err != nil {
		// doesn't yet exist--should provision
		if errors.As(err, &panelNotFound) {
			return libraryElementOperationCreate, nil
		}

		return libraryElementOperationNoop, err
	}

	// it can happen that the user does not utilize `.spec.uid` but updates
	// the UID within the content model itself. this will create a conflict b/c
	// we are effectively requesting a change to the uid, which is immutable.
	if content.IsUpdatedUID(cr, desired.UID) {
		return libraryElementOperationNoop, fmt.Errorf("library panel uid is immutable, but was updated on the content model")
	}

	// handle content caching
	if content.Unchanged(cr, desired.ModelWithHash.Hash) && !cr.ResyncPeriodHasElapsed() {
		return libraryElementOperationNoop, nil
	}

	// mutate(!) to provide the version--this is a bit clunky, but allows us
	// to keep the logic of computing the create vs. update in a separate function
	// while minimizing calls to the Grafana API.
	desired.Version = resp.Payload.Result.Version

	return libraryElementOperationUpdate, nil
}

func (r *GrafanaLibraryPanelReconciler) finalize(ctx context.Context, libraryPanel *v1beta1.GrafanaLibraryPanel) error {
	log := logf.FromContext(ctx)
	log.Info("finalizing GrafanaLibraryPanel")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, libraryPanel)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}
	for _, i := range instances {
		instance := i
		if err := r.removeFromInstance(ctx, &instance, libraryPanel); err != nil {
			return fmt.Errorf("removing notification template from instance: %w", err)
		}
	}

	return nil
}

func (r *GrafanaLibraryPanelReconciler) removeFromInstance(ctx context.Context, instance *v1beta1.Grafana, libraryPanel *v1beta1.GrafanaLibraryPanel) error {
	log := logf.FromContext(ctx)

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, instance)
	if err != nil {
		return err
	}

	updateInstanceRefs := func() error {
		instance.Status.LibraryPanels = instance.Status.LibraryPanels.Remove(libraryPanel.Namespace, libraryPanel.Name)
		return r.Client.Status().Update(ctx, instance)
	}

	uid := libraryPanel.Status.UID

	resp, err := grafanaClient.LibraryElements.GetLibraryElementByUID(uid)
	if err != nil {
		var notFound *library_elements.GetLibraryElementByUIDNotFound
		if errors.As(err, &notFound) {
			// ensure we update the list of managed panels, otherwise we will have dangling references
			return updateInstanceRefs()
		}
		return fmt.Errorf("fetching library panel from instance %s: %w", instance.Status.AdminUrl, err)
	}

	var wrappedRes *models.LibraryElementResponse
	if resp != nil {
		wrappedRes = resp.GetPayload()
	}
	var elem *models.LibraryElementDTO
	if wrappedRes != nil {
		elem = wrappedRes.Result
	}

	switch hasConnections, err := libraryElementHasConnections(grafanaClient, uid); {
	case err != nil:
		return fmt.Errorf("fetching library panel from instance %s: %w", instance.Status.AdminUrl, err)
	case hasConnections:
		return fmt.Errorf("library panel %s on instance %s has existing connections", uid, instance.Status.AdminUrl) //nolint
	}

	_, err = grafanaClient.LibraryElements.DeleteLibraryElementByUID(uid) //nolint:errcheck
	if err != nil {
		var notFound *library_elements.DeleteLibraryElementByUIDNotFound
		if !errors.As(err, &notFound) {
			return err
		}
	}

	if elem != nil && elem.Meta != nil && elem.Meta.FolderUID != "" {
		resp, err := r.DeleteFolderIfEmpty(grafanaClient, elem.Meta.FolderUID)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			log.Info("unused folder successfully removed")
		}
		if resp.StatusCode == 432 {
			log.Info("folder still in use by other resources")
		}
	}

	return updateInstanceRefs()
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaLibraryPanelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaLibraryPanel{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}

func (r *GrafanaLibraryPanelReconciler) DeleteFolderIfEmpty(client *genapi.GrafanaHTTPAPI, folderUID string) (http.Response, error) {
	params := search.NewSearchParams().WithFolderUIDs([]string{folderUID})
	results, err := client.Search.Search(params)
	if err != nil {
		return http.Response{
			Status:     "internal grafana client error getting library panels",
			StatusCode: 500,
		}, err
	}
	if len(results.GetPayload()) > 0 {
		return http.Response{
			Status:     "resource is still in use",
			StatusCode: http.StatusLocked,
		}, err
	}

	deleteParams := folders.NewDeleteFolderParams().WithFolderUID(folderUID)
	if _, err = client.Folders.DeleteFolder(deleteParams); err != nil { //nolint:errcheck
		var notFound *folders.DeleteFolderNotFound
		if !errors.As(err, &notFound) {
			return http.Response{
				Status:     "internal grafana client error deleting grafana folder",
				StatusCode: 500,
			}, err
		}
	}
	return http.Response{
		Status:     "grafana folder deleted",
		StatusCode: 200,
	}, nil
}
