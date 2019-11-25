package grafana

import (
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
)

type GrafanaReconciler struct {
	ConfigHash string
	PluginsEnv string
	Plugins    *PluginsHelperImpl
}

func NewGrafanaReconciler() *GrafanaReconciler {
	return &GrafanaReconciler{
		ConfigHash: "",
		PluginsEnv: "",
		Plugins:    newPluginsHelper(),
	}
}

func (i *GrafanaReconciler) Reconcile(state *common.ClusterState, cr *v1alpha1.Grafana) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddAction(i.getGrafanaServiceDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaServiceAccountDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaConfigDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaExternalAccessDesiredState(state, cr))

	// Consolidate plugins
	// No action, will update init container env var
	desired = desired.AddAction(i.getGrafanaPluginsDesiredState(cr))

	// Reconcile the deployment last because it depends on the configuration
	// and plugins list computed in previous steps
	desired = desired.AddAction(i.getGrafanaDeploymentDesiredState(state, cr))

	// Check Deployment and Route readiness
	desired = desired.AddActions(i.getGrafanaReadiness(state))

	return desired
}

func (i *GrafanaReconciler) getGrafanaReadiness(state *common.ClusterState) []common.ClusterAction {
	var actions []common.ClusterAction
	cfg := config.GetControllerConfig()
	openshift := cfg.GetConfigBool(config.ConfigOpenshift, false)
	if openshift {
		actions = append(actions, common.RouteReadyAction{
			Ref: state.GrafanaRoute,
			Msg: "check route readiness",
		})
	}

	return append(actions, common.DeploymentReadyAction{
		Ref: state.GrafanaDeployment,
		Msg: "check deployment readiness",
	})
}

func (i *GrafanaReconciler) getGrafanaServiceDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaService == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaService(cr),
			Msg: "create grafana service",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaServiceReconciled(cr, state.GrafanaService),
		Msg: "update grafana service",
	}
}

func (i *GrafanaReconciler) getGrafanaServiceAccountDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaServiceAccount == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaServiceAccount(cr),
			Msg: "create grafana service account",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaServiceAccountReconciled(cr, state.GrafanaServiceAccount),
		Msg: "update grafana service account",
	}
}

func (i *GrafanaReconciler) getGrafanaConfigDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaConfig == nil {
		config, err := model.GrafanaConfig(cr)
		if err != nil {
			log.Error(err, "error creating grafana config")
			return nil
		}

		// Store the last config hash for the duration of this reconciliation for
		// later usage in the deployment
		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		return common.GenericCreateAction{
			Ref: config,
			Msg: "create grafana config",
		}
	} else {
		config, err := model.GrafanaConfigReconciled(cr, state.GrafanaConfig)
		if err != nil {
			log.Error(err, "error updating grafana config")
			return nil
		}

		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		return common.GenericUpdateAction{
			Ref: config,
			Msg: "update grafana config",
		}

	}
}

func (i *GrafanaReconciler) getGrafanaExternalAccessDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	cfg := config.GetControllerConfig()
	isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift, false)

	if !cr.Spec.Ingress.Enabled {
		// external access not enabled: remote the route/ingress if it exists or
		// do nothing
		if isOpenshift && state.GrafanaRoute != nil {
			return common.GenericDeleteAction{
				Ref: state.GrafanaRoute,
				Msg: "delete grafana route",
			}
		} else if !isOpenshift && state.GrafanaIngress != nil {
			return common.GenericDeleteAction{
				Ref: state.GrafanaIngress,
				Msg: "delete grafana ingress",
			}
		}
		return nil
	} else {
		// external access enabled: create route/ingress
		if isOpenshift {
			return i.getGrafanaRouteDesiredState(state, cr)
		}
		return i.getGrafanaIngressDesiredState(state, cr)
	}
}

func (i *GrafanaReconciler) getGrafanaIngressDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaIngress == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaIngress(cr),
			Msg: "create grafana ingress",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.GrafanaIngressReconciled(cr, state.GrafanaIngress),
		Msg: "update grafana ingress",
	}
}

func (i *GrafanaReconciler) getGrafanaRouteDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaRoute == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaRoute(cr),
			Msg: "create grafana route",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.GrafanaRouteReconciled(cr, state.GrafanaRoute),
		Msg: "update grafana route",
	}
}

func (i *GrafanaReconciler) getGrafanaDeploymentDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaDeployment == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaDeployment(cr),
			Msg: "create grafana deployment",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaDeploymentReconciled(cr, state.GrafanaDeployment, i.ConfigHash, i.PluginsEnv),
		Msg: "update grafana deployment",
	}
}

func (i *GrafanaReconciler) getGrafanaPluginsDesiredState(cr *v1alpha1.Grafana) common.ClusterAction {
	// Waited long enough for dashboards to be ready?
	if !i.Plugins.CanUpdatePlugins() {
		// If not, still set the plugins to their last known state
		// because otherwise this could trigger a restart if plugins
		// were installed
		i.PluginsEnv = i.Plugins.BuildEnv(cr)
		return common.LogAction{
			Msg: "waiting for dashboards",
		}
	}

	// Fetch all plugins of all dashboards
	var requestedPlugins v1alpha1.PluginList
	for _, v := range config.GetControllerConfig().Plugins {
		requestedPlugins = append(requestedPlugins, v...)
	}

	// Consolidate plugins and create the new list of plugin requirements
	// If 'updated' is false then no changes have to be applied
	filteredPlugins, updated := i.Plugins.FilterPlugins(cr, requestedPlugins)
	if updated {
		i.reconcilePlugins(cr, filteredPlugins)

		// Build the new list of plugins for the init container to consume
		i.PluginsEnv = i.Plugins.BuildEnv(cr)

		// Reset the list of known dashboards to force the dashboard controller
		// to reimport them
		cfg := config.GetControllerConfig()
		cfg.InvalidateDashboards()

		return common.LogAction{
			Msg: fmt.Sprintf("plugins updated to %s", i.PluginsEnv),
		}
	} else {
		// Rebuild the env var from the installed plugins
		i.PluginsEnv = i.Plugins.BuildEnv(cr)
		return common.LogAction{
			Msg: "plugins unchanged",
		}
	}
}

func (i *GrafanaReconciler) reconcilePlugins(cr *v1alpha1.Grafana, plugins v1alpha1.PluginList) {
	var validPlugins []v1alpha1.GrafanaPlugin
	var failedPlugins []v1alpha1.GrafanaPlugin

	for _, plugin := range plugins {
		if i.Plugins.PluginExists(plugin) == false {
			log.Info(fmt.Sprintf("invalid plugin: %s@%s", plugin.Name, plugin.Version))
			failedPlugins = append(failedPlugins, plugin)
			continue
		}

		log.Info(fmt.Sprintf("installing plugin: %s@%s", plugin.Name, plugin.Version))
		validPlugins = append(validPlugins, plugin)
	}

	cr.Status.InstalledPlugins = validPlugins
	cr.Status.FailedPlugins = failedPlugins
}
