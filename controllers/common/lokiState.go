package common

import (
	"context"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	v12 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LokiState struct {
	LokiService        *v1.Service
	LokiServiceAccount *v1.ServiceAccount
	LokiDeployment     *v13.Deployment
	LokiConfigMap      *v1.ConfigMap
	LokiRoute          *v12.Route
	LokiIngress        *v1beta1.Ingress
	LokiDataSource     *grafanav1alpha1.GrafanaDataSource
}

func NewLokiState() *LokiState {
	return &LokiState{}
}

func (i *LokiState) Read(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
	cfg := config.GetControllerConfig()
	isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift, false)

	err := i.readLokiService(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readLokiServiceAccount(ctx, cr, client)
	if err != nil {
		return err
	}

	if isOpenshift {
		err = i.readLokiRoute(ctx, cr, client)
	} else {
		err = i.readLokiIngress(ctx, cr, client)
	}
	// TODO
	//if err := i.readLokiConfig(ctx,cr,client); err != nil {
	//
	//}
	return err
}

func (i *LokiState) readLokiService(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
	currentState := &v1.Service{}
	selector := model.LokiServiceSelector(cr)
	if err := client.Get(ctx, selector, currentState); err != nil {
		if errors.IsNotFound(err) {
			return err
		}
	}
	i.LokiService = currentState.DeepCopy()
	return nil
}

func (i *LokiState) readLokiRoute(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
	currentState := &v12.Route{}
	selector := model.LokiRouteSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.LokiRoute = currentState.DeepCopy()
	return nil
}

func (i *LokiState) readLokiIngress(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
	currentState := &v1beta1.Ingress{}
	selector := model.LokiIngressSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.LokiIngress = currentState.DeepCopy()
	return nil
}

func (i *LokiState) readLokiServiceAccount(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
	currentState := &v1.ServiceAccount{}
	selector := model.LokiServiceAccountSelector(cr)
	if err := client.Get(ctx, selector, currentState); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.LokiServiceAccount = currentState.DeepCopy()
	return nil
}

// TODO
//func (i *ClusterState) readLokiConfig(ctx context.Context, cr *grafanav1alpha1.Loki, client client.Client) error {
//	currentState, err := model.LokiConfig(cr)
//	if err != nil {
//		return err
//	}
//	selector := model.LokiConfigSelector(cr)
//	err = client.Get(ctx, selector, currentState)
//	if err != nil {
//		if errors.IsNotFound(err) {
//			return nil
//		}
//		return err
//	}
//	i.GrafanaConfig = currentState.DeepCopy()
//	return nil
//}
