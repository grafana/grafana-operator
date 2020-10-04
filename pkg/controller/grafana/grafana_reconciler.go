package grafana

import (
	"fmt"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
)

type GrafanaReconciler struct {
	DsHash     string
	ConfigHash string
	PluginsEnv string
	Plugins    *PluginsHelperImpl
}

func NewGrafanaReconciler() *GrafanaReconciler {
	return &GrafanaReconciler{
		DsHash:     "",
		ConfigHash: "",
		PluginsEnv: "",
		Plugins:    newPluginsHelper(),
	}
}

func (i *GrafanaReconciler) Reconcile(state *common.ClusterState, cr *v1alpha1.Grafana) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddAction(i.getGrafanaAdminUserSecretDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaServiceDesiredState(state, cr))

	if cr.UsedPersistentVolume() {
		desired = desired.AddAction(i.getGrafanaDataPvcDesiredState(state, cr))
	}

	desired = desired.AddAction(i.getGrafanaServiceAccountDesiredState(state, cr))
	desired = desired.AddActions(i.getGrafanaConfigDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaDatasourceConfigDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaExternalAccessDesiredState(state, cr))

	// Consolidate plugins
	// No action, will update init container env var
	desired = desired.AddAction(i.getGrafanaPluginsDesiredState(cr))

	// Reconcile the deployment last because it depends on the configuration
	// and plugins list computed in previous steps
	desired = desired.AddAction(i.getGrafanaDeploymentDesiredState(state, cr))
	desired = desired.AddActions(i.getEnvVarsDesiredState(state, cr))

	// Check Deployment and Route readiness
	desired = desired.AddActions(i.getGrafanaReadiness(state, cr))

	return desired
}

func (i *GrafanaReconciler) getGrafanaReadiness(state *common.ClusterState, cr *v1alpha1.Grafana) []common.ClusterAction {
	var actions []common.ClusterAction
	cfg := config.GetControllerConfig()
	openshift := cfg.GetConfigBool(config.ConfigOpenshift, false)
	if openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled && (cr.Spec.Client == nil || !cr.Spec.Client.PreferService) {
		// On OpenShift, check the route, only if preferService is false
		actions = append(actions, common.RouteReadyAction{
			Ref: state.GrafanaRoute,
			Msg: "check route readiness",
		})
	}
	if !openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled && (cr.Spec.Client == nil || !cr.Spec.Client.PreferService) {
		// On vanilla Kubernetes, check the ingress,only if preferService is false
		actions = append(actions, common.IngressReadyAction{
			Ref: state.GrafanaIngress,
			Msg: "check ingress readiness",
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

func (i *GrafanaReconciler) getGrafanaDataPvcDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaDataPersistentVolumeClaim == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaDataPVC(cr),
			Msg: "create grafana data persistentVolumeClaim",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaPVCReconciled(cr, state.GrafanaDataPersistentVolumeClaim),
		Msg: "update grafana data persistentVolumeClaim",
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

func (i *GrafanaReconciler) getGrafanaConfigDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) []common.ClusterAction {
	actions := []common.ClusterAction{}

	if state.GrafanaConfig == nil {
		config, err := model.GrafanaConfig(cr)
		if err != nil {
			log.Error(err, "error creating grafana config")
			return nil
		}

		// Store the last config hash for the duration of this reconciliation for
		// later usage in the deployment
		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		actions = append(actions, common.GenericCreateAction{
			Ref: config,
			Msg: "create grafana config",
		})
	} else {
		config, err := model.GrafanaConfigReconciled(cr, state.GrafanaConfig)
		if err != nil {
			log.Error(err, "error updating grafana config")
			return nil
		}

		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		actions = append(actions, common.GenericUpdateAction{
			Ref: config,
			Msg: "update grafana config",
		})
	}
	return actions
}

func (i *GrafanaReconciler) getGrafanaDatasourceConfigDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	// Only create the datasources configmap if it doesn't exist. Updates
	// are handled by the datasources controller
	if state.GrafanaDataSourceConfig == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaDatasourcesConfig(cr),
			Msg: "create grafanadatasource config",
		}
	} else {
		if state.GrafanaDataSourceConfig.Annotations != nil {
			i.DsHash = state.GrafanaDataSourceConfig.Annotations[model.LastConfigAnnotation]
		}
	}
	return nil
}

func (i *GrafanaReconciler) getGrafanaExternalAccessDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	cfg := config.GetControllerConfig()
	isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift, false)

	if cr.Spec.Ingress == nil || !cr.Spec.Ingress.Enabled {
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

func (i *GrafanaReconciler) getGrafanaAdminUserSecretDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SkipCreateAdminAccount != nil && *cr.Spec.Deployment.SkipCreateAdminAccount {
		return nil
	}

	if state.AdminSecret == nil {
		return common.GenericCreateAction{
			Ref: model.AdminSecret(cr),
			Msg: "create admin credentials secret",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.AdminSecretReconciled(cr, state.AdminSecret),
		Msg: "update admin credentials secret",
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
			Ref: model.GrafanaDeployment(cr, i.ConfigHash, i.DsHash),
			Msg: "create grafana deployment",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaDeploymentReconciled(cr, state.GrafanaDeployment,
			i.ConfigHash, i.PluginsEnv, i.DsHash),
		Msg: "update grafana deployment",
	}
}

func (i *GrafanaReconciler) getEnvVarsDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) []common.ClusterAction {
	if state.GrafanaDeployment == nil {
		return nil
	}

	// Don't look for external admin credentials if the operator created an account
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SkipCreateAdminAccount != nil && *cr.Spec.Deployment.SkipCreateAdminAccount == false {
		return nil
	}

	var actions []common.ClusterAction

	for _, container := range state.GrafanaDeployment.Spec.Template.Spec.Containers {
		if container.Name != "grafana" {
			continue
		}
		for _, source := range container.EnvFrom {
			if source.ConfigMapRef != nil && source.ConfigMapRef.Name != "" {
				actions = append(actions, common.ExposeConfigMapEnvVarAction{
					Ref:       source.ConfigMapRef,
					Namespace: cr.Namespace,
					Msg:       fmt.Sprintf("looking for admin credentials in config map %s", source.ConfigMapRef.Name),
				})
			} else if source.SecretRef != nil && source.SecretRef.Name != "" {
				actions = append(actions, common.ExposeSecretEnvVarAction{
					Ref:       source.SecretRef,
					Namespace: cr.Namespace,
					Msg:       fmt.Sprintf("looking for admin credentials in secret %s", source.SecretRef.Name),
				})
			}
		}
	}

	return actions
}

func (i *GrafanaReconciler) getGrafanaPluginsDesiredState(cr *v1alpha1.Grafana) common.ClusterAction {
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
