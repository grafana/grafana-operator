package grafanadatasource

import (
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
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
	"strings"
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
		helper:   common.NewKubeHelper(),
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
		log.Info("failed to instantiate datasource manager")
		return err
	}

	// Watch for changes to primary resource GrafanaDataSource
	err = c.Watch(&source.Kind{Type: &i8ly.GrafanaDataSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	ref := r.(*ReconcileGrafanaDataSource)
	ticker := time.NewTicker(config.RequeueDelay)
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
	helper   *common.KubeHelperImpl
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
	state    common.ControllerState
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDataSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// If Grafana is not running there is no need to continue
	if r.state.GrafanaReady == false {
		log.Info("no grafana instance available")
		return reconcile.Result{Requeue: false}, nil
	}

	// Read the current state of known and cluster datasources
	currentState := common.NewDataSourcesState()
	err := currentState.Read(r.context, r.client, request.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileDataSources(currentState)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafanaDataSource) reconcileDataSources(state *common.DataSourcesState) error {
	var dataSourcesToAddOrUpdate []i8ly.GrafanaDataSource
	var dataSourcesToDelete []string

	// datasources in the configmap are identified by their namespace and name
	// concatenated
	getKey := func(ds *i8ly.GrafanaDataSource) string {
		return fmt.Sprintf("%s_%s", ds.Namespace, strings.ToLower(ds.Name))
	}

	// check if a given datasource (by its key) is found on the cluster
	foundOnCluster := func(key string) bool {
		for _, ds := range state.ClusterDataSources.Items {
			if key == getKey(&ds) {
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

	// Apply dataSourcesToDelete
	for _, ds := range dataSourcesToDelete {
		log.Info(fmt.Sprintf("deleting datasource %v", ds))
		delete(state.KnownDataSources.Data, ds)
	}

	// Apply dataSourcesToAddOrUpdate
	for _, ds := range dataSourcesToAddOrUpdate {
		key := getKey(&ds)
		val, err := r.parseDataSource(&ds)
		if err != nil {
			log.Error(err, "error parsing datasource")
			r.manageError(&ds, err)
			continue
		}

		log.Info(fmt.Sprintf("importing datasource %v", key))
		state.KnownDataSources.Data[key] = val
	}

	// Update the configmap
	r.client.Update(r.context, state.KnownDataSources)

	return nil
}

func (r *ReconcileGrafanaDataSource) parseDataSource(cr *i8ly.GrafanaDataSource) (string, error) {
	datasources := struct {
		ApiVersion  int                            `json:"apiVersion"`
		Datasources []i8ly.GrafanaDataSourceFields `json:"datasources"`
	}{
		ApiVersion:  DatasourcesApiVersion,
		Datasources: cr.Spec.Datasources,
	}

	bytes, err := yaml.Marshal(datasources)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Handle error case: update datasource with error message and status
func (r *ReconcileGrafanaDataSource) manageError(datasource *i8ly.GrafanaDataSource, issue error) {
	r.recorder.Event(datasource, "Warning", "ProcessingError", issue.Error())
	datasource.Status.Phase = i8ly.PhaseFailing
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
func (r *ReconcileGrafanaDataSource) manageSuccess(datasource *i8ly.GrafanaDataSource) error {
	log.Info(fmt.Sprintf("datasource %v/%v successfully submitted",
		datasource.Namespace,
		datasource.Name))

	datasource.Status.Phase = i8ly.PhaseReconciling

	return r.client.Status().Update(r.context, datasource)
}
