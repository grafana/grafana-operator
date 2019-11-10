package common

import (
	"context"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
	v12 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterState struct {
	GrafanaService        *v1.Service
	GrafanaServiceAccount *v1.ServiceAccount
	GrafanaConfig         *v1.ConfigMap
	GrafanaRoute          *v12.Route
	GrafanaIngress        *v1beta1.Ingress
	GrafanaDeployment     *v13.Deployment
}

func NewClusterState() *ClusterState {
	return &ClusterState{}
}

func (i *ClusterState) Read(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	cfg := config.GetControllerConfig()
	isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift, false)

	err := i.readGrafanaService(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaServiceAccount(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaConfig(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaDeployment(ctx, cr, client)
	if err != nil {
		return err
	}

	if isOpenshift {
		err = i.readGrafanaRoute(ctx, cr, client)
	} else {
		err = i.readGrafanaIngress(ctx, cr, client)
	}

	return err
}

func (i *ClusterState) readGrafanaService(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaService(cr)
	selector := model.GrafanaServiceSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			i.GrafanaService = nil
			return nil
		}
		return err
	}

	i.GrafanaService = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaServiceAccount(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaServiceAccount(cr)
	selector := model.GrafanaServiceAccountSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		i.GrafanaServiceAccount = currentState.DeepCopy()
	}
	return nil
}

func (i *ClusterState) readGrafanaConfig(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState, err := model.GrafanaConfig(cr)
	if err != nil {
		return err
	}
	selector := model.GrafanaConfigSelector(cr)
	err = client.Get(ctx, selector, currentState)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		i.GrafanaConfig = currentState.DeepCopy()
	}
	return nil
}

func (i *ClusterState) readGrafanaRoute(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaRoute(cr)
	selector := model.GrafanaRouteSelector(cr)

	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		i.GrafanaRoute = currentState.DeepCopy()
	}
	return nil
}

func (i *ClusterState) readGrafanaIngress(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaIngress(cr)
	selector := model.GrafanaIngressSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		i.GrafanaIngress = currentState.DeepCopy()
	}
	return nil
}

func (i *ClusterState) readGrafanaDeployment(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaDeployment(cr)
	selector := model.GrafanaDeploymentSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		i.GrafanaDeployment = currentState.DeepCopy()
	}
	return nil
}
