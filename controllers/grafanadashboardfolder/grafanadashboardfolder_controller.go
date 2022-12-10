package grafanadashboardfolder

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"net/http"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v4/controllers/common"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ControllerName = "controller_grafanadashboardfolder"
)

// GrafanaDashboardFolderReconciler reconciles a GrafanaFolder object
type GrafanaDashboardFolderReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver

	Client    client.Client
	Scheme    *runtime.Scheme
	transport *http.Transport
	config    *config.ControllerConfig
	context   context.Context
	cancel    context.CancelFunc
	recorder  record.EventRecorder
	state     common.ControllerState
	Log       logr.Logger
}

// Add creates a new GrafanaDashboardFolder Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, namespace string) error {
	return SetupWithManager(mgr, newReconciler(mgr), namespace)
}

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr ctrl.Manager, r reconcile.Reconciler, namespace string) error {
	c, err := controller.New("grafanadashboardfolder-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDashboard
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.GrafanaFolder{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Log.Info("Starting dashboardfolder controller")
	}

	ref := r.(*GrafanaDashboardFolderReconciler) //nolint
	ticker := time.NewTicker(config.GetControllerConfig().RequeueDelay)
	sendEmptyRequest := func() {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "",
			},
		}
		_, err = r.Reconcile(ref.context, request)
		if err != nil {
			return
		}
	}

	go func() {
		for range ticker.C {
			log.Log.Info("running periodic dashboardfolder resync")
			sendEmptyRequest()
		}
	}()

	go func() {
		for stateChange := range common.DashboardFolderControllerEvents {
			// Controller state updated
			ref.state = stateChange
		}
	}()
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1alpha1.GrafanaFolder{}).
		Complete(r)
}

var _ reconcile.Reconciler = &GrafanaDashboardFolderReconciler{}

// +kubebuilder:rbac:groups=integreatly.org,resources=grafanafolders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanafolders/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the GrafanaDashboard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GrafanaDashboardFolderReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues(ControllerName, request.NamespacedName)

	// If Grafana is not running there is no need to continue
	if !r.state.GrafanaReady {
		logger.Info("no grafana instance available")
		return reconcile.Result{Requeue: false}, nil
	}

	grafanaClient, err := r.getClient()
	if err != nil {
		return reconcile.Result{RequeueAfter: config.GetControllerConfig().RequeueDelay}, err
	}

	// Initial request?
	if request.Name == "" {
		return r.reconcileDashboardFolders(request, grafanaClient)
	}

	// Check if the label selectors are available yet. If not then the grafana controller
	// has not finished initializing and we can't continue. Reschedule for later.
	if r.state.DashboardSelectors == nil {
		return reconcile.Result{RequeueAfter: config.GetControllerConfig().RequeueDelay}, nil
	}

	// Fetch the GrafanaDashboard instance
	instance := &grafanav1alpha1.GrafanaFolder{}
	err = r.Client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// If some dashboard has been deleted, then always re sync the world
			logger.Info("deleting dashboard", "namespace", request.Namespace, "name", request.Name)
			return r.reconcileDashboardFolders(request, grafanaClient)
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the dashboard does not match the label selectors then we ignore it
	cr := instance.DeepCopy()
	if !r.isMatch(cr) {
		logger.V(1).Info(fmt.Sprintf("folder %v/%v found but selectors do not match", cr.Namespace, cr.Name))
		return ctrl.Result{}, nil
	}
	// Otherwise always re sync all dashboards in the namespace
	return r.reconcileDashboardFolders(request, grafanaClient)
}

func (r *GrafanaDashboardFolderReconciler) reconcileDashboardFolders(request ctrl.Request, grafanaClient GrafanaClient) (reconcile.Result, error) {
	foldersInNamespace := &grafanav1alpha1.GrafanaFolderList{}

	opts := &client.ListOptions{
		Namespace: request.Namespace,
	}
	err := r.Client.List(r.context, foldersInNamespace, opts)
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range foldersInNamespace.Items {
		folder := foldersInNamespace.Items[i]
		if !r.isMatch(&folder) {
			log.Log.Info("dashboard found but selectors do not match", "namespace", folder.Namespace, "name", folder.Name)
			continue
		}
		_, err := grafanaClient.ApplyFolderPermissions(folder.Spec.FolderName, folder.GetPermissions())
		if err != nil {
			r.manageError(&folder, err)
			continue
		}

		r.manageSuccess(&folder)
	}

	// Mark the folders as synced so that the current state can be written
	// to the Grafana CR by the grafana controller
	r.config.AddConfigItem(config.ConfigGrafanaFoldersSynced, true)

	return reconcile.Result{Requeue: false}, nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &GrafanaDashboardFolderReconciler{
		Client: mgr.GetClient(),
		/* #nosec G402 */
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Log:      mgr.GetLogger(),
		config:   config.GetControllerConfig(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(ControllerName),
		state:    common.ControllerState{},
	}
}

// Get an authenticated grafana API client
func (r *GrafanaDashboardFolderReconciler) getClient() (GrafanaClient, error) {
	url := r.state.AdminUrl
	if url == "" {
		return nil, errors.New("cannot get grafana admin url")
	}

	username := os.Getenv(constants.GrafanaAdminUserEnvVar)
	if username == "" {
		return nil, errors.New("invalid credentials (username)")
	}

	password := os.Getenv(constants.GrafanaAdminPasswordEnvVar)
	if password == "" {
		return nil, errors.New("invalid credentials (password)")
	}

	duration := time.Duration(r.state.ClientTimeout)

	return NewGrafanaClient(url, username, password, r.transport, duration), nil
}

// Handle success case: update dashboardfolder metadata (name, hash)
func (r *GrafanaDashboardFolderReconciler) manageSuccess(folder *grafanav1alpha1.GrafanaFolder) {
	msg := fmt.Sprintf("folder %v/%v successfully submitted", folder.Namespace, folder.Name)
	r.recorder.Event(folder, "Normal", "Success", msg)
	log.Log.Info("folder successfully submitted", "name", folder.Name, "namespace", folder.Namespace)
	r.config.AddFolder(folder)
}

// Handle error case: update dashboardfolder with error message and status
func (r *GrafanaDashboardFolderReconciler) manageError(folder *grafanav1alpha1.GrafanaFolder, issue error) {
	r.recorder.Event(folder, "Warning", "ProcessingError", issue.Error())
	// Ignore conflicts. Resource might just be outdated, also ignore if grafana isn't available.
	if k8serrors.IsConflict(issue) || k8serrors.IsServiceUnavailable(issue) {
		return
	}
	log.Log.Error(issue, "error updating folder", "name", folder.Name, "namespace", folder.Namespace)
}

// Test if a given dashboardfolder matches an array of label selectors
func (r *GrafanaDashboardFolderReconciler) isMatch(item *grafanav1alpha1.GrafanaFolder) bool {
	if r.state.DashboardSelectors == nil {
		return false
	}

	match, err := item.MatchesSelectors(r.state.DashboardSelectors)
	if err != nil {
		log.Log.Error(err, "error matching selectors",
			"item.Namespace", item.Namespace,
			"item.Name", item.Name)
		return false
	}
	return match
}

func (r *GrafanaDashboardFolderReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1alpha1.GrafanaFolder{}).
		Complete(r)
}
