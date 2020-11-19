package loki

import (
	"context"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"k8s.io/apimachinery/pkg/api/errors"
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
	config   *config.ControllerConfig
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
	state    common.ControllerState
}

func (r ReconcileLoki) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.V(1).Info("running loki controller")

	// Fetch the GrafanaDashboard instance
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
		return r.manageSuccess(loki, grafanav1alpha1.PhaseReconciling, "")
	} else {
		// TODO
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

// Handle error case: update dashboard with error message and status
func (r *ReconcileLoki) manageError(loki *grafanav1alpha1.Loki, issue error) {
	r.recorder.Event(loki, "Warning", "ProcessingError", issue.Error())

	// Ignore conclicts. Resource might just be outdated.
	if errors.IsConflict(issue) {
		return
	}

	log.Error(issue, "error updating loki")
}
