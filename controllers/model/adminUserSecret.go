package model

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getAdminUser(cr *v1alpha1.Grafana, current *v12.Secret) []byte {
	if cr.Spec.Config.Security == nil || cr.Spec.Config.Security.AdminUser == "" {
		// If a user is already set, don't change it
		if current != nil && current.Data[constants.GrafanaAdminUserEnvVar] != nil {
			return current.Data[constants.GrafanaAdminUserEnvVar]
		}
		return []byte(constants.DefaultAdminUser)
	}
	return []byte(cr.Spec.Config.Security.AdminUser)
}

func getAdminPassword(cr *v1alpha1.Grafana, current *v12.Secret) []byte {
	if cr.Spec.Config.Security == nil || cr.Spec.Config.Security.AdminPassword == "" {
		// If a password is already set, don't change it
		if current != nil && current.Data[constants.GrafanaAdminPasswordEnvVar] != nil {
			return current.Data[constants.GrafanaAdminPasswordEnvVar]
		}
		return []byte(RandStringRunes(10))
	}
	return []byte(cr.Spec.Config.Security.AdminPassword)
}

func getData(cr *v1alpha1.Grafana, current *v12.Secret) (map[string][]byte, string) {
	user := getAdminUser(cr, current)
	password := getAdminPassword(cr, current)

	credentials := map[string][]byte{
		constants.GrafanaAdminUserEnvVar:     user,
		constants.GrafanaAdminPasswordEnvVar: password,
	}

	h := sha256.New()
	h.Write(bytes.Join([][]byte{user, password}, []byte(":")))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// Make the credentials available to the environment when running the operator
	// outside of the cluster
	os.Setenv(constants.GrafanaAdminUserEnvVar, string(credentials[constants.GrafanaAdminUserEnvVar]))
	os.Setenv(constants.GrafanaAdminPasswordEnvVar, string(credentials[constants.GrafanaAdminPasswordEnvVar]))

	return credentials, hash
}

func AdminSecret(cr *v1alpha1.Grafana) *v12.Secret {
	data, hash := getData(cr, nil)

	return &v12.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      constants.GrafanaAdminSecretName,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				constants.LastCredentialsAnnotation: hash,
			},
		},
		Data: data,
		Type: v12.SecretTypeOpaque,
	}
}

func AdminSecretReconciled(cr *v1alpha1.Grafana, currentState *v12.Secret) *v12.Secret {
	data, hash := getData(cr, currentState)

	reconciled := currentState.DeepCopy()
	reconciled.Data = data

	if reconciled.Annotations == nil {
		reconciled.Annotations = map[string]string{}
	}
	reconciled.Annotations[constants.LastCredentialsAnnotation] = hash

	return reconciled
}

func AdminSecretSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      constants.GrafanaAdminSecretName,
	}
}
