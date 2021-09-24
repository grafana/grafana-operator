package grafana

import (
	"fmt"
	"regexp"

	"github.com/integr8ly/grafana-operator/api/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/controllers/common"
	"github.com/integr8ly/grafana-operator/controllers/config"
	"github.com/integr8ly/grafana-operator/controllers/constants"
	"github.com/integr8ly/grafana-operator/controllers/model"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Plugins:    NewPluginsHelper(),
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
	if openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled && !cr.GetPreferServiceValue() {
		// On OpenShift, check the route, only if preferService is false
		actions = append(actions, common.RouteReadyAction{
			Ref: state.GrafanaRoute,
			Msg: "check route readiness",
		})
	}
	if !openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled && !cr.GetPreferServiceValue() {
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
	if cr.Status.PreviousServiceName != "" && state.GrafanaService.Name != "" {
		// if the previously known service is not the current service then delete the previous service
		// validate the service
		if cr.Status.PreviousServiceName != state.GrafanaService.Name && i.validateServiceName(cr.Spec.Service.Name) {
			serviceName := cr.Status.PreviousServiceName
			// reset the status before next loop
			cr.Status.PreviousServiceName = ""
			return common.GenericDeleteAction{
				Ref: &v1.Service{
					ObjectMeta: v12.ObjectMeta{
						Name:      serviceName,
						Namespace: cr.Namespace,
					},
				},
				Msg: "delete obsolete grafana service",
			}
		}
	}
	if cr.Status.PreviousServiceName == "" {
		cr.Status.PreviousServiceName = state.GrafanaService.Name
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
	if cr.Spec.ServiceAccount != nil && cr.Spec.ServiceAccount.Skip != nil && *cr.Spec.ServiceAccount.Skip {
		return nil
	}
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
		config := model.GrafanaConfig(cr)

		// Store the last config hash for the duration of this reconciliation for
		// later usage in the deployment
		i.ConfigHash = config.Annotations[constants.LastConfigAnnotation]

		actions = append(actions, common.GenericCreateAction{
			Ref: config,
			Msg: "create grafana config",
		})
	} else {
		config := model.GrafanaConfigReconciled(cr, state.GrafanaConfig)

		i.ConfigHash = config.Annotations[constants.LastConfigAnnotation]

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
			i.DsHash = state.GrafanaDataSourceConfig.Annotations[constants.LastConfigAnnotation]
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
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.SkipCreateAdminAccount != nil && !*cr.Spec.Deployment.SkipCreateAdminAccount {
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
	requestedPlugins := config.GetControllerConfig().GetAllPlugins()

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
	var validPlugins []v1alpha1.GrafanaPlugin  // nolint
	var failedPlugins []v1alpha1.GrafanaPlugin // nolint

	for _, plugin := range plugins {
		if !i.Plugins.PluginExists(plugin) {
			log.V(1).Info(fmt.Sprintf("invalid plugin: %s@%s", plugin.Name, plugin.Version))
			failedPlugins = append(failedPlugins, plugin)
			continue
		}

		log.V(1).Info(fmt.Sprintf("installing plugin: %s@%s", plugin.Name, plugin.Version))
		validPlugins = append(validPlugins, plugin)
	}

	cr.Status.InstalledPlugins = validPlugins
	cr.Status.FailedPlugins = failedPlugins
}

func (i *GrafanaReconciler) validateServiceName(string string) bool {
	// a DNS-1035 label must consist of lower case alphanumeric
	//    characters or '-', start with an alphabetic character, and end with an
	//    alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for
	//    validation is '[a-z]([-a-z0-9]*[a-z0-9])?
	b, err := regexp.MatchString("[a-z]([-a-z0-9]*[a-z0-9])?", string)
	if err != nil {
		return false
	}
	return b
}
