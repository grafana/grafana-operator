package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1a1 "k8s.io/api/rbac/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaProxyRoleBinding(cr *v1alpha1.GrafanaProxy) *v1a1.RoleBinding {
	return &v1a1.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:        GrafanaProxyRoleBindingName,
			Namespace:   cr.Namespace,
			Labels:      getProxyPodLabels(cr),
			Annotations: getProxyPodAnnotations(cr),
		},
		Subjects: []v1a1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      GrafanaProxyServiceAccountName,
				Namespace: cr.Namespace,
			},
		},
		RoleRef: v1a1.RoleRef{
			Kind:     "ClusterRole",
			Name:     GrafanaProxyServiceAccountName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func GrafanaProxyRoleBindingReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1a1.RoleBinding) *v1a1.RoleBinding {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getProxyPodLabels(cr)
	return reconciled
}

func GrafanaProxyRoleBindingSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyRoleBindingName,
	}
}
