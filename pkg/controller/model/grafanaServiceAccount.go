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

func getServiceAccountAnnotations(cr *v1alpha1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return existing
	}
	return MergeAnnotations(cr.Spec.ServiceAccount.Annotations, existing)
}

func mergeImagePullSecrets(requested []v1.LocalObjectReference, existing []v1.LocalObjectReference) []v1.LocalObjectReference {
	appendIfAbsent := func(secrets []v1.LocalObjectReference, secret v1.LocalObjectReference) []v1.LocalObjectReference {
		for _, s := range secrets {
			if s.Name == secret.Name {
				return secrets
			}
		}
		return append(secrets, secret)
	}

	for _, s := range requested {
		existing = appendIfAbsent(existing, s)
	}

	return existing
}

func getServiceAccountImagePullSecrets(cr *v1alpha1.Grafana, existing []v1.LocalObjectReference) []v1.LocalObjectReference {
	if cr.Spec.ServiceAccount == nil {
		return existing
	}
	return mergeImagePullSecrets(cr.Spec.ServiceAccount.ImagePullSecrets, existing)
}

func GrafanaServiceAccount(cr *v1alpha1.Grafana) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaServiceAccountName,
			Namespace:   cr.Namespace,
			Labels:      getServiceAccountLabels(cr),
			Annotations: getServiceAccountAnnotations(cr, nil),
		},
		ImagePullSecrets: getServiceAccountImagePullSecrets(cr, nil),
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
	reconciled.Annotations = getServiceAccountAnnotations(cr, currentState.Annotations)
	reconciled.ImagePullSecrets = getServiceAccountImagePullSecrets(cr, currentState.ImagePullSecrets)
	return reconciled
}
