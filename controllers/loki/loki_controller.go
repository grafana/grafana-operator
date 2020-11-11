package loki

import (
	"context"
	"fmt"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	"time"
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

	ref := r.(*ReconcileLoki)
	ticker := time.NewTicker(config.RequeueDelay)
	sendEmptyRequest := func() {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "",
			},
		}
		_, _ = r.Reconcile(request)
	}

	go func() {
		for range ticker.C {
			log.V(1).Info("running periodic loki resync")
			sendEmptyRequest()
		}
	}()

	go func() {
		for stateChange := range common.ControllerEvents {
			// Controller state updated
			ref.state = stateChange
		}
	}()

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

	// Initial request?
	if request.Name == "" {
		return r.reconcileLoki(request)
	}

	// Check if the label selectors are available yet. If not then the grafana controller
	// has not finished initializing and we can't continue. Reschedule for later.
	if r.state.LokiSelector == nil {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
	}

	// Fetch the GrafanaDashboard instance
	instance := &grafanav1alpha1.Loki{}
	err := r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {

		if errors.IsNotFound(err) {
			// If some dashboard has been deleted, then always re sync the world
			log.Info(fmt.Sprintf("Reconciling loki %v/%v", request.Namespace, request.Name))
			return r.reconcileLoki(request)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the dashboard does not match the label selectors then we ignore it
	cr := instance.DeepCopy()
	if !r.isMatch(cr) {
		log.V(1).Info(fmt.Sprintf("loki %v/%v found but selectors do not match",
			cr.Namespace, cr.Name))
		return reconcile.Result{}, nil
	}

	// Otherwise always re sync all dashboards in the namespace
	//return r.reconcileLoki(request)
	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

func (r *ReconcileLoki) reconcileLoki(request reconcile.Request) (reconcile.Result, error) {
	// Collect known and namespace dashboards
	knownLokis := r.config.GetLokis(request.Namespace)
	namespaceLokis := &grafanav1alpha1.LokiList{}

	opts := &client.ListOptions{
		Namespace: request.Namespace,
	}

	err := r.client.List(r.context, namespaceLokis, opts)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Prepare lists
	var LokisToDelete []*grafanav1alpha1.LokiRef

	// Check if a given dashboard (by name) is present in the list of
	// dashboards in the namespace
	inNamespace := func(item *grafanav1alpha1.LokiRef) bool {
		for _, l := range namespaceLokis.Items {
			if l.Name == item.Name && l.Namespace == item.Namespace {
				return true
			}
		}
		return false
	}

	// Dashboards to delete: dashboards that are known but not found
	// any longer in the namespace
	for _, loki := range knownLokis {
		if !inNamespace(loki) {
			LokisToDelete = append(LokisToDelete, loki)
		}

	}

	// Process new/updated dashboards
	for _, loki := range namespaceLokis.Items {
		// Is this a dashboard we care about (matches the label selectors)?
		if !r.isMatch(&loki) {
			log.V(1).Info(fmt.Sprintf("loki %v/%v found but selectors do not match",
				loki.Namespace, loki.Name))
			continue
		}

		pipeline := NewLokiPipeline(&loki, r.context)
		_, err := pipeline.ProcessLoki(loki)

		if err != nil {
			log.Error(err, fmt.Sprintf("cannot process Loki %v/%v", loki.Namespace, loki.Name))
			r.manageError(&loki, err)
			continue
		}

		// Check labels only when DashboardNamespaceSelector isnt empty
		if r.state.LokiNamespaceSelector != nil {
			matchesNamespaceLabels, err := r.checkNamespaceLabels(&loki)
			if err != nil {
				r.manageError(&loki, err)
				continue
			}

			if matchesNamespaceLabels == false {
				log.Info(fmt.Sprintf("loki %v skipped because the namespace labels do not match", loki.Name))
				continue
			}
		}

		r.manageSuccess(&loki)
	}

	if LokisToDelete != nil {
		for _, loki := range LokisToDelete {

			r.config.RemoveLoki(loki.Namespace, loki.Name)

			log.V(1).Info("loki removed")
			// Refresh the list of known lokis
			knownLokis = r.config.GetLokis(request.Namespace)

		}
	}

	// Mark the dashboards as synced so that the current state can be written
	// to the Grafana CR by the grafana controller
	r.config.AddConfigItem(config.ConfigGrafanaDashboardsSynced, true)

	return reconcile.Result{Requeue: false}, nil
}

// Test if a given dashboard matches an array of label selectors
func (r *ReconcileLoki) isMatch(item *grafanav1alpha1.Loki) bool {

	match, err := item.MatchesSelectors(r.state.LokiSelector)
	if err != nil {
		log.Error(err, fmt.Sprintf("error matching selectors against %v/%v",
			item.Namespace,
			item.Name))
		return false
	}
	return match
}

// Handle success case: update dashboard metadata (id, uid) and update the list
// of plugins
func (r *ReconcileLoki) manageSuccess(loki *grafanav1alpha1.Loki, ) {
	msg := fmt.Sprintf("dashboard %v/%v successfully submitted",
		loki.Namespace,
		loki.Name)
	r.recorder.Event(loki, "Normal", "Success", msg)
	log.Info(msg)
	r.config.AddLoki(loki)
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

// check if the labels on a namespace match a given label selector
func (r *ReconcileLoki) checkNamespaceLabels(loki *grafanav1alpha1.Loki) (bool, error) {
	key := client.ObjectKey{
		Name: loki.Namespace,
	}
	ns := &v1.Namespace{}
	err := r.client.Get(r.context, key, ns)
	if err != nil {
		return false, err
	}
	selector, err := metav1.LabelSelectorAsSelector(r.state.LokiNamespaceSelector)

	if err != nil {
		return false, err
	}

	return selector.Empty() || selector.Matches(labels.Set(ns.Labels)), nil
}
