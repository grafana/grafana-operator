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
	"github.com/go-logr/logr"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
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

	folder := &v1beta1.GrafanaFolder{}

	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, folder)
	if err != nil {
		if errors.IsNotFound(err) {
			//TODO on folder deleted
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana folder cr")
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	if folder.Spec.InstanceSelector == nil {
		controllerLog.Info("no instance selector found for dashboard, nothing to do", "name", folder.Name, "namespace", folder.Namespace)
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
		if grafana.Status.AdminUrl == "" || grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
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
func (r *GrafanaFolderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaFolder{}).
		Complete(r)
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

	folderuid := "1"         //TODO replace me
	title := "my-new-folder" //TODO replace me

	folderFromClient, err := grafanaClient.NewFolder(title, folderuid)
	if err != nil {
		//TODO Check error conditions
		return err
	}

	//FIXME our current version of the client doesn't return response codes, or any response for
	//FIXME that matter, this needs an issue/feature request upstream
	//FIXME for now, use the returned URL as an indicator that the folder was created instead
	if folderFromClient.URL == "" && len(folderFromClient.URL) <= 0 {
		return errors.NewBadRequest(fmt.Sprintf("something went wrong trying to create folder %s in grafana %s", cr.Name, grafana.Name))
	}

	return nil

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
