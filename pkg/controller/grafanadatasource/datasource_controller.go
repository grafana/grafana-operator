package grafanadatasource

import (
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	DatasourcesApiVersion = 1
	ControllerName        = "controller_grafanadatasource"
)

var log = logf.Log.WithName(ControllerName)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDataSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, autodetectChannel chan schema.GroupVersionKind) error {
	return add(mgr, newReconciler(mgr))
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
func add(mgr manager.Manager, r reconcile.Reconciler) error {
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

	log.Info("Starting datasource controller")

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
	_, err := r.checkForDeletedDataSources(request)
	if err != nil{
		log.Info("error deleting")
		return reconcile.Result{}, err
	}

	instance := &i8ly.GrafanaDataSource{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	cr := instance.DeepCopy()
	res, err := r.reconcileDatasource(cr)

	// Requeue periodically to find datasources that have not been updated
	// but are not yet imported (can happen if Grafana is uninstalled and
	// then reinstalled without an Operator restart
	res.RequeueAfter = config.RequeueDelay
	return res, err
}

func (r *ReconcileGrafanaDataSource) getKnownDatasources(request reconcile.Request) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	selector := client.ObjectKey{
		Namespace: request.Namespace,
		Name:      config.GrafanaDatasourcesConfigMapName,
	}

	err := r.client.Get(r.context, selector, configMap)
	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *ReconcileGrafanaDataSource) checkForDeletedDataSources(request reconcile.Request) (reconcile.Result, error) {
	datasources := &i8ly.GrafanaDataSourceList{}
	opts := client.ListOptions{
		Namespace: request.Namespace,
	}
	err := r.client.List(r.context, datasources, &opts)
	if err != nil {
		return reconcile.Result{}, err
	}

	known, err := r.getKnownDatasources(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	hasCr := func (key string) bool {
		for _, datasource := range datasources.Items {
			dsKey := fmt.Sprintf("%v_%v", datasource.Namespace, datasource.Spec.Name)
			if key == dsKey {
				return true
			}
		}
		return false
	}

	for key, _ := range known.Data {
		if hasCr(key) {
			continue
		}
		err = r.helper.DeleteDataSources(key)
		if err != nil {
			log.Info(fmt.Sprintf("failed to delete datasource %v", key))
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafanaDataSource) reconcileDatasource(cr *i8ly.GrafanaDataSource) (reconcile.Result, error) {
	ds, err := r.parseDataSource(cr)
	if err != nil {
		log.Error(err, "error parsing datasource")
		return reconcile.Result{}, err
	}

	_, uerr := r.helper.UpdateDataSources(cr.Spec.Name, cr.Namespace, ds)
	if uerr != nil {
		return reconcile.Result{}, uerr
	}

	log.Info("updated datasource")
	return reconcile.Result{}, err
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
