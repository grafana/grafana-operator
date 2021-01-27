package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getAdminUser(cr *v1alpha1.Grafana, current *v12.Secret) []byte {
	if cr.Spec.Config.Security == nil || cr.Spec.Config.Security.AdminUser == "" {
		// If a user is already set, don't change it
		if current != nil && current.Data[GrafanaAdminUserEnvVar] != nil {
			return current.Data[GrafanaAdminUserEnvVar]
		}
		return []byte(DefaultAdminUser)
	}
	return []byte(cr.Spec.Config.Security.AdminUser)
}

func getAdminPassword(cr *v1alpha1.Grafana, current *v12.Secret) []byte {
	if cr.Spec.Config.Security == nil || cr.Spec.Config.Security.AdminPassword == "" {
		// If a password is already set, don't change it
		if current != nil && current.Data[GrafanaAdminPasswordEnvVar] != nil {
			return current.Data[GrafanaAdminPasswordEnvVar]
		}
		return []byte(RandStringRunes(10))
	}
	return []byte(cr.Spec.Config.Security.AdminPassword)
}

func getData(cr *v1alpha1.Grafana, current *v12.Secret) map[string][]byte {
	credentials := map[string][]byte{
		GrafanaAdminUserEnvVar:     getAdminUser(cr, current),
		GrafanaAdminPasswordEnvVar: getAdminPassword(cr, current),
	}

	// Make the credentials available to the environment when running the operator
	// outside of the cluster
	os.Setenv(GrafanaAdminUserEnvVar, string(credentials[GrafanaAdminUserEnvVar]))
	os.Setenv(GrafanaAdminPasswordEnvVar, string(credentials[GrafanaAdminPasswordEnvVar]))

	return credentials
}

func AdminSecret(cr *v1alpha1.Grafana) *v12.Secret {
	return &v12.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      GrafanaAdminSecretName,
			Namespace: cr.Namespace,
		},
		Data: getData(cr, nil),
		Type: v12.SecretTypeOpaque,
	}
}

func AdminSecretReconciled(cr *v1alpha1.Grafana, currentState *v12.Secret) *v12.Secret {
	reconciled := currentState.DeepCopy()
	reconciled.Data = getData(cr, currentState)
	return reconciled
}

func AdminSecretSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaAdminSecretName,
	}
}
