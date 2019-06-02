package grafanadatasource

import (
	"context"
	"fmt"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"gopkg.in/yaml.v2"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_grafanadatasource")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDataSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGrafanaDataSource{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		helper: common.NewKubeHelper(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafanadatasource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDataSource
	err = c.Watch(&source.Kind{Type: &i8ly.GrafanaDataSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGrafanaDataSource{}

// ReconcileGrafanaDataSource reconciles a GrafanaDataSource object
type ReconcileGrafanaDataSource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	helper *common.KubeHelperImpl
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDataSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &i8ly.GrafanaDataSource{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	if cr.DeletionTimestamp != nil {
		return r.DeleteDatasource(cr)
	}

	switch cr.Status.Phase {
	case common.StatusResourceUninitialized:
		// New resource
		return r.updatePhase(cr, common.StatusResourceSetFinalizer)
	case common.StatusResourceSetFinalizer:
		// Set finalizer first
		if len(cr.Finalizers) > 0 {
			return r.updatePhase(cr, common.StatusResourceCreated)
		} else {
			return r.setFinalizer(cr)
		}
	case common.StatusResourceCreated:
		return r.reconcileDatasource(cr)
	default:
		return reconcile.Result{}, nil
	}
}

func (r *ReconcileGrafanaDataSource) reconcileDatasource(cr *i8ly.GrafanaDataSource) (reconcile.Result, error) {
	ds, err := r.parseDataSource(cr)
	if err != nil {
		log.Error(err, "error parsing datasource")
		return reconcile.Result{}, err
	}

	updated, err := r.helper.UpdateDataSources(cr.Spec.Name, cr.Namespace, ds)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !updated {
		return reconcile.Result{RequeueAfter: time.Second * common.RequeueDelaySeconds}, err
	}

	err = r.helper.RestartGrafana(cr.Namespace)
	if err != nil {
		log.Error(err, "error restarting grafana")
	}

	log.Info(fmt.Sprintf("datasource '%s' updated", cr.Spec.Name))
	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDataSource) DeleteDatasource(cr *i8ly.GrafanaDataSource) (reconcile.Result, error) {
	err := r.helper.DeleteDataSources(cr.Spec.Name, cr.Namespace)
	if err != nil {
		log.Error(err, "error deleting datasource")
		return reconcile.Result{}, err
	}

	err = r.removeFinalizer(cr)
	if err != nil {
		log.Error(err, "error removing finalizer")
	}

	err = r.helper.RestartGrafana(cr.Namespace)
	if err != nil {
		log.Error(err, "error restarting grafana")
	}

	log.Info(fmt.Sprintf("datasource '%s' deleted", cr.Spec.Name))
	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDataSource) parseDataSource(cr *i8ly.GrafanaDataSource) (string, error) {
	datasources := struct {
		Datasources []i8ly.GrafanaDataSourceFields `json:"datasources"`
	}{
		Datasources: cr.Spec.Datasources,
	}

	bytes, err := yaml.Marshal(datasources)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (r *ReconcileGrafanaDataSource) setFinalizer(cr *i8ly.GrafanaDataSource) (reconcile.Result, error) {
	if len(cr.Finalizers) == 0 {
		cr.Finalizers = append(cr.Finalizers, common.ResourceFinalizerName)
	}
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDataSource) removeFinalizer(cr *i8ly.GrafanaDataSource) error {
	cr.Finalizers = nil
	return r.client.Update(context.TODO(), cr)
}

func (r *ReconcileGrafanaDataSource) updatePhase(cr *i8ly.GrafanaDataSource, phase int) (reconcile.Result, error) {
	cr.Status.Phase = phase
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}
