package grafanadatasource

import (
	"context"
	"crypto/sha256"
	"fmt"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sort"
	"time"
)

const (
	DatasourcesApiVersion = 1
	ControllerName        = "controller_grafanadatasource"
)

var log = logf.Log.WithName(ControllerName)

// Add creates a new GrafanaDataSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ chan schema.GroupVersionKind, namespace string) error {
	return add(mgr, newReconciler(mgr), namespace)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileGrafanaDataSource{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		state:    common.ControllerState{},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, namespace string) error {
	// Create a new controller
	c, err := controller.New("grafanadatasource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDataSource
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.GrafanaDataSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	ref := r.(*ReconcileGrafanaDataSource)

	// The datasources should not change very often, only revisit them
	// half as often as the dashboards
	ticker := time.NewTicker(config.RequeueDelay * 2)
	sendEmptyRequest := func() {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "",
			},
		}
		r.Reconcile(request)
	}

	go func() {
		for range ticker.C {
			log.Info("running periodic datasource resync")
			sendEmptyRequest()
		}
	}()

	// Listen for config change events
	go func() {
		for stateChange := range common.ControllerEvents {
			// Controller state updated
			ref.state = stateChange
		}
	}()

	return nil
}

var _ reconcile.Reconciler = &ReconcileGrafanaDataSource{}

// ReconcileGrafanaDataSource reconciles a GrafanaDataSource object
type ReconcileGrafanaDataSource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
	state    common.ControllerState
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDataSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Read the current state of known and cluster datasources
	currentState := common.NewDataSourcesState()
	err := currentState.Read(r.context, r.client, request.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	if currentState.KnownDataSources == nil {
		log.Info(fmt.Sprintf("no datasources configmap found"))
		return reconcile.Result{Requeue: false}, nil
	}

	// Reconcile all data sources
	err = r.reconcileDataSources(currentState)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileGrafanaDataSource) reconcileDataSources(state *common.DataSourcesState) error {
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
	for _, ds := range state.ClusterDataSources.Items {
		dataSourcesToAddOrUpdate = append(dataSourcesToAddOrUpdate, ds)
	}

	// Data sources to delete: if a datasourcedashboard is in the configmap but cannot
	// be found on the cluster then we assume it has been deleted and remove
	// it from the configmap
	for ds, _ := range state.KnownDataSources.Data {
		if !foundOnCluster(ds) {
			dataSourcesToDelete = append(dataSourcesToDelete, ds)
		}
	}

	// apply dataSourcesToDelete
	for _, ds := range dataSourcesToDelete {
		log.Info(fmt.Sprintf("deleting datasource %v", ds))
		if state.KnownDataSources.Data != nil {
			delete(state.KnownDataSources.Data, ds)
		}
	}

	// apply dataSourcesToAddOrUpdate
	updated := []grafanav1alpha1.GrafanaDataSource{}
	for _, ds := range dataSourcesToAddOrUpdate {
		pipeline := NewDatasourcePipeline(&ds)
		err := pipeline.ProcessDatasource(state.KnownDataSources)
		if err != nil {
			r.manageError(&ds, err)
			continue
		}
		updated = append(updated, ds)
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
	lastHash := state.KnownDataSources.Annotations[model.LastConfigAnnotation]
	if lastHash != hash {
		state.KnownDataSources.Annotations[model.LastConfigAnnotation] = hash

		// finally, update the configmap
		err = r.client.Update(r.context, state.KnownDataSources)
		if err != nil {
			r.recorder.Event(state.KnownDataSources, "Warning", "UpdateError", err.Error())
		} else {
			r.manageSuccess(updated)
		}
	}
	return nil
}

func (i *ReconcileGrafanaDataSource) updateHash(known *v1.ConfigMap) (string, error) {
	if known == nil || known.Data == nil {
		return "", nil
	}

	// Make sure that we always use the same order when creating the hash
	var keys []string
	for key, _ := range known.Data {
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
func (r *ReconcileGrafanaDataSource) manageError(datasource *grafanav1alpha1.GrafanaDataSource, issue error) {
	r.recorder.Event(datasource, "Warning", "ProcessingError", issue.Error())

	// datasource deleted
	if datasource == nil {
		return
	}

	datasource.Status.Phase = grafanav1alpha1.PhaseFailing
	datasource.Status.Message = issue.Error()

	err := r.client.Status().Update(r.context, datasource)
	if err != nil {
		// Ignore conclicts. Resource might just be outdated.
		if errors.IsConflict(err) {
			return
		}
		log.Error(err, "error updating datasource status")
	}
}

// manage success case: datasource has been imported successfully and the configmap
// is updated
func (r *ReconcileGrafanaDataSource) manageSuccess(datasources []grafanav1alpha1.GrafanaDataSource) {
	for _, datasource := range datasources {
		log.Info(fmt.Sprintf("datasource %v/%v successfully imported",
			datasource.Namespace,
			datasource.Name))

		datasource.Status.Phase = grafanav1alpha1.PhaseReconciling
		datasource.Status.Message = "success"

		err := r.client.Status().Update(r.context, &datasource)
		if err != nil {
			r.recorder.Event(&datasource, "Warning", "UpdateError", err.Error())
		}
	}
}
