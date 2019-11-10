package grafanadashboard

import (
	"context"
	"encoding/json"
	defaultErrors "errors"
	"fmt"
	"github.com/grafana-tools/sdk"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"net/http"
	"net/url"

	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ControllerName = "controller_grafanadashboard"
)

var log = logf.Log.WithName(ControllerName)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDashboard Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, autodetectChannel chan schema.GroupVersionKind) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	board := sdk.NewBoard("")
	log.Info(fmt.Sprintf("%v", board))

	return &ReconcileGrafanaDashboard{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		config:   config.GetControllerConfig(),
		helper:   common.NewKubeHelper(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafanadashboard-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDashboard
	err = c.Watch(&source.Kind{Type: &i8ly.GrafanaDashboard{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Info("Starting dashboard controller")
	}

	return err
}

var _ reconcile.Reconciler = &ReconcileGrafanaDashboard{}

// ReconcileGrafanaDashboard reconciles a GrafanaDashboard object
type ReconcileGrafanaDashboard struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	config   *config.ControllerConfig
	helper   *common.KubeHelperImpl
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDashboard) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	dashboardLabelSelectors := r.config.GetConfigItem(config.ConfigDashboardLabelSelector, nil)
	if dashboardLabelSelectors == nil {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
	}

	// Fetch the GrafanaDashboard instance
	instance := &i8ly.GrafanaDashboard{}
	err := r.client.Get(r.context, request.NamespacedName, instance)
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
	if match, err := cr.MatchesSelectors(dashboardLabelSelectors.([]*v1.LabelSelector)); err != nil {
		return reconcile.Result{}, err
	} else if !match {
		log.Info(fmt.Sprintf("found dashboard '%s/%s' but labels do not match", cr.Namespace, cr.Name))
		return reconcile.Result{}, nil
	}

	// Resource deleted?
	if cr.DeletionTimestamp != nil {
		return r.deleteDashboard(cr)
	}

	return r.importDashboard(cr)
}

func (r *ReconcileGrafanaDashboard) isValidDashboard(d *i8ly.GrafanaDashboard, dashboardJson string) bool {
	var js map[string]interface{}
	err := json.Unmarshal([]byte(dashboardJson), &js)
	return err == nil
}

func (r *ReconcileGrafanaDashboard) importDashboard(d *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	r.config.SetPluginsFor(d)
	return reconcile.Result{}, nil
}

func (r *ReconcileGrafanaDashboard) deleteDashboard(d *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	r.config.RemovePluginsFor(d)
	return reconcile.Result{}, nil
}

// Try to load remote dashboard json from an url
func (r *ReconcileGrafanaDashboard) loadDashboardFromURL(d *i8ly.GrafanaDashboard) (string, error) {
	_, err := url.ParseRequestURI(d.Spec.Url)
	if err != nil {
		return "", defaultErrors.New("dashboard url specified is not valid")
	}

	resp, err := http.Get(d.Spec.Url)
	if err != nil {
		return "", defaultErrors.New("request to import dashboard failed")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	response := string(body)

	if resp.StatusCode != 200 {
		return "", defaultErrors.New(fmt.Sprintf("request to import dashboard returned with status %v", resp.StatusCode))
	}

	return response, nil
}

func (r *ReconcileGrafanaDashboard) manageError(cr *i8ly.Grafana, issue error) (reconcile.Result, error) {
	r.recorder.Event(cr, "Warning", "ProcessingError", issue.Error())

	cr.Status.Phase = i8ly.PhaseFailing

	err := r.client.Update(r.context, cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileGrafanaDashboard) manageSuccess(cr *i8ly.Grafana) (reconcile.Result, error) {
	cr.Status.Phase = i8ly.PhaseReconciling
	err := r.client.Update(r.context, cr)
	if err != nil {
		return r.manageError(cr, err)
	}

	log.Info("desired cluster state met")
	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}
