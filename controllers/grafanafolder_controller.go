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
	"errors"
	"fmt"
	"strings"
	"time"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/models"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionFolderSynchronized = "FolderSynchronized"
)

// GrafanaFolderReconciler reconciles a GrafanaFolder object
type GrafanaFolderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders/finalizers,verbs=update

func (r *GrafanaFolderReconciler) syncFolders(ctx context.Context) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	foldersSynced := 0

	// get all grafana instances
	grafanas := &grafanav1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, grafanas, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// no instances, no need to sync
	if len(grafanas.Items) == 0 {
		return ctrl.Result{Requeue: false}, nil
	}

	// get all folders
	allFolders := &grafanav1beta1.GrafanaFolderList{}
	err = r.Client.List(ctx, allFolders, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// sync folders, delete folders from grafana that do no longer have a cr
	foldersToDelete := map[*grafanav1beta1.Grafana][]grafanav1beta1.NamespacedResource{}
	for _, grafana := range grafanas.Items {
		grafana := grafana
		for _, folder := range grafana.Status.Folders {
			if allFolders.Find(folder.Namespace(), folder.Name()) == nil {
				foldersToDelete[&grafana] = append(foldersToDelete[&grafana], folder)
			}
		}
	}

	// delete all folders that no longer have a cr
	for grafana, existingFolders := range foldersToDelete {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{}, err
		}

		for _, folder := range existingFolders {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to a certain number of folders per cycle. This means that it will take longer to sync
			// a large number of deleted dashboard crs, but that should be an edge case.
			if foldersSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := folder.Split()

			reftrue := true
			params := folders.NewDeleteFolderParams().WithFolderUID(uid).WithForceDeleteRules(&reftrue)
			_, err = grafanaClient.Folders.DeleteFolder(params) //nolint
			if err != nil {
				var notFound *folders.DeleteFolderNotFound
				if errors.As(err, &notFound) {
					log.Info("folder no longer exists", "namespace", namespace, "name", name)
				} else {
					return ctrl.Result{}, err
				}
			}

			grafana.Status.Folders = grafana.Status.Folders.Remove(namespace, name)
			foldersSynced += 1
		}

		// one update per grafana - this will trigger a reconcile of the grafana controller
		// so we should minimize those updates
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if foldersSynced > 0 {
		log.Info("successfully synced folders", "folders", foldersSynced)
	}
	return ctrl.Result{Requeue: false}, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GrafanaFolder object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *GrafanaFolderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaFolderReconciler")
	logf.IntoContext(ctx, log)

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncFolders(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialFoldersSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	folder := &grafanav1beta1.GrafanaFolder{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, folder)
	if err != nil {
		if kuberr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("error getting grafana folder cr: %w", err)
	}

	if folder.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(folder, grafanaFinalizer) {
			if err := r.finalize(ctx, folder); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaFolder: %w", err)
			}
			if err := removeFinalizer(ctx, r.Client, folder); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	defer func() {
		folder.Status.Hash = folder.Hash()
		folder.Status.LastResync = metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, folder); err != nil {
			log.Error(err, "updating status")
		}
		if meta.IsStatusConditionTrue(folder.Status.Conditions, conditionNoMatchingInstance) {
			if err := removeFinalizer(ctx, r.Client, folder); err != nil {
				log.Error(err, "failed to remove finalizer")
			}
		} else {
			if err := addFinalizer(ctx, r.Client, folder); err != nil {
				log.Error(err, "failed to set finalizer")
			}
		}
	}()

	if folder.Spec.ParentFolderUID == folder.CustomUIDOrUID() {
		setInvalidSpec(&folder.Status.Conditions, folder.Generation, "CyclicParent", "The value of parentFolderUID must not be the uid of the current folder")
		meta.RemoveStatusCondition(&folder.Status.Conditions, conditionFolderSynchronized)
		return ctrl.Result{}, fmt.Errorf("cyclic folder reference")
	}
	removeInvalidSpec(&folder.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, folder)
	if err != nil {
		setNoMatchingInstancesCondition(&folder.Status.Conditions, folder.Generation, err)
		meta.RemoveStatusCondition(&folder.Status.Conditions, conditionFolderSynchronized)
		folder.Status.NoMatchingInstances = true
		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&folder.Status.Conditions, folder.Generation, err)
		meta.RemoveStatusCondition(&folder.Status.Conditions, conditionFolderSynchronized)
		folder.Status.NoMatchingInstances = true
		return ctrl.Result{RequeueAfter: RequeueDelay}, nil
	}

	removeNoMatchingInstance(&folder.Status.Conditions)
	folder.Status.NoMatchingInstances = false
	log.Info("found matching Grafana instances for folder", "count", len(instances))

	applyErrors := make(map[string]string)
	for _, grafana := range instances {
		grafana := grafana

		err = r.onFolderCreated(ctx, &grafana, folder)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	condition := buildSynchronizedCondition("Folder", conditionFolderSynchronized, folder.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&folder.Status.Conditions, condition)

	return ctrl.Result{RequeueAfter: folder.Spec.ResyncPeriod.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaFolderReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaFolder{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)

	if err == nil {
		go func() {
			log := logf.FromContext(ctx).WithName("GrafanaFolderReconciler")
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(initialSyncDelay):
					result, err := r.Reconcile(ctx, ctrl.Request{})
					if err != nil {
						log.Error(err, "error synchronizing folders")
						continue
					}
					if result.Requeue {
						log.Info("more folders left to synchronize")
						continue
					}
					log.Info("folder sync complete")
					return
				}
			}
		}()
	}

	return err
}

func (r *GrafanaFolderReconciler) finalize(ctx context.Context, folder *grafanav1beta1.GrafanaFolder) error {
	instances, err := GetScopedMatchingInstances(ctx, r.Client, folder)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	reftrue := true
	params := folders.NewDeleteFolderParams().WithForceDeleteRules(&reftrue)

	for _, grafana := range instances {
		grafana := grafana
		if found, uid := grafana.Status.Folders.Find(folder.Namespace, folder.Name); found {
			grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			_, err = grafanaClient.Folders.DeleteFolder(params.WithFolderUID(*uid)) //nolint
			if err != nil {
				var notFound *folders.DeleteFolderNotFound
				if !errors.As(err, &notFound) {
					return err
				}
			}

			grafana.Status.Folders = grafana.Status.Folders.Remove(folder.Namespace, folder.Name)
			if err = r.Client.Status().Update(ctx, &grafana); err != nil {
				return fmt.Errorf("removing Folder from Grafana cr: %w", err)
			}
		}
	}

	return nil
}

func (r *GrafanaFolderReconciler) onFolderCreated(ctx context.Context, grafana *grafanav1beta1.Grafana, cr *grafanav1beta1.GrafanaFolder) error {
	title := cr.GetTitle()
	uid := cr.CustomUIDOrUID()

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	parentFolderUID, err := getFolderUID(ctx, r.Client, cr)
	if err != nil {
		return err
	}

	exists, remoteUID, remoteParent, err := r.Exists(grafanaClient, cr)
	if err != nil {
		return err
	}

	// always update after resync period has elapsed even if cr is unchanged.
	if exists && cr.Unchanged() && !cr.ResyncPeriodHasElapsed() && parentFolderUID == remoteParent {
		return nil
	}

	if exists {
		// make sure we use the correct UID
		uid = remoteUID
		// Add to status to cover cases:
		// - operator have previously failed to update status
		// - the folder was created outside of operator
		// - the folder was created through dashboard controller
		if found, _ := grafana.Status.Folders.Find(cr.Namespace, cr.Name); !found {
			grafana.Status.Folders = grafana.Status.Folders.Add(cr.Namespace, cr.Name, uid)
			err = r.Client.Status().Update(ctx, grafana)
			if err != nil {
				return err
			}
		}

		if !cr.Unchanged() {
			_, err = grafanaClient.Folders.UpdateFolder(remoteUID, &models.UpdateFolderCommand{ //nolint
				Overwrite: true,
				Title:     title,
			})
			if err != nil {
				return err
			}
		}

		if parentFolderUID != remoteParent {
			_, err = grafanaClient.Folders.MoveFolder(remoteUID, &models.MoveFolderCommand{ //nolint
				ParentUID: parentFolderUID,
			})
			if err != nil {
				return err
			}
		}
	} else {
		body := &models.CreateFolderCommand{
			Title:     title,
			UID:       uid,
			ParentUID: parentFolderUID,
		}

		folderResp, err := grafanaClient.Folders.CreateFolder(body)
		if err != nil {
			return err
		}

		grafana.Status.Folders = grafana.Status.Folders.Add(cr.Namespace, cr.Name, folderResp.Payload.UID)
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return err
		}
	}

	// NOTE: it's up to a user to reset permissions with correct json
	if cr.Spec.Permissions != "" {
		permissions := models.UpdateDashboardACLCommand{}
		err = json.Unmarshal([]byte(cr.Spec.Permissions), &permissions)
		if err != nil {
			return fmt.Errorf("failed to unmarshal spec.permissions: %w", err)
		}

		_, err = grafanaClient.FolderPermissions.UpdateFolderPermissions(uid, &permissions) //nolint
		if err != nil {
			return fmt.Errorf("failed to update folder permissions: %w", err)
		}
	}

	return nil
}

// Check if the folder exists. Matches UID first and fall back to title. Title matching only works for non-nested folders
func (r *GrafanaFolderReconciler) Exists(client *genapi.GrafanaHTTPAPI, cr *grafanav1beta1.GrafanaFolder) (bool, string, string, error) {
	title := cr.GetTitle()
	uid := cr.CustomUIDOrUID()

	uidResp, err := client.Folders.GetFolderByUID(uid)
	if err == nil {
		return true, uidResp.Payload.UID, uidResp.Payload.ParentUID, nil
	}

	page := int64(1)
	limit := int64(10000)
	for {
		params := folders.NewGetFoldersParams().WithPage(&page).WithLimit(&limit)

		foldersResp, err := client.Folders.GetFolders(params)
		if err != nil {
			return false, "", "", err
		}
		folders := foldersResp.GetPayload()

		for _, remoteFolder := range folders {
			if strings.EqualFold(remoteFolder.Title, title) {
				return true, remoteFolder.UID, remoteFolder.ParentUID, nil
			}
		}
		if len(folders) < int(limit) {
			return false, "", "", nil
		}
		page++
	}
}
