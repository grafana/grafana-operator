package model

import (
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getAlertServiceName() string {
	return "grafana-alert"
}

func getAlertServiceLabels(cr *v1alpha1.Grafana) map[string]string {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service.Labels
}

func GrafanaAlertService(cr *v1alpha1.Grafana) *v1.Service {
	return &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        getAlertServiceName(),
			Namespace:   cr.Namespace,
			Labels:      getAlertServiceLabels(cr),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       "grafana-alert",
					Port:       9094,
					TargetPort: intstr.FromString("grafana-alert"),
				},
			},
			Selector: map[string]string{
				"app": constants.GrafanaPodLabel,
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
}

func GrafanaAlertServiceReconciled(cr *v1alpha1.Grafana, currentState *v1.Service) *v1.Service {
	reconciled := currentState.DeepCopy()
	return reconciled
}

func GrafanaAlertServiceSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      getAlertServiceName(),
	}
}
