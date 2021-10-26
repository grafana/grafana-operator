/*
Copyright 2021.

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

package grafanadatasource

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sort"

	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/common"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	integreatlyorgv1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
)

// GrafanaDatasourceReconciler reconciles a GrafanaDatasource object
type GrafanaDatasourceReconciler struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	Client   client.Client
	Scheme   *runtime.Scheme
	Context  context.Context
	Cancel   context.CancelFunc
	Recorder record.EventRecorder
	Logger   logr.Logger
}

const (
	DatasourcesApiVersion = 1
	ControllerName        = "controller_grafanadatasource"
)

var log = logf.Log.WithName(ControllerName)

var _ reconcile.Reconciler = &GrafanaDatasourceReconciler{}

// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadatasources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadatasources/status,verbs=get;update;patch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GrafanaDatasourceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log = r.Logger.WithValues("grafanadatasource", request.NamespacedName)
	// Read the current state of known and cluster datasources
	currentState := common.NewDataSourcesState()
	err := currentState.Read(ctx, r.Client, request.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	if currentState.KnownDataSources == nil {
		log.Info("no datasources configmap found")
		return reconcile.Result{Requeue: false}, nil
	}

	// Reconcile all data sources
	err = r.reconcileDataSources(currentState)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: false}, nil
}

func (r *GrafanaDatasourceReconciler) reconcileDataSources(state *common.DataSourcesState) error {
	var dataSourcesToAddOrUpdate []grafanav1alpha1.GrafanaDataSource
	var dataSourcesToDelete []string

	// check if a given datasource (by its key) is found on the cluster
	foundOnCluster := func(key string) bool {
		for _, ds := range state.ClusterDataSources.Items {
			if key == ds.Filename() {
				return true
			}
		}
		return false
	}

	// Data sources to add or update: we always update the config map and let
	// Kubernetes figure out if any changes have to be applied
	dataSourcesToAddOrUpdate = append(dataSourcesToAddOrUpdate, state.ClusterDataSources.Items...)

	// Data sources to delete: if a datasourcedashboard is in the configmap but cannot
	// be found on the cluster then we assume it has been deleted and remove
	// it from the configmap
	for ds := range state.KnownDataSources.Data {
		if !foundOnCluster(ds) {
			dataSourcesToDelete = append(dataSourcesToDelete, ds)
		}
	}

	// apply dataSourcesToDelete
	for _, ds := range dataSourcesToDelete {
		log.Info("deleting datasource", "datasource", ds)
		if state.KnownDataSources.Data != nil {
			delete(state.KnownDataSources.Data, ds)
		}
	}

	// apply dataSourcesToAddOrUpdate
	var updated []grafanav1alpha1.GrafanaDataSource // nolint
	for i := range dataSourcesToAddOrUpdate {
		pipeline := NewDatasourcePipeline(&dataSourcesToAddOrUpdate[i])
		err := pipeline.ProcessDatasource(state.KnownDataSources)
		if err != nil {
			r.manageError(&dataSourcesToAddOrUpdate[i], err)
			continue
		}
		updated = append(updated, dataSourcesToAddOrUpdate[i])
	}

	// update the hash of the newly reconciled datasources
	hash, err := r.updateHash(state.KnownDataSources)
	if err != nil {
		r.manageError(nil, err)
		return err
	}

	if state.KnownDataSources.Annotations == nil {
		state.KnownDataSources.Annotations = map[string]string{}
	}

	// Compare the last hash to the previous one, update if changed
	lastHash := state.KnownDataSources.Annotations[constants.LastConfigAnnotation]
	if lastHash != hash {
		state.KnownDataSources.Annotations[constants.LastConfigAnnotation] = hash

		// finally, update the configmap
		err = r.Client.Update(r.Context, state.KnownDataSources)
		if err != nil {
			r.Recorder.Event(state.KnownDataSources, "Warning", "UpdateError", err.Error())
		} else {
			r.manageSuccess(updated)
		}
	}
	return nil
}

func (i *GrafanaDatasourceReconciler) updateHash(known *v1.ConfigMap) (string, error) {
	if known == nil || known.Data == nil {
		return "", nil
	}

	// Make sure that we always use the same order when creating the hash
	keys := make([]string, 0, len(known.Data))

	for key := range known.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	hash := sha256.New()
	for _, key := range keys {
		_, err := io.WriteString(hash, key)
		if err != nil {
			return "", err
		}

		_, err = io.WriteString(hash, known.Data[key])
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Handle error case: update datasource with error message and status
func (r *GrafanaDatasourceReconciler) manageError(datasource *grafanav1alpha1.GrafanaDataSource, issue error) {
	r.Recorder.Event(datasource, "Warning", "ProcessingError", issue.Error())

	// datasource deleted
	if datasource == nil {
		return
	}

	datasource.Status.Phase = grafanav1alpha1.PhaseFailing
	datasource.Status.Message = issue.Error()

	err := r.Client.Status().Update(r.Context, datasource)
	if err != nil {
		// Ignore conclicts. Resource might just be outdated.
		if k8serrors.IsConflict(err) {
			return
		}
		log.Error(err, "error updating datasource status")
	}
}

// manage success case: datasource has been imported successfully and the configmap
// is updated
func (r *GrafanaDatasourceReconciler) manageSuccess(datasources []grafanav1alpha1.GrafanaDataSource) {
	for i, datasource := range datasources {
		log.Info("datasource successfully imported",
			"datasource.Namespace", datasource.Namespace,
			"datasource.Name", datasource.Name)

		datasource.Status.Phase = grafanav1alpha1.PhaseReconciling
		datasource.Status.Message = "success"

		err := r.Client.Status().Update(r.Context, &datasources[i])
		if err != nil {
			r.Recorder.Event(&datasources[i], "Warning", "UpdateError", err.Error())
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDatasourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&integreatlyorgv1alpha1.GrafanaDataSource{}).
		Complete(r)
}
