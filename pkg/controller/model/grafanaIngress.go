package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getIngressTLS(cr *v1alpha1.Grafana) []v1beta1.IngressTLS {
	if cr.Spec.Ingress.TLSEnabled {
		return []v1beta1.IngressTLS{
			{
				Hosts:      []string{cr.Spec.Ingress.Hostname},
				SecretName: cr.Spec.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func GrafanaIngress(cr *v1alpha1.Grafana) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        GrafanaIngressName,
			Namespace:   cr.Namespace,
			Labels:      cr.Spec.Ingress.Labels,
			Annotations: cr.Spec.Ingress.Annotations,
		},
		Spec: v1beta1.IngressSpec{
			TLS: getIngressTLS(cr),
			Rules: []v1beta1.IngressRule{
				{
					Host: cr.Spec.Ingress.Hostname,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: cr.Spec.Ingress.Path,
									Backend: v1beta1.IngressBackend{
										ServiceName: GrafanaServiceName,
										ServicePort: intstr.FromInt(int(GetGrafanaPort(cr))),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func GrafanaIngressReconciled(cr *v1alpha1.Grafana, currentState *v1beta1.Ingress) *v1beta1.Ingress {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = cr.Spec.Ingress.Labels
	reconciled.Annotations = cr.Spec.Ingress.Annotations
	reconciled.Spec = v1beta1.IngressSpec{
		TLS: getIngressTLS(cr),
		Rules: []v1beta1.IngressRule{
			{
				Host: cr.Spec.Ingress.Hostname,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: cr.Spec.Ingress.Path,
								Backend: v1beta1.IngressBackend{
									ServiceName: GrafanaServiceName,
									ServicePort: intstr.FromInt(int(GetGrafanaPort(cr))),
								},
							},
						},
					},
				},
			},
		},
	}
	return reconciled
}

func GrafanaIngressSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaIngressName,
	}
}
