package grafana

import (
	"context"
	stdErr "errors"
	"fmt"
	"time"

	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/grafanadashboard"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	routev1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
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

const ControllerName = "grafana-controller"
const DefaultClientTimeoutSeconds = 10

var log = logf.Log.WithName(ControllerName)

// Add creates a new Grafana Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, autodetectChannel chan schema.GroupVersionKind, _ string) error {
	return add(mgr, newReconciler(mgr), autodetectChannel)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileGrafana{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		plugins:  newPluginsHelper(),
		context:  ctx,
		cancel:   cancel,
		config:   config.GetControllerConfig(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
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

	if err = watchSecondaryResource(c, &v1beta12.Ingress{}); err != nil {
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
			if cfg.GetConfigBool(config.ConfigRouteWatch, false) == true {
				return
			}

			// Watch routes if they exist on the cluster
			if gvk.String() == routev1.SchemeGroupVersion.WithKind(common.RouteKind).String() {
				if err = watchSecondaryResource(c, &routev1.Route{}); err != nil {
					log.Error(err, fmt.Sprintf("error adding secondary watch for %v", common.RouteKind))
				} else {
					cfg.AddConfigItem(config.ConfigRouteWatch, true)
					log.Info(fmt.Sprintf("added secondary watch for %v", common.RouteKind))
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
	client   client.Client
	scheme   *runtime.Scheme
	plugins  *PluginsHelperImpl
	context  context.Context
	cancel   context.CancelFunc
	config   *config.ControllerConfig
	recorder record.EventRecorder
}

func watchSecondaryResource(c controller.Controller, resource runtime.Object) error {
	return c.Watch(&source.Kind{Type: resource}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &grafanav1alpha1.Grafana{},
	})
}

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
func (r *ReconcileGrafana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if request.Namespace == "grafana-operator" {
		return reconcile.Result{Requeue: false}, nil
	}

	instance := &grafanav1alpha1.Grafana{}
	err := r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Stop the dashboard controller from reconciling when grafana is not installed
			r.config.RemoveConfigItem(config.ConfigDashboardLabelSelector)
			r.config.Cleanup(true)
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	cr := instance.DeepCopy()

	// Read current state
	currentState := common.NewClusterState()
	err = currentState.Read(r.context, cr, r.client)
	if err != nil {
		log.Error(err, "error reading state")
		return r.manageError(cr, err)
	}

	// Get the actions required to reach the desired state
	reconciler := NewGrafanaReconciler()
	desiredState := reconciler.Reconcile(currentState, cr)

	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(r.context, r.client, r.scheme, cr)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(cr, err)
	}

	return r.manageSuccess(cr, currentState)
}

func (r *ReconcileGrafana) manageError(cr *grafanav1alpha1.Grafana, issue error) (reconcile.Result, error) {
	r.recorder.Event(cr, "Warning", "ProcessingError", issue.Error())
	log.Error(issue, "error creating grafana")
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

	r.config.InvalidateDashboards()

	return reconcile.Result{RequeueAfter: config.RequeueDelayOnError}, nil
}

// Try to find a suitable url to grafana
func (r *ReconcileGrafana) getGrafanaAdminUrl(cr *grafanav1alpha1.Grafana, state *common.ClusterState) (string, error) {
	// If preferService is true, we skip the routes and try to access grafana
	// by using the serivce.
	preferService := false
	if cr.Spec.Client != nil {
		preferService = cr.Spec.Client.PreferService
	}

	if cr.Spec.Config.Server.RootUrl != "" {
		return cr.Spec.Config.Server.RootUrl, nil
	}
	// First try to use the route if it exists. Prefer the route because it also works
	// when running the operator outside of the cluster
	if state.GrafanaRoute != nil && !preferService {
		return fmt.Sprintf("https://%v", state.GrafanaRoute.Spec.Host), nil
	}

	// Try the ingress first if on vanilla Kubernetes
	if state.GrafanaIngress != nil && !preferService {
		for _, ingress := range state.GrafanaIngress.Status.LoadBalancer.Ingress {
			if ingress.Hostname != "" {
				return fmt.Sprintf("https://%v", ingress.Hostname), nil
			}
			return fmt.Sprintf("https://%v", ingress.IP), nil
		}
	}

	var servicePort = int32(model.GetGrafanaPort(cr))

	// Otherwise rely on the service
	if state.GrafanaService != nil && state.GrafanaService.Spec.ClusterIP != "" {
		return fmt.Sprintf("http://%v:%d", state.GrafanaService.Spec.ClusterIP,
			servicePort), nil
	} else if state.GrafanaService != nil {
		return fmt.Sprintf("http://%v:%d", state.GrafanaService.Name,
			servicePort), nil
	}

	return "", stdErr.New("failed to find admin url")
}

func (r *ReconcileGrafana) manageSuccess(cr *grafanav1alpha1.Grafana, state *common.ClusterState) (reconcile.Result, error) {
	cr.Status.Phase = grafanav1alpha1.PhaseReconciling
	cr.Status.Message = "success"

	// Only update the status if the dashboard controller had a chance to sync the cluster
	// dashboards first. Otherwise reuse the existing dashboard config from the CR.
	if r.config.GetConfigBool(config.ConfigGrafanaDashboardsSynced, false) {
		cr.Status.InstalledDashboards = r.config.Dashboards
	} else {
		r.config.SetDashboards(cr.Status.InstalledDashboards)
		if r.config.Dashboards == nil {
			r.config.SetDashboards(make(map[string][]*grafanav1alpha1.GrafanaDashboardRef))
		}
	}

	if state.AdminSecret == nil || state.AdminSecret.Data == nil {
		return r.manageError(cr, stdErr.New("admin secret not found or invalud"))
	}

	err := r.client.Status().Update(r.context, cr)
	if err != nil {
		return r.manageError(cr, err)
	}

	// Make the Grafana API URL available to the dashboard controller
	url, err := r.getGrafanaAdminUrl(cr, state)
	if err != nil {
		return r.manageError(cr, err)
	}

	// Try to fix annotations on older dashboards?
	fixAnnotations := false
	if cr.Spec.Compat != nil && cr.Spec.Compat.FixAnnotations {
		fixAnnotations = true
	}

	// Try to fix heights that are in the wrong format?
	fixHeights := false
	if cr.Spec.Compat != nil && cr.Spec.Compat.FixHeights {
		fixHeights = true
	}

	if err = r.addDefaultDashboards(cr); err != nil {
		return r.manageError(cr, err)
	}

	// Publish controller state
	controllerState := common.ControllerState{
		DashboardSelectors: cr.Spec.DashboardLabelSelector,
		AdminUsername:      string(state.AdminSecret.Data[model.GrafanaAdminUserEnvVar]),
		AdminPassword:      string(state.AdminSecret.Data[model.GrafanaAdminPasswordEnvVar]),
		AdminUrl:           url,
		GrafanaReady:       true,
		ClientTimeout:      DefaultClientTimeoutSeconds,
		FixAnnotations:     fixAnnotations,
		FixHeights:         fixHeights,
	}

	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		seconds := DefaultClientTimeoutSeconds
		if seconds < 0 {
			seconds = DefaultClientTimeoutSeconds
		}
		controllerState.ClientTimeout = seconds
	}

	log.Info("desired cluster state met")
	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

func (r *ReconcileGrafana) addDefaultDashboards(g *grafanav1alpha1.Grafana) (err error) {
	crl := &grafanav1alpha1.GrafanaDashboardList{}
	if err = r.client.List(r.context, crl, client.InNamespace("grafana-operator")); err != nil {
		return
	}

	if len(crl.Items) == 0 {
		return stdErr.New("no default dashboards found")
	}
	client, err := r.getClient(g)
	if err != nil {
		return
	}
	for _, d := range crl.Items {
		pipeline := grafanadashboard.NewDashboardPipeline(&d, g.Spec.Compat.FixAnnotations, g.Spec.Compat.FixHeights)
		knownHash, ok := g.Spec.Config.Dashboards.DashboardHash[d.Name]
		if !ok {
			knownHash = ""
		}
		processed, err := pipeline.ProcessDashboard(knownHash)
		if err != nil {
			return err
		}
		if processed == nil {
			continue
		}
		_, err = client.CreateOrUpdateDashboard(processed)
		if err != nil {
			return err
		}
		if g.Spec.Config.Dashboards.DashboardHash == nil {
			g.Spec.Config.Dashboards.DashboardHash = make(map[string]string)
		}
		g.Spec.Config.Dashboards.DashboardHash[d.Name] = pipeline.NewHash()
		return r.client.Update(r.context, g)
	}

	return
}

// Get an authenticated grafana API client
func (r *ReconcileGrafana) getClient(g *grafanav1alpha1.Grafana) (grafanadashboard.GrafanaClient, error) {
	url := g.Spec.Config.Server.RootUrl
	if url == "" {
		return nil, stdErr.New("cannot get grafana admin url")
	}

	username := g.Spec.Config.Security.AdminUser
	if username == "" {
		return nil, stdErr.New("invalid credentials (username)")
	}

	password := g.Spec.Config.Security.AdminPassword
	if password == "" {
		return nil, stdErr.New("invalid credentials (password)")
	}

	duration := time.Duration(10 * time.Second)
	return grafanadashboard.NewGrafanaClient(url, username, password, duration), nil
}
