package grafanaproxy

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
)

type GrafanaProxyReconciler struct {
	DsHash     string
	ConfigHash string
	PluginsEnv string
}

func NewGrafanaProxyReconciler() *GrafanaProxyReconciler {
	return &GrafanaProxyReconciler{
		DsHash:     "",
		ConfigHash: "",
		PluginsEnv: "",
	}
}

func (i *GrafanaProxyReconciler) Reconcile(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddActions(i.getGrafanaProxyConfigDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaProxyServiceAccountDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaProxyServiceDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaProxyIngressDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaProxyDeploymentDesiredState(state, cr))

	// Check Deployment and Route readiness
	desired = desired.AddActions(i.getGrafanaProxyReadiness(state, cr))

	return desired
}

func (i *GrafanaProxyReconciler) getGrafanaProxyReadiness(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) []common.ClusterAction {
	var actions []common.ClusterAction
	// On vanilla Kubernetes, check the ingress
	actions = append(actions, common.IngressReadyAction{
		Ref: state.GrafanaProxyIngress,
		Msg: "check proxy ingress readiness",
	})

	return append(actions, common.DeploymentReadyAction{
		Ref: state.GrafanaProxyDeployment,
		Msg: "check proxy deployment readiness",
	})
}

func (i *GrafanaProxyReconciler) getGrafanaProxyServiceDesiredState(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) common.ClusterAction {
	if state.GrafanaProxyService == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaProxyService(cr),
			Msg: "create grafana proxy service",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaProxyServiceReconciled(cr, state.GrafanaProxyService),
		Msg: "update grafana proxy service",
	}
}

func (i *GrafanaProxyReconciler) getGrafanaProxyServiceAccountDesiredState(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) common.ClusterAction {
	if state.GrafanaProxyServiceAccount == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaProxyServiceAccount(cr),
			Msg: "create grafana proxy service account",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaProxyServiceAccountReconciled(cr, state.GrafanaProxyServiceAccount),
		Msg: "update grafana proxy service account",
	}
}

func (i *GrafanaProxyReconciler) getGrafanaProxyConfigDesiredState(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) []common.ClusterAction {
	actions := []common.ClusterAction{}

	if state.GrafanaProxyConfig == nil {
		config, err := model.GrafanaProxyConfig(cr)
		if err != nil {
			log.Error(err, "error creating grafana proxy config")
			return nil
		}

		// Store the last config hash for the duration of this reconciliation for
		// later usage in the deployment
		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		actions = append(actions, common.GenericCreateAction{
			Ref: config,
			Msg: "create grafana proxy config",
		})
	} else {
		config, err := model.GrafanaProxyConfigReconciled(cr, state.GrafanaProxyConfig)
		if err != nil {
			log.Error(err, "error updating grafana proxy config")
			return nil
		}

		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		actions = append(actions, common.GenericUpdateAction{
			Ref: config,
			Msg: "update grafana proxy config",
		})
	}
	return actions
}

func (i *GrafanaProxyReconciler) getGrafanaProxyIngressDesiredState(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) common.ClusterAction {
	if state.GrafanaProxyIngress == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaProxyIngress(cr),
			Msg: "create grafana proxy ingress",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.GrafanaProxyIngressReconciled(cr, state.GrafanaProxyIngress),
		Msg: "update grafana proxy ingress",
	}
}

func (i *GrafanaProxyReconciler) getGrafanaProxyDeploymentDesiredState(state *common.ClusterState, cr *v1alpha1.GrafanaProxy) common.ClusterAction {
	if state.GrafanaProxyDeployment == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaProxyDeployment(cr, i.ConfigHash),
			Msg: "create grafana proxy deployment",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaProxyDeploymentReconciled(cr, state.GrafanaProxyDeployment,
			i.ConfigHash),
		Msg: "update grafana proxy deployment",
	}
}
