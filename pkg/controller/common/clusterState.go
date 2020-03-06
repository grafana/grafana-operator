package common

import (
	"context"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	v12 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterState struct {
	GrafanaService             *v1.Service
	GrafanaProxyService        *v1.Service
	GrafanaServiceAccount      *v1.ServiceAccount
	GrafanaProxyServiceAccount *v1.ServiceAccount
	GrafanaConfig              *v1.ConfigMap
	GrafanaProxyConfig         *v1.ConfigMap
	GrafanaRoute               *v12.Route
	GrafanaIngress             *v1beta1.Ingress
	GrafanaProxyIngress        *v1beta1.Ingress
	GrafanaProxyRoleBinding    *rbac.RoleBinding
	GrafanaDeployment          *v13.Deployment
	GrafanaProxyDeployment     *v13.Deployment
	GrafanaDataSourceConfig    *v1.ConfigMap
	AdminSecret                *v1.Secret
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

	err = i.readGrafanaDatasourceConfig(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaDeployment(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaAdminUserSecret(ctx, cr, client)
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

func (i *ClusterState) ReadProxy(ctx context.Context, crp *v1alpha1.GrafanaProxy, client client.Client) (err error) {
	if err = i.readGrafanaProxyService(ctx, crp, client); err != nil {
		return
	}

	if err = i.readGrafanaProxyServiceAccount(ctx, crp, client); err != nil {
		return
	}

	if err = i.readGrafanaProxyConfig(ctx, crp, client); err != nil {
		return
	}

	if err = i.readGrafanaProxyDeployment(ctx, crp, client); err != nil {
		return
	}
	if err != nil {
		return
	}

	if err = i.readGrafanaProxyIngress(ctx, crp, client); err != nil {
		return err
	}

	if err = i.readGrafanaProxyRoleBinding(ctx, crp, client); err != nil {
		return err
	}

	return err
}

func (i *ClusterState) readGrafanaService(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaService(cr)
	selector := model.GrafanaServiceSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaService = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyService(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState := model.GrafanaProxyService(cr)
	selector := model.GrafanaProxyServiceSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyService = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaServiceAccount(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaServiceAccount(cr)
	selector := model.GrafanaServiceAccountSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaServiceAccount = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyServiceAccount(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState := model.GrafanaProxyServiceAccount(cr)
	selector := model.GrafanaProxyServiceAccountSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyServiceAccount = currentState.DeepCopy()
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
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaConfig = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyConfig(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState, err := model.GrafanaProxyConfig(cr)
	if err != nil {
		return err
	}
	selector := model.GrafanaProxyConfigSelector(cr)
	err = client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyConfig = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaDatasourceConfig(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaDatasourcesConfig(cr)
	selector := model.GrafanaDatasourceConfigSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaDataSourceConfig = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaRoute(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaRoute(cr)
	selector := model.GrafanaRouteSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaRoute = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaIngress(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaIngress(cr)
	selector := model.GrafanaIngressSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaIngress = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyIngress(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState := model.GrafanaProxyIngress(cr)
	selector := model.GrafanaProxyIngressSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyIngress = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaDeployment(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.GrafanaDeployment(cr, "", "")
	selector := model.GrafanaDeploymentSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaDeployment = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyDeployment(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState := model.GrafanaProxyDeployment(cr, "")
	selector := model.GrafanaProxyDeploymentSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyDeployment = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaAdminUserSecret(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	currentState := model.AdminSecret(cr)
	selector := model.AdminSecretSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.AdminSecret = currentState.DeepCopy()
	return nil
}

func (i *ClusterState) readGrafanaProxyRoleBinding(ctx context.Context, cr *v1alpha1.GrafanaProxy, client client.Client) error {
	currentState := model.GrafanaProxyRoleBinding(cr)
	selector := model.GrafanaProxyRoleBindingSelector(cr)
	err := client.Get(ctx, selector, currentState)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	i.GrafanaProxyRoleBinding = currentState.DeepCopy()
	return nil
}
