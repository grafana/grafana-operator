package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getLokiIngressTLS(cr *v1alpha1.Loki) []v1beta1.IngressTLS {
	if cr.Spec.Ingress == nil {
		return nil
	}

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

func getLokiIngressSpec(cr *v1alpha1.Loki) v1beta1.IngressSpec {
	return v1beta1.IngressSpec{
		TLS: getLokiIngressTLS(cr),
		Rules: []v1beta1.IngressRule{
			{
				Host: GetLokiHost(cr),
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: GetLokiPath(cr),
								Backend: v1beta1.IngressBackend{
									ServiceName: LokiServiceName,
									ServicePort: GetLokiIngressTargetPort(cr),
								},
							},
						},
					},
				},
			},
		},
	}
}

func LokiIngress(cr *v1alpha1.Loki) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        LokiIngressName,
			Namespace:   cr.Namespace,
			Labels:      GetLokiIngressLabels(cr),
			Annotations: GetLokiIngressAnnotations(cr, nil),
		},
		Spec: getLokiIngressSpec(cr),
	}
}

func LokiIngressReconciled(cr *v1alpha1.Loki, currentState *v1beta1.Ingress) *v1beta1.Ingress {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = GetLokiIngressLabels(cr)
	reconciled.Annotations = GetLokiIngressAnnotations(cr, currentState.Annotations)
	reconciled.Spec = getLokiIngressSpec(cr)
	return reconciled
}

func LokiIngressSelector(cr *v1alpha1.Loki) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiIngressName,
	}
}
