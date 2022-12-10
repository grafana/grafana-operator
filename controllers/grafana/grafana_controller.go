package grafana

import (
	"context"
	stdErr "errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/common"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	"github.com/grafana-operator/grafana-operator/v4/controllers/model"
	routev1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ControllerName              = "grafana-controller"
	DefaultClientTimeoutSeconds = 5
)

var log = logf.Log.WithName(ControllerName)

// +kubebuilder:rbac:groups=integreatly.org,resources=grafanas;grafanas/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions;apps,resources=deployments;deployments/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts;services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileGrafana) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1alpha1.Grafana{}).
		Owns(&grafanav1alpha1.Grafana{}).
		Complete(r)
}

// Add creates a new Grafana Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, autodetectChannel chan schema.GroupVersionKind, _ string) error {
	return add(mgr, NewReconciler(mgr), autodetectChannel)
}

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileGrafana{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Plugins:  NewPluginsHelper(),
		Context:  ctx,
		Cancel:   cancel,
		Config:   config.GetControllerConfig(),
		Recorder: mgr.GetEventRecorderFor(ControllerName),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, autodetectChannel chan schema.GroupVersionKind) error {
	// Create a new controller
	c, err := controller.New("grafana-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Grafana
	err = c.Watch(&source.Kind{Type: &grafanav1alpha1.Grafana{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err = watchSecondaryResource(c, &v12.Deployment{}); err != nil {
		return err
	}

	if err = watchSecondaryResource(c, &netv1.Ingress{}); err != nil {
		return err
	}

	if err = watchSecondaryResource(c, &v1.ConfigMap{}); err != nil {
		return err
	}

	if err = watchSecondaryResource(c, &v1.Service{}); err != nil {
		return err
	}

	if err = watchSecondaryResource(c, &v1.ServiceAccount{}); err != nil {
		return err
	}

	go func() {
		for gvk := range autodetectChannel {
			cfg := config.GetControllerConfig()

			// Route already watched?
			if cfg.GetConfigBool(config.ConfigRouteWatch, false) {
				return
			}

			// Watch routes if they exist on the cluster
			if gvk.String() == routev1.SchemeGroupVersion.WithKind(common.RouteKind).String() {
				if err = watchSecondaryResource(c, &routev1.Route{}); err != nil {
					log.Error(err, fmt.Sprintf("error adding secondary watch for %v", common.RouteKind))
				} else {
					cfg.AddConfigItem(config.ConfigRouteWatch, true)
					log.V(1).Info(fmt.Sprintf("added secondary watch for %v", common.RouteKind))
				}
			}
		}
	}()

	return nil
}

var _ reconcile.Reconciler = &ReconcileGrafana{}

// ReconcileGrafana reconciles a Grafana object
type ReconcileGrafana struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client   client.Client
	Scheme   *runtime.Scheme
	Plugins  *PluginsHelperImpl
	Context  context.Context
	Log      logr.Logger
	Cancel   context.CancelFunc
	Config   *config.ControllerConfig
	Recorder record.EventRecorder
}

func watchSecondaryResource(c controller.Controller, resource client.Object) error {
	return c.Watch(&source.Kind{Type: resource}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &grafanav1alpha1.Grafana{},
	})
}

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
func (r *ReconcileGrafana) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Stop the dashboard controller from reconciling when grafana is not installed
			r.Config.RemoveConfigItem(config.ConfigDashboardLabelSelector)
			r.Config.Cleanup(true)

			common.ControllerEvents <- common.ControllerState{
				GrafanaReady: false,
			}

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	log.V(1).Info("Found grafana-instance, proceed Reconcile with deepcopy...")
	cr := instance.DeepCopy()

	log.V(1).Info("determine clusterState...")
	// Read current state
	currentState := common.NewClusterState()
	err = currentState.Read(ctx, cr, r.Client)
	if err != nil {
		log.Error(err, "error reading state")
		return r.manageError(cr, err, request)
	}

	// Get the actions required to reach the desired state
	log.V(1).Info("Create GrafanaReconciler and determine desiredState...")
	reconciler := NewGrafanaReconciler()
	desiredState := reconciler.Reconcile(currentState, cr)

	log.V(1).Info("Determined desiredStates - starting actionRunner")
	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(ctx, r.Client, r.Scheme, cr)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(cr, err, request)
	}

	// Run the config map reconciler to discover jsonnet libraries
	err = reconcileConfigMaps(cr, r)
	if err != nil {
		return r.manageError(cr, err, request)
	}

	return r.manageSuccess(cr, currentState, request)
}

func (r *ReconcileGrafana) manageError(cr *grafanav1alpha1.Grafana, issue error, request reconcile.Request) (reconcile.Result, error) {
	r.Recorder.Event(cr, "Warning", "ProcessingError", issue.Error())
	cr.Status.Phase = grafanav1alpha1.PhaseFailing
	cr.Status.Message = issue.Error()

	log.Error(issue, "error processing GrafanaInstance", "name", cr.Name, "namespace", cr.Namespace)

	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, request.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(cr.Status, instance.Status) {
		err := r.Client.Status().Update(r.Context, cr)
		if err != nil {
			// Ignore conflicts, resource might just be outdated.
			if errors.IsConflict(err) {
				err = nil
			}
			return reconcile.Result{}, err
		}
	}

	r.Config.InvalidateDashboards()

	common.ControllerEvents <- common.ControllerState{
		GrafanaReady: false,
	}

	return reconcile.Result{RequeueAfter: r.Config.RequeueDelay}, nil
}

// Try to find a suitable url to grafana
func (r *ReconcileGrafana) getGrafanaAdminUrl(cr *grafanav1alpha1.Grafana, state *common.ClusterState) (string, error) {
	// If preferService is true, we skip the routes and try to access grafana
	// by using the service.
	preferService := cr.GetPreferServiceValue()

	// First try to use the route if it exists. Prefer the route because it also works
	// when running the operator outside of the cluster
	if state.GrafanaRoute != nil && !preferService {
		return fmt.Sprintf("https://%v", state.GrafanaRoute.Spec.Host), nil
	}

	// Try the ingress first if on vanilla Kubernetes
	if state.GrafanaIngress != nil && !preferService {
		// If provided, use the hostname from the CR
		if cr.Spec.Ingress != nil && cr.Spec.Ingress.Hostname != "" {
			return fmt.Sprintf("https://%v", cr.Spec.Ingress.Hostname), nil
		}

		// Otherwise try to find something suitable, hostname or IP
		if len(state.GrafanaIngress.Status.LoadBalancer.Ingress) > 0 {
			ingress := state.GrafanaIngress.Status.LoadBalancer.Ingress[0]
			if ingress.Hostname != "" {
				return fmt.Sprintf("https://%v", ingress.Hostname), nil
			}
			return fmt.Sprintf("https://%v", ingress.IP), nil
		}
	}

	servicePort := int32(model.GetGrafanaPort(cr))

	// Otherwise rely on the service
	if state.GrafanaService != nil {
		protocol := "http"

		if cr.Spec.Config.Server != nil {
			switch cr.Spec.Config.Server.Protocol {
			case "", "http":
				protocol = "http"
			case "https":
				protocol = "https"
			default:
				return "", fmt.Errorf("server protocol %v is not supported, please use either http or https", cr.Spec.Config.Server.Protocol)
			}
		}

		return fmt.Sprintf("%v://%v.%v:%d", protocol, state.GrafanaService.Name, cr.Namespace,
			servicePort), nil
	}

	return "", stdErr.New("failed to find admin url")
}

func (r *ReconcileGrafana) manageSuccess(cr *grafanav1alpha1.Grafana, state *common.ClusterState, request reconcile.Request) (reconcile.Result, error) {
	cr.Status.Phase = grafanav1alpha1.PhaseReconciling
	cr.Status.Message = "success"

	log.V(1).Info("ReconcileGrafana success")
	// Only update the status if the dashboard controller had a chance to sync the cluster
	// dashboards first. Otherwise reuse the existing dashboard config from the CR.
	if r.Config.GetConfigBool(config.ConfigGrafanaDashboardsSynced, false) {
		cr.Status.InstalledDashboards = r.Config.GetDashboards("")
	}

	if cr.Spec.DashboardContentCacheDuration == nil {
		cr.Spec.DashboardContentCacheDuration = &metav1.Duration{Duration: 0}
	}

	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, request.NamespacedName, instance)
	if err != nil {
		return r.manageError(cr, err, request)
	}

	if !reflect.DeepEqual(cr.Status, instance.Status) {
		instance.Status = cr.Status
		err := r.Client.Status().Update(r.Context, instance)
		if err != nil {
			return r.manageError(cr, err, request)
		}
	}
	// Make the Grafana API URL available to the dashboard controller
	url, err := r.getGrafanaAdminUrl(cr, state)
	if err != nil {
		return r.manageError(cr, err, request)
	}

	// Publish controller state
	controllerState := common.ControllerState{
		DashboardSelectors:            cr.Spec.DashboardLabelSelector,
		DashboardNamespaceSelector:    cr.Spec.DashboardNamespaceSelector,
		DashboardContentCacheDuration: cr.Spec.DashboardContentCacheDuration,
		AdminUrl:                      url,
		GrafanaReady:                  true,
		ClientTimeout:                 DefaultClientTimeoutSeconds,
	}

	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		seconds := *cr.Spec.Client.TimeoutSeconds
		if seconds < 0 {
			seconds = DefaultClientTimeoutSeconds
		}
		controllerState.ClientTimeout = seconds
	}

	common.ControllerEvents <- controllerState

	log.V(1).Info("desired cluster state met")

	return reconcile.Result{RequeueAfter: r.Config.RequeueDelay}, nil
}
