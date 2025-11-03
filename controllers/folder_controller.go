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

	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/models"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionFolderSynchronized = "FolderSynchronized"
	conditionReasonCyclicParent = "CyclicParent"
)

// GrafanaFolderReconciler reconciles a GrafanaFolder object
type GrafanaFolderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaFolderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaFolderReconciler")
	ctx = logf.IntoContext(ctx, log)

	folder := &grafanav1beta1.GrafanaFolder{}

	err := r.Get(ctx, req.NamespacedName, folder)
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

	defer UpdateStatus(ctx, r.Client, folder)

	if folder.Spec.Suspend {
		setSuspended(&folder.Status.Conditions, folder.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&folder.Status.Conditions)

	if folder.Spec.ParentFolderUID == folder.CustomUIDOrUID() {
		setInvalidSpec(&folder.Status.Conditions, folder.Generation, conditionReasonCyclicParent, "The value of parentFolderUID must not be the uid of the current folder")
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

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&folder.Status.Conditions)
	folder.Status.NoMatchingInstances = false

	parentFolderUID, err := getFolderUID(ctx, r.Client, folder)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf(ErrFetchingFolder, err)
	}

	log.Info("found matching Grafana instances for folder", "count", len(instances))

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err = r.onFolderCreated(ctx, &grafana, folder, parentFolderUID)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Folder", conditionFolderSynchronized, folder.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&folder.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	folder.Status.Hash = folder.Hash()

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(folder.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaFolderReconciler) finalize(ctx context.Context, folder *grafanav1beta1.GrafanaFolder) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaFolder")

	uid := folder.CustomUIDOrUID()

	instances, err := GetScopedMatchingInstances(ctx, r.Client, folder)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	reftrue := true
	params := folders.NewDeleteFolderParams().WithForceDeleteRules(&reftrue)

	for _, grafana := range instances {
		grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		_, err = grafanaClient.Folders.DeleteFolder(params.WithFolderUID(uid)) //nolint
		if err != nil {
			var notFound *folders.DeleteFolderNotFound
			if !errors.As(err, &notFound) {
				return err
			}
		}

		// Update grafana instance Status
		err = grafana.RemoveNamespacedResource(ctx, r.Client, folder)
		if err != nil {
			return fmt.Errorf("removing Folder from Grafana cr: %w", err)
		}
	}

	return nil
}

func (r *GrafanaFolderReconciler) onFolderCreated(ctx context.Context, grafana *grafanav1beta1.Grafana, cr *grafanav1beta1.GrafanaFolder, parentFolderUID string) error {
	log := logf.FromContext(ctx)

	title := cr.GetTitle()
	uid := cr.CustomUIDOrUID()

	grafanaClient, err := client2.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	exists, remoteUID, remoteParent, err := r.Exists(grafanaClient, cr)
	if err != nil {
		return err
	}

	// Update when missing, the CR is updated or parentFolder has changed.
	if exists && cr.Unchanged() && parentFolderUID == remoteParent {
		log.V(1).Info("folder unchanged. skipping remaining requests")
		return nil
	}

	if exists {
		// make sure we use the correct UID
		uid = remoteUID

		if !cr.Unchanged() {
			_, err = grafanaClient.Folders.UpdateFolder(remoteUID, &models.UpdateFolderCommand{ //nolint:errcheck
				Overwrite: true,
				Title:     title,
			})
			if err != nil {
				return err
			}
		}

		if parentFolderUID != remoteParent {
			_, err = grafanaClient.Folders.MoveFolder(remoteUID, &models.MoveFolderCommand{ //nolint:errcheck
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

		_, err := grafanaClient.Folders.CreateFolder(body) // nolint:errcheck
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

		_, err = grafanaClient.FolderPermissions.UpdateFolderPermissions(uid, &permissions) //nolint:errcheck
		if err != nil {
			return fmt.Errorf("failed to update folder permissions: %w", err)
		}
	}

	// Update grafana instance Status
	return grafana.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource(uid))
}

// Check if the folder exists. Matches UID first and fall back to title. Title matching only works for non-nested folders
func (r *GrafanaFolderReconciler) Exists(client *genapi.GrafanaHTTPAPI, cr *grafanav1beta1.GrafanaFolder) (bool, string, string, error) {
	title := cr.GetTitle()
	uid := cr.CustomUIDOrUID()

	uidResp, err := client.Folders.GetFolderByUID(uid)
	if err == nil {
		return true, uidResp.Payload.UID, uidResp.Payload.ParentUID, nil
	}

	// If we could not find the UID in the Grafana but a CustomUID is set in the CR we must assume the folder does not exist.
	if cr.Spec.CustomUID != "" {
		return false, uid, "", nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaFolderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaFolder{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
