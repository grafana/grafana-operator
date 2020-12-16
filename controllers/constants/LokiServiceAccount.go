package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LokiServiceAccountSelector(cr *v1alpha1.Loki) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiServiceAccountName,
	}
}

func getLokiServiceAccountLabels(cr *v1alpha1.Loki) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return nil
	}
	return cr.Spec.ServiceAccount.Labels
}

func getLokiServiceAccountAnnotations(cr *v1alpha1.Loki, existing map[string]string) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return existing
	}
	return MergeAnnotations(cr.Spec.ServiceAccount.Annotations, existing)
}

func LokiServiceAccount(cr *v1alpha1.Loki) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaServiceAccountName,
			Namespace:   cr.Namespace,
			Labels:      getLokiServiceAccountLabels(cr),
			Annotations: getLokiServiceAccountAnnotations(cr, nil),
		},
	}
}

func LokiServiceAccountReconciled(cr *v1alpha1.Loki, currentState *v1.ServiceAccount) *v1.ServiceAccount {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getLokiServiceAccountLabels(cr)
	reconciled.Annotations = getLokiServiceAccountAnnotations(cr, currentState.Annotations)
	return reconciled
}
