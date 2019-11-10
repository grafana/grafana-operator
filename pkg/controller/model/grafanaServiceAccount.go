package model

import (
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const OpenShiftOAuthRedirect = "serviceaccounts.openshift.io/oauth-redirectreference.primary"

func applyAnnotations(sa *v1.ServiceAccount) *v1.ServiceAccount {
	cfg := config.GetControllerConfig()
	openshift := cfg.GetConfigBool(config.ConfigOpenshift, false)
	if openshift {
		sa.Annotations = map[string]string{
			OpenShiftOAuthRedirect: fmt.Sprintf(`{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"%s"}}`, GrafanaRouteName),
		}
	}
	return sa
}

func GrafanaServiceAccount(cr *v1alpha1.Grafana) *v1.ServiceAccount {
	return applyAnnotations(&v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:      GrafanaServiceAccountName,
			Namespace: cr.Namespace,
		},
	})
}

func GrafanaServiceAccountSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaServiceAccountName,
	}
}

func GrafanaServiceAccountReconciled(cr *v1alpha1.Grafana, currentState *v1.ServiceAccount) *v1.ServiceAccount {
	return applyAnnotations(currentState.DeepCopy())
}
