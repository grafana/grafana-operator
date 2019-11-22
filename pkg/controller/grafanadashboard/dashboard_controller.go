package grafanadashboard

import (
	"context"
	defaultErrors "errors"
	"fmt"
	"github.com/grafana-tools/sdk"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	return &ReconcileGrafanaDashboard{
		client:   mgr.GetClient(),
		config:   config.GetControllerConfig(),
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
	config   *config.ControllerConfig
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

	client, err := r.getClient()
	if err != nil {
		return reconcile.Result{}, err
	}

	// Fetch the GrafanaDashboard instance
	instance := &i8ly.GrafanaDashboard{}
	err = r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("deleting dashboard %v/%v", request.Namespace, request.Name))
			return r.reconcileDashboards(request, client)
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

	return r.reconcileDashboards(request, client)
}

func (r *ReconcileGrafanaDashboard) reconcileDashboards(request reconcile.Request, grafanaClient GrafanaClient) (reconcile.Result, error) {
	// Collect known and namespace dashboards
	knownDashboards := r.config.GetDashboards(request.Namespace)
	namespaceDashboards := &i8ly.GrafanaDashboardList{}
	err := r.client.List(r.context, namespaceDashboards)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Prepare lists
	dashboardsToDelete := []i8ly.GrafanaDashboardRef{}
	inNamespace := func(item string) bool {
		for _, dashboard := range namespaceDashboards.Items {
			if dashboard.Name == item {
				return true
			}
		}
		return false
	}

	// Dashboards to delete: dashboards that are known but not found
	// any longer in the namespace
	for _, dashboard := range knownDashboards {
		if !inNamespace(dashboard.Name) {
			dashboardsToDelete = append(dashboardsToDelete, dashboard)
		}
	}

	// Process new/updated dashboards
	for _, dashboard := range namespaceDashboards.Items {
		pipeline := NewDashboardPipeline(&dashboard)
		processed, err := pipeline.ProcessDashboard()
		if err != nil {
			r.manageError(&dashboard, err)
			continue
		}

		status, err := grafanaClient.CreateOrUpdateDashboard(*processed)
		if err != nil {
			r.manageError(&dashboard, err)
			continue
		}

		err = r.manageSuccess(&dashboard, status)
		if err != nil {
			r.manageError(&dashboard, err)
		}
	}

	for _, dashboard := range dashboardsToDelete {
		status, err := grafanaClient.DeleteDashboardByUID(dashboard.UID)
		if err != nil {
			log.Error(err, fmt.Sprintf("error deleting dashboard %v/, status was %v/%v",
				dashboard.UID,
				status.Status,
				status.Message))
			continue
		}

		log.Info(fmt.Sprintf("delete result was %v", *status.Message))

		r.config.RemovePluginsFor(request.Namespace, request.Name)
		r.config.RemoveDashboard(request.Namespace, request.Name)
	}

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

// Handle success case: update dashboard metadata (id, uid) and update the list
// of plugins
func (r *ReconcileGrafanaDashboard) manageSuccess(dashboard *i8ly.GrafanaDashboard, status sdk.StatusMessage) error {
	log.Info(fmt.Sprintf("dashboard %v/%v successfully submitted",
		dashboard.Namespace,
		dashboard.Name))

	dashboard.Status.UID = *status.UID
	dashboard.Status.ID = *status.ID
	dashboard.Status.Slug = *status.Slug
	dashboard.Status.Phase = i8ly.PhaseReconciling

	r.config.AddDashboard(dashboard)
	r.config.SetPluginsFor(dashboard)

	return r.client.Status().Update(r.context, dashboard)
}

// Handle error case: update dashboard with error message and status
func (r *ReconcileGrafanaDashboard) manageError(dashboard *i8ly.GrafanaDashboard, issue error) {
	r.recorder.Event(dashboard, "Warning", "ProcessingError", issue.Error())
	dashboard.Status.Phase = i8ly.PhaseFailing
	dashboard.Status.Message = issue.Error()
	err := r.client.Status().Update(r.context, dashboard)
	if err != nil {
		log.Error(err, "error updating dashboard status")
	}
}

// Get an authenticated grafana API client
func (r *ReconcileGrafanaDashboard) getClient() (GrafanaClient, error) {
	url := r.config.GetConfigString(config.ConfigGrafanaAdminRoute, "")
	if url == "" {
		return nil, defaultErrors.New("cannot get grafana admin url")
	}

	username := r.config.GetConfigString(config.ConfigGrafanaAdminUsername, "")
	if username == "" {
		return nil, defaultErrors.New("invalid credentials (username)")
	}

	password := r.config.GetConfigString(config.ConfigGrafanaAdminPassword, "")
	if password == "" {
		return nil, defaultErrors.New("invalid credentials (password)")
	}

	return NewGrafanaClient(url, username, password), nil
}
