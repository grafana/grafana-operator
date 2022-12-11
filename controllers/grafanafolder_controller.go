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
	"time"

	"github.com/go-logr/logr"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/metrics"
	grapi "github.com/grafana/grafana-api-golang-client"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
)

// GrafanaFolderReconciler reconciles a GrafanaFolder object
type GrafanaFolderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanafolders/finalizers,verbs=update

func (r *GrafanaFolderReconciler) syncFolders(ctx context.Context) (ctrl.Result, error) {
	syncLog := log.FromContext(ctx)
	foldersSynced := 0

	// get all grafana instances
	grafanas := &v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, grafanas, opts...)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	// no instances, no need to sync
	if len(grafanas.Items) == 0 {
		return ctrl.Result{Requeue: false}, nil
	}

	// get all folders
	allFolders := &v1beta1.GrafanaFolderList{}
	err = r.Client.List(ctx, allFolders, opts...)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	// sync folders, delete folders from grafana that do no longer have a cr
	foldersToDelete := map[*v1beta1.Grafana][]v1beta1.NamespacedResource{}
	for _, grafana := range grafanas.Items {
		for _, folder := range grafana.Status.Folders {
			if allFolders.Find(folder.Namespace(), folder.Name()) == nil {
				foldersToDelete[&grafana] = append(foldersToDelete[&grafana], folder)
			}
		}
	}

	// delete all dashboards that no longer have a cr
	for grafana, folders := range foldersToDelete {
		grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		for _, folder := range folders {
			// avoid bombarding the grafana instance with a large number of requests at once, limit
			// the sync to a certain number of folders per cycle. This means that it will take longer to sync
			// a large number of deleted dashboard crs, but that should be an edge case.
			if foldersSynced >= syncBatchSize {
				return ctrl.Result{Requeue: true}, nil
			}

			namespace, name, uid := folder.Split()
			err = grafanaClient.DeleteDashboardByUID(uid)
			if err != nil {
				if strings.Contains(err.Error(), "status: 404") {
					syncLog.Info("folder no longer exists", "namespace", namespace, "name", name)
				} else {
					return ctrl.Result{Requeue: false}, err
				}
			}

			grafana.Status.Folders = grafana.Status.Folders.Remove(namespace, name)
			foldersSynced += 1
		}

		// one update per grafana - this will trigger a reconcile of the grafana controller
		// so we should minimize those updates
		err = r.Client.Status().Update(ctx, grafana)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}
	}

	if foldersSynced > 0 {
		syncLog.Info("successfully synced folders", "folders", foldersSynced)
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
	controllerLog := log.FromContext(ctx)
	r.Log = controllerLog

	// periodic sync reconcile
	if req.Namespace == "" && req.Name == "" {
		start := time.Now()
		syncResult, err := r.syncFolders(ctx)
		elapsed := time.Since(start).Milliseconds()
		metrics.InitialFoldersSyncDuration.Set(float64(elapsed))
		return syncResult, err
	}

	folder := &v1beta1.GrafanaFolder{}

	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, folder)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.onFolderDeleted(ctx, req.Namespace, req.Name); err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelayError}, err
			}
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana folder cr")
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	if folder.Spec.InstanceSelector == nil {
		controllerLog.Info("no instance selector found for folder, nothing to do", "name", folder.Name, "namespace", folder.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, nil
	}

	instances, err := GetMatchingInstances(ctx, r.Client, folder.Spec.InstanceSelector)
	if err != nil {
		controllerLog.Error(err, "could not find matching instances", "name", folder.Name)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}
	// your logic here

	if len(instances.Items) == 0 {
		controllerLog.Info("no matching instances found for folder")
		return ctrl.Result{Requeue: false}, nil
	}

	controllerLog.Info("found matching Grafana instances for folder", "count", len(instances.Items))

	for _, grafana := range instances.Items {
		//if grafana.Status.AdminUrl == "" || grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
		if grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		err = r.onFolderCreated(ctx, &grafana, folder)
		if err != nil {
			controllerLog.Error(err, "error reconciling folder", "folder", folder.Name, "grafana", grafana.Name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaFolderReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaFolder{}).
		Complete(r)

	if err == nil {
		d, err := time.ParseDuration(initialSyncDelay)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d):
					result, err := r.Reconcile(ctx, ctrl.Request{})
					if err != nil {
						r.Log.Error(err, "error synchronizing folders")
						continue
					}
					if result.Requeue {
						r.Log.Info("more folders left to synchronize")
						continue
					}
					r.Log.Info("folder sync complete")
					return
				}
			}
		}()
	}

	return err
}

func (r *GrafanaFolderReconciler) onFolderDeleted(ctx context.Context, namespace string, name string) error {
	list := v1beta1.GrafanaList{}
	var opts []client.ListOption
	err := r.Client.List(ctx, &list, opts...)
	if err != nil {
		return err
	}

	for _, grafana := range list.Items {
		if found, uid := grafana.Status.Folders.Find(namespace, name); found {
			grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			err = grafanaClient.DeleteFolder(*uid)
			if err != nil {
				if !strings.Contains(err.Error(), "status: 404") {
					return err
				}
			}

			grafana.Status.Folders = grafana.Status.Folders.Remove(namespace, name)
			err = r.Client.Status().Update(ctx, &grafana)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *GrafanaFolderReconciler) onFolderCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaFolder) error {
	if cr.Spec.Json == "" {
		return nil
	}

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	exists, err := r.Exists(grafanaClient, cr)
	if err != nil {
		return err
	}
	if exists && cr.Unchanged() {
		return nil
	}

	var folderFromJson map[string]interface{}
	err = json.Unmarshal([]byte(cr.Spec.Json), &folderFromJson)
	if err != nil {
		return err
	}

	title := fmt.Sprintf("%v", folderFromJson["title"])
	if title == "" {
		title = cr.Name
	}

	// folder exists, update only
	if exists && !cr.Unchanged() {
		err = grafanaClient.UpdateFolder(string(cr.UID), title)
		if err != nil {
			return err
		}

		return r.UpdateStatus(ctx, cr)
	}

	folderFromClient, err := grafanaClient.NewFolder(title, string(cr.UID))
	if err != nil {
		// folder already exists in grafana, do nothing
		if strings.Contains(err.Error(), "status: 409") {
			return nil
		}
		return err
	}

	//FIXME our current version of the client doesn't return response codes, or any response for
	//FIXME that matter, this needs an issue/feature request upstream
	//FIXME for now, use the returned URL as an indicator that the folder was created instead
	if folderFromClient.URL == "" && len(folderFromClient.URL) <= 0 {
		return errors.NewBadRequest(fmt.Sprintf("something went wrong trying to create folder %s in grafana %s", cr.Name, grafana.Name))
	}

	grafana.Status.Folders = grafana.Status.Folders.Add(cr.Namespace, cr.Name, folderFromClient.UID)
	err = r.Client.Status().Update(ctx, grafana)
	if err != nil {
		return err
	}

	return r.UpdateStatus(ctx, cr)
}

func (r *GrafanaFolderReconciler) UpdateStatus(ctx context.Context, cr *v1beta1.GrafanaFolder) error {
	cr.Status.Hash = cr.Hash()
	return r.Client.Status().Update(ctx, cr)
}

func (r *GrafanaFolderReconciler) Exists(client *grapi.Client, cr *v1beta1.GrafanaFolder) (bool, error) {
	folders, err := client.Folders()
	if err != nil {
		return false, err
	}
	for _, folder := range folders {
		if folder.UID == string(cr.UID) {
			return true, nil
		}
	}
	return false, nil
}
