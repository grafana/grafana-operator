package grafana

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	"github.com/grafana-operator/grafana-operator/v5/controllers/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type AdminSecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *AdminSecretReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {
	secret := model.GetGrafanaAdminSecret(cr, r.Scheme) // TODO: inline model
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = getData(cr, secret)
		return nil
	})
	if err != nil {
		return nil, err // TODO: error condition
	}
	return nil, nil // todo: success condition
}

func getAdminUser(cr *v1beta1.Grafana, current *v1.Secret) []byte {
	if cr.Spec.Config["security"] == nil || cr.Spec.Config["security"]["admin_user"] == "" {
		// If a user is already set, don't change it
		if current != nil && current.Data[config.GrafanaAdminUserEnvVar] != nil {
			return current.Data[config.GrafanaAdminUserEnvVar]
		}
		return []byte(config.DefaultAdminUser)
	}
	return []byte(cr.Spec.Config["security"]["admin_user"])
}

func getAdminPassword(cr *v1beta1.Grafana, current *v1.Secret) []byte {
	if cr.Spec.Config["security"] == nil || cr.Spec.Config["security"]["admin_password"] == "" {
		// If a password is already set, don't change it
		if current != nil && current.Data[config.GrafanaAdminPasswordEnvVar] != nil {
			return current.Data[config.GrafanaAdminPasswordEnvVar]
		}
		return []byte(model.RandStringRunes(10))
	}
	return []byte(cr.Spec.Config["security"]["admin_password"])
}

func getData(cr *v1beta1.Grafana, current *v1.Secret) map[string][]byte {
	credentials := map[string][]byte{
		config.GrafanaAdminUserEnvVar:     getAdminUser(cr, current),
		config.GrafanaAdminPasswordEnvVar: getAdminPassword(cr, current),
	}

	// Make the credentials available to the environment when running the operator
	// outside of the cluster
	os.Setenv(config.GrafanaAdminUserEnvVar, string(credentials[config.GrafanaAdminUserEnvVar]))
	os.Setenv(config.GrafanaAdminPasswordEnvVar, string(credentials[config.GrafanaAdminPasswordEnvVar]))

	return credentials
}
