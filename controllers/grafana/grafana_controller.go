/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grafana

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/api/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/controllers/common"
	"github.com/integr8ly/grafana-operator/v3/controllers/config"
	"github.com/integr8ly/grafana-operator/v3/controllers/model"
)

const ControllerName = "grafana-controller"
const DefaultClientTimeoutSeconds = 5

var log = logf.Log.WithName(ControllerName)

// GrafanaReconciler reconciles a Grafana object
type GrafanaReconciler struct {
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

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
// +kubebuilder:rbac:groups=integreatly.integreatly.org,resources=grafanas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.integreatly.org,resources=grafanas/status,verbs=get;update;patch
func (r *GrafanaReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("grafana", req.NamespacedName)

	instance := &grafanav1alpha1.Grafana{}
	err := r.client.Get(r.context, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Stop the dashboard controller from reconciling when grafana is not installed
			r.config.RemoveConfigItem(config.ConfigDashboardLabelSelector)
			r.config.Cleanup(true)

			common.ControllerEvents <- common.ControllerState{
				GrafanaReady: false,
			}

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
	reconciler := NewGrafanaReconcile()
	desiredState := reconciler.Reconcile(currentState, cr)

	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(r.context, r.client, r.scheme, cr)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(cr, err)
	}

	// Run the config map reconciler to discover jsonnet libraries
	err = reconcileConfigMaps(cr, r)
	if err != nil {
		return r.manageError(cr, err)
	}

	return r.manageSuccess(cr, currentState)
}

func (r *ReconcileGrafana) manageError(cr *grafanav1alpha1.Grafana, issue error) (reconcile.Result, error) {
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

	r.config.InvalidateDashboards()

	common.ControllerEvents <- common.ControllerState{
		GrafanaReady: false,
	}

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

// Try to find a suitable url to grafana
func (r *ReconcileGrafana) getGrafanaAdminUrl(cr *grafanav1alpha1.Grafana, state *common.ClusterState) (string, error) {
	// If preferService is true, we skip the routes and try to access grafana
	// by using the service.
	preferService := false
	if cr.Spec.Client != nil {
		preferService = cr.Spec.Client.PreferService
	}

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
		if r.config.Dashboards == nil {
			r.config.SetDashboards(make(map[string][]*grafanav1alpha1.GrafanaDashboardRef))
		}
	}

	if state.AdminSecret == nil || state.AdminSecret.Data == nil {
		return r.manageError(cr, stdErr.New("admin secret not found or invalid"))
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

	// Publish controller state
	controllerState := common.ControllerState{
		DashboardSelectors:         cr.Spec.DashboardLabelSelector,
		DashboardNamespaceSelector: cr.Spec.DashboardNamespaceSelector,
		AdminUsername:              string(state.AdminSecret.Data[model.GrafanaAdminUserEnvVar]),
		AdminPassword:              string(state.AdminSecret.Data[model.GrafanaAdminPasswordEnvVar]),
		AdminUrl:                   url,
		GrafanaReady:               true,
		ClientTimeout:              DefaultClientTimeoutSeconds,
	}

	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		seconds := DefaultClientTimeoutSeconds
		if seconds < 0 {
			seconds = DefaultClientTimeoutSeconds
		}
		controllerState.ClientTimeout = seconds
	}

	common.ControllerEvents <- controllerState

	log.Info("desired cluster state met")

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

func (r *GrafanaReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

	return ctrl.NewControllerManagedBy(mgr).
		For(&integreatlyv1alpha1.Grafana{}).
		Owns(&v12.Deployment{}).
		Owns(&v1beta12.Ingress{}).
		Owns(&v1.ConfigMap{}).
		Owns(&v1.Service{}).
		Owns(&v1.ServiceAccount{}).
		Complete(r)
}
