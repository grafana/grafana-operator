package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetIngressPathType(cr *v1alpha1.Grafana) *v12.PathType {
	t := v12.PathType(cr.Spec.Ingress.PathType)
	switch t {
	case v12.PathTypeExact, v12.PathTypePrefix:
		return &t
	}
	t = v12.PathTypeImplementationSpecific
	return &t
}

func GetIngressClassName(cr *v1alpha1.Grafana) *string {
	if cr.Spec.Ingress.IngressClassName == "" {
		return nil
	}

	return &cr.Spec.Ingress.IngressClassName
}

func getIngressTLS(cr *v1alpha1.Grafana) []v12.IngressTLS {
	if cr.Spec.Ingress == nil {
		return nil
	}

	if cr.Spec.Ingress.TLSEnabled {
		return []v12.IngressTLS{
			{
				Hosts:      []string{cr.Spec.Ingress.Hostname},
				SecretName: cr.Spec.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func getIngressSpec(cr *v1alpha1.Grafana) v12.IngressSpec {
	serviceName := func(cr *v1alpha1.Grafana) string {
		if cr.Spec.Service != nil && cr.Spec.Service.Name != "" {
			return cr.Spec.Service.Name
		}
		return GrafanaServiceName
	}
	return v12.IngressSpec{
		TLS:              getIngressTLS(cr),
		IngressClassName: GetIngressClassName(cr),
		Rules: []v12.IngressRule{
			{
				Host: GetHost(cr),
				IngressRuleValue: v12.IngressRuleValue{
					HTTP: &v12.HTTPIngressRuleValue{
						Paths: []v12.HTTPIngressPath{
							{
								Path:     GetPath(cr),
								PathType: GetIngressPathType(cr),
								Backend: v12.IngressBackend{
									Service: &v12.IngressServiceBackend{
										Name: serviceName(cr),
										Port: v12.ServiceBackendPort{
											Name:   "http",
											Number: GetIngressTargetPort(cr).IntVal,
										},
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

func GrafanaIngress(cr *v1alpha1.Grafana) *v12.Ingress {
	return &v12.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        GrafanaIngressName,
			Namespace:   cr.Namespace,
			Labels:      GetIngressLabels(cr),
			Annotations: GetIngressAnnotations(cr, nil),
		},
		Spec: getIngressSpec(cr),
	}
}

func GrafanaIngressReconciled(cr *v1alpha1.Grafana, currentState *v12.Ingress) *v12.Ingress {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = GetIngressLabels(cr)
	reconciled.Annotations = GetIngressAnnotations(cr, currentState.Annotations)
	reconciled.Spec = getIngressSpec(cr)
	return reconciled
}

func GrafanaIngressSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaIngressName,
	}
}
