package loki

import (
	"context"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ControllerName = "controller-loki"
)

var log = logf.Log.WithName(ControllerName)

// Add creates a new GrafanaDataSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ chan schema.GroupVersionKind, namespace string) error {
	return add(mgr, newReconciler(mgr), namespace)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, namespace string) error {
	// Create a new controller
	c, err := controller.New("loki-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDashboard
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.Loki{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Info("Starting loki controller")
	}

	return err
}

var _ reconcile.Reconciler = &ReconcileLoki{}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileLoki{
		client:   mgr.GetClient(),
		config:   config.GetControllerConfig(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		state:    common.ControllerState{},
	}
}

var _ reconcile.Reconciler = &ReconcileLoki{}

type ReconcileLoki struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	config   *config.ControllerConfig
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
	state    common.ControllerState
}

func (r ReconcileLoki) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.V(1).Info("running loki controller")

	// Fetch the Loki instance
	instance := &grafanav1alpha1.Loki{}
	err := r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return r.reconcileLoki(instance)
}

func (r *ReconcileLoki) reconcileLoki(loki *grafanav1alpha1.Loki) (reconcile.Result, error) {
	if loki.Spec.External != nil {
		loki.Status.Url = loki.Spec.External.Url
		return r.manageSuccess(loki, grafanav1alpha1.PhaseReconciling, "External Loki reconciled")
	} else {

		cr := loki.DeepCopy()

		// Read current state
		currentState := common.NewLokiState()
		err := currentState.Read(r.context, loki, r.client)
		if err != nil {
			log.Error(err, "error reading state")
			return r.manageError(loki, err)
		}

		// Get the actions required to reach the desired state
		reconciler := NewLokiReconciler()
		desiredState := reconciler.Reconcile(currentState, cr)

		// Run the actions to reach the desired state
		actionRunner := common.NewClusterActionRunner(r.context, r.client, r.scheme, cr)
		err = actionRunner.RunAll(desiredState)
		if err != nil {
			return r.manageError(cr, err)
		}
	}

	return reconcile.Result{}, nil
}

// Handle success case: update dashboard metadata (id, uid) and update the list
// of plugins
func (r *ReconcileLoki) manageSuccess(loki *grafanav1alpha1.Loki, phase grafanav1alpha1.StatusPhase, message string) (reconcile.Result, error) {
	current := loki.Status.DeepCopy()
	loki.Status.Phase = phase
	loki.Status.Message = message

	if !reflect.DeepEqual(current, &loki.Status) {
		err := r.client.Status().Update(r.context, loki)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileLoki) manageError(cr *grafanav1alpha1.Loki, issue error) (reconcile.Result, error) {
	r.recorder.Event(cr, "Warning", "ProcessingError", issue.Error())
	cr.Status.Phase = grafanav1alpha1.PhaseFailing
	cr.Status.Message = issue.Error()

	err := r.client.Status().Update(r.context, cr)
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			err = nil
		}
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}
