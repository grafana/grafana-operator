package loki

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
)

type LokiReconciler struct {
	ConfigHash string
	status     string
}

func NewLokiReconciler() *LokiReconciler {
	return &LokiReconciler{
		ConfigHash: "",
		status:     "",
	}
}

func (i *LokiReconciler) Reconcile(state *common.LokiState, cr *v1alpha1.Loki) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddAction(i.getLokiServiceDesiredState(state, cr))

	desired = desired.AddAction(i.getLokiServiceAccountDesiredState(state, cr))
	desired = desired.AddActions(i.getLokiConfigDesiredState(state, cr))

	// Reconcile the deployment last because it depends on the configuration
	// and plugins list computed in previous steps
	desired = desired.AddAction(i.getLokiDeploymentDesiredState(state, cr))

	// Check Deployment and Route readiness
	desired = desired.AddActions(i.getLokiReadiness(state, cr))

	return desired
}

func (i *LokiReconciler) getLokiReadiness(state *common.LokiState, cr *v1alpha1.Loki) []common.ClusterAction {
	var actions []common.ClusterAction
	cfg := config.GetControllerConfig()
	openshift := cfg.GetConfigBool(config.ConfigOpenshift, false)
	if openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled {
		// On OpenShift, check the route, only if preferService is false
		actions = append(actions, common.RouteReadyAction{
			Ref: state.LokiRoute,
			Msg: "check route readiness",
		})
	}
	if !openshift && cr.Spec.Ingress != nil && cr.Spec.Ingress.Enabled {
		// On vanilla Kubernetes, check the ingress,only if preferService is false
		actions = append(actions, common.IngressReadyAction{
			Ref: state.LokiIngress,
			Msg: "check ingress readiness",
		})
	}
	return append(actions, common.DeploymentReadyAction{
		Ref: state.LokiDeployment,
		Msg: "check deployment readiness",
	})
}

func (i *LokiReconciler) getLokiServiceDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	if state.LokiService == nil {
		return common.GenericCreateAction{
			Ref: model.LokiService(cr),
			Msg: "create loki service",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.LokiServiceReconciled(cr, state.LokiService),
		Msg: "update loki service",
	}
}

func (i *LokiReconciler) getLokiServiceAccountDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	if state.LokiServiceAccount == nil {
		return common.GenericCreateAction{
			Ref: model.LokiServiceAccount(cr),
			Msg: "create grafana service account",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.LokiServiceAccountReconciled(cr, state.LokiServiceAccount),
		Msg: "update grafana service account",
	}
}

func (i *LokiReconciler) getLokiConfigDesiredState(state *common.LokiState, cr *v1alpha1.Loki) []common.ClusterAction {
	actions := []common.ClusterAction{}

	if state.LokiConfigMap == nil {
		config, err := model.LokiConfig(cr)
		if err != nil {
			log.Error(err, "error creating grafana config")
			return nil
		}

		// Store the last config hash for the duration of this reconciliation for
		// later usage in the deployment
		i.ConfigHash = config.Annotations[model.LastConfigAnnotation]

		actions = append(actions, common.GenericCreateAction{
			Ref: config,
			Msg: "create loki config",
		})
	} else {
		config, err := model.LokiConfigReconciled(cr, state.LokiConfigMap)
		if err != nil {
			log.Error(err, "error updating loki config")
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

func (i *LokiReconciler) getLokiExternalAccessDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	cfg := config.GetControllerConfig()
	isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift, false)

	if cr.Spec.Ingress == nil || !cr.Spec.Ingress.Enabled {
		// external access not enabled: remote the route/ingress if it exists or
		// do nothing
		if isOpenshift && state.LokiRoute != nil {
			return common.GenericDeleteAction{
				Ref: state.LokiRoute,
				Msg: "delete loki route",
			}
		} else if !isOpenshift && state.LokiIngress != nil {
			return common.GenericDeleteAction{
				Ref: state.LokiIngress,
				Msg: "delete loki ingress",
			}
		}
		return nil
	} else {
		// external access enabled: create route/ingress
		if isOpenshift {
			return i.getLokiRouteDesiredState(state, cr)
		}
		return i.getLokiRouteDesiredState(state, cr)
	}
}

func (i *LokiReconciler) getGrafanaAdminUserSecretDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
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

func (i *LokiReconciler) getLokiIngressDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	if state.LokiIngress == nil {
		return common.GenericCreateAction{
			Ref: model.LokiIngress(cr),
			Msg: "create grafana ingress",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.LokiIngressReconciled(cr, state.LokiIngress),
		Msg: "update grafana ingress",
	}
}

func (i *LokiReconciler) getLokiRouteDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	if state.LokiRoute == nil {
		return common.GenericCreateAction{
			Ref: model.LokiRoute(cr),
			Msg: "create loki route",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.LokiRouteReconciled(cr, state.LokiRoute),
		Msg: "update loki route",
	}
}

func (i *LokiReconciler) getLokiDeploymentDesiredState(state *common.LokiState, cr *v1alpha1.Loki) common.ClusterAction {
	if state.LokiDeployment == nil {
		return common.GenericCreateAction{
			Ref: model.LokiDeployment(cr, i.ConfigHash),
			Msg: "create Loki deployment",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.LokiDeploymentReconciled(cr, state.LokiDeployment,
			i.ConfigHash),
		Msg: "update Loki deployment",
	}
}
