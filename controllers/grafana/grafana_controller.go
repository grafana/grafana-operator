/*
Copyright 2021.

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
	stdErr "errors"
	"fmt"
	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/api/integreatly/v1alpha1"
	integreatlyorgv1alpha1 "github.com/integr8ly/grafana-operator/api/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/controllers/common"
	"github.com/integr8ly/grafana-operator/controllers/config"
	"github.com/integr8ly/grafana-operator/controllers/model"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ControllerName = "grafana-controller"
const DefaultClientTimeoutSeconds = 5

// GrafanaReconciler reconciles a Grafana object
type GrafanaReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Plugins  *PluginsHelperImpl
	Log      logr.Logger
	Context  context.Context
	Cancel   context.CancelFunc
	Config   *config.ControllerConfig
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=integreatly.org.integreatly.org,resources=grafanas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=integreatly.org.integreatly.org,resources=grafanas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=integreatly.org.integreatly.org,resources=grafanas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Grafana object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GrafanaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Stop the dashboard controller from reconciling when grafana is not installed
			r.Config.RemoveConfigItem(config.ConfigDashboardLabelSelector)
			r.Config.Cleanup(true)

			common.ControllerEvents <- common.ControllerState{
				GrafanaReady: false,
			}

			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cr := instance.DeepCopy()

	// Read current state
	currentState := common.NewClusterState()
	err = currentState.Read(r.Context, cr, r.Client)
	if err != nil {
		log.Log.Error(err, "error reading state")
		return r.manageError(cr, err, req)
	}

	// Get the actions required to reach the desired state
<<<<<<< HEAD:pkg/controller/grafana/grafana_controller.go

	reconciler := NewGrafanaReconciler()
	desiredState := reconciler.Reconcile(currentState, cr)
=======
	grafanaState := NewGrafanaState()
<<<<<<< HEAD
	desiredState := grafanaState.getGrafanaDesiredState(currentState, cr)
>>>>>>> 3cb2e6f6... restructure project and begin moving grafana_controller:controllers/grafana/grafana_controller.go
=======
	actions := grafanaState.getReconcileActions(currentState, cr)
>>>>>>> 625e8f51... add grafana controller/reconciler classes and misc files

	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(r.Context, r.Client, r.Scheme, cr)
	err = actionRunner.RunAll(actions)
	if err != nil {
		return r.manageError(cr, err, req)
	}

	// Run the config map grafanaState to discover jsonnet libraries
	err = reconcileConfigMaps(cr, r)
	if err != nil {
		return r.manageError(cr, err, req)
	}

	return r.manageSuccess(cr, currentState, req)
}

func (r *GrafanaReconciler) manageError(cr *grafanav1alpha1.Grafana, issue error, request reconcile.Request) (ctrl.Result, error) {
	r.Recorder.Event(cr, "Warning", "ProcessingError", issue.Error())
	cr.Status.Phase = grafanav1alpha1.PhaseFailing
	cr.Status.Message = issue.Error()

	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, request.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(cr.Status, instance.Status) {
		err := r.Client.Status().Update(r.Context, cr)
		if err != nil {
			// Ignore conflicts, resource might just be outdated.
			if errors.IsConflict(err) {
				err = nil
			}
			return ctrl.Result{}, err
		}
	}

	r.Config.InvalidateDashboards()

	common.ControllerEvents <- common.ControllerState{
		GrafanaReady: false,
	}

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

// Try to find a suitable url to grafana
func (r *GrafanaReconciler) getGrafanaAdminUrl(cr *grafanav1alpha1.Grafana, state *common.ClusterState) (string, error) {
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

	var servicePort = model.GetGrafanaPort(cr)

	// Otherwise rely on the service
	if state.GrafanaService != nil {
		return fmt.Sprintf("http://%v:%d", state.GrafanaService.Name,
			servicePort), nil
	}

	return "", stdErr.New("failed to find admin url")
}

func (r *GrafanaReconciler) manageSuccess(cr *grafanav1alpha1.Grafana, state *common.ClusterState, request reconcile.Request) (reconcile.Result, error) {
	cr.Status.Phase = grafanav1alpha1.PhaseReconciling
	cr.Status.Message = "success"

	// Only update the status if the dashboard controller had a chance to sync the cluster
	// dashboards first. Otherwise reuse the existing dashboard config from the CR.
	if r.Config.GetConfigBool(config.ConfigGrafanaDashboardsSynced, false) {

		cr.Status.InstalledDashboards = r.Config.Dashboards
	} else {
		if r.Config.Dashboards == nil {
			r.Config.SetDashboards(make(map[string][]*grafanav1alpha1.GrafanaDashboardRef))
		}
	}

	instance := &grafanav1alpha1.Grafana{}
	err := r.Client.Get(r.Context, request.NamespacedName, instance)
	if err != nil {
		return r.manageError(cr, err, request)
	}

	if !reflect.DeepEqual(cr.Status, instance.Status) {
		err := r.Client.Status().Update(r.Context, cr)
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
		DashboardSelectors:         cr.Spec.DashboardLabelSelector,
		DashboardNamespaceSelector: cr.Spec.DashboardNamespaceSelector,
		AdminUrl:                   url,
		GrafanaReady:               true,
		ClientTimeout:              DefaultClientTimeoutSeconds,
	}

	if cr.Spec.Client != nil && cr.Spec.Client.TimeoutSeconds != nil {
		seconds := *cr.Spec.Client.TimeoutSeconds
		if seconds < 0 {
			seconds = DefaultClientTimeoutSeconds
		}
		controllerState.ClientTimeout = seconds
	}

	common.ControllerEvents <- controllerState

	log.Log.V(1).Info("desired cluster state met")

	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&integreatlyorgv1alpha1.Grafana{}).
		Complete(r)
}
