package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const OpenShiftOAuthRedirect = "serviceaccounts.openshift.io/oauth-redirectreference.primary"

func getServiceAccountLabels(cr *v1alpha1.Grafana) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return nil
	}
	return cr.Spec.ServiceAccount.Labels
}

func getServiceAccountAnnotations(cr *v1alpha1.Grafana) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return nil
	}
	return cr.Spec.ServiceAccount.Annotations
}

func GrafanaServiceAccount(cr *v1alpha1.Grafana) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaServiceAccountName,
			Namespace:   cr.Namespace,
			Labels:      getServiceAccountLabels(cr),
			Annotations: getServiceAccountAnnotations(cr),
		},
	}
}

func GrafanaServiceAccountSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaServiceAccountName,
	}
}

func GrafanaServiceAccountReconciled(cr *v1alpha1.Grafana, currentState *v1.ServiceAccount) *v1.ServiceAccount {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getServiceAccountLabels(cr)
	reconciled.Annotations = getServiceAccountAnnotations(cr)
	return reconciled
}
