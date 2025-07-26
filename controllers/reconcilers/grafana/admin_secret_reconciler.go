package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type AdminSecretReconciler struct {
	client client.Client
}

func NewAdminSecretReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &AdminSecretReconciler{
		client: client,
	}
}

func (r *AdminSecretReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	if cr.Spec.DisableDefaultAdminSecret {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	secret := model.GetGrafanaAdminSecret(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, secret, func() error {
		secret.Data = getData(cr, secret)

		if scheme != nil {
			err := controllerutil.SetControllerReference(cr, secret, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(secret, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getAdminUser(cr *v1beta1.Grafana, current *v1.Secret) []byte {
	adminUser := cr.GetConfigSectionValue("security", "admin_user")
	if adminUser != "" {
		return []byte(adminUser)
	}

	// If a user is already set, don't change it
	if current != nil && current.Data != nil && current.Data[config.GrafanaAdminUserEnvVar] != nil {
		return current.Data[config.GrafanaAdminUserEnvVar]
	}

	return []byte(config.DefaultAdminUser)
}

func getAdminPassword(cr *v1beta1.Grafana, current *v1.Secret) []byte {
	adminPassword := cr.GetConfigSectionValue("security", "admin_password")
	if adminPassword != "" {
		return []byte(adminPassword)
	}

	// If a password is already set, don't change it
	if current != nil && current.Data != nil && current.Data[config.GrafanaAdminPasswordEnvVar] != nil {
		return current.Data[config.GrafanaAdminPasswordEnvVar]
	}

	return []byte(model.RandStringRunes(10))
}

func getData(cr *v1beta1.Grafana, current *v1.Secret) map[string][]byte {
	credentials := map[string][]byte{
		config.GrafanaAdminUserEnvVar:     getAdminUser(cr, current),
		config.GrafanaAdminPasswordEnvVar: getAdminPassword(cr, current),
	}

	return credentials
}
