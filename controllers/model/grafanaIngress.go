package model

import (
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/constants"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getIngressTLS(cr *v1alpha1.Grafana) []netv1.IngressTLS {
	if cr.Spec.Ingress == nil {
		return nil
	}

	if cr.Spec.Ingress.TLSEnabled {
		return []netv1.IngressTLS{
			{
				Hosts:      []string{cr.Spec.Ingress.Hostname},
				SecretName: cr.Spec.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func GetIngressPathType(cr *v1alpha1.Grafana) *netv1.PathType {
	t := netv1.PathType(cr.Spec.Ingress.PathType)
	switch t {
	case netv1.PathTypeExact, netv1.PathTypePrefix:
		return &t
	case netv1.PathTypeImplementationSpecific:
		t = netv1.PathTypeImplementationSpecific
		return &t
	}
	return nil
}

func GetIngressClassName(cr *v1alpha1.Grafana) *string {
	if cr.Spec.Ingress.IngressClassName == "" {
		return nil
	}

	return &cr.Spec.Ingress.IngressClassName
}

func getIngressSpec(cr *v1alpha1.Grafana) netv1.IngressSpec {
	serviceName := func(cr *v1alpha1.Grafana) string {
		if cr.Spec.Service != nil && cr.Spec.Service.Name != "" {
			return cr.Spec.Service.Name
		}
		return constants.GrafanaServiceName
	}
	port := GetIngressTargetPort(cr)

	if port.IntVal != 0 {
		return netv1.IngressSpec{
			TLS:              getIngressTLS(cr),
			IngressClassName: GetIngressClassName(cr),
			Rules: []netv1.IngressRule{
				{
					Host: GetHost(cr),
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     GetPath(cr),
									PathType: GetIngressPathType(cr),
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: serviceName(cr),
											Port: netv1.ServiceBackendPort{
												Number: port.IntVal,
											},
										},
										Resource: nil,
									},
								},
							},
						},
					},
				},
			},
		}
	}
	return netv1.IngressSpec{
		TLS:              getIngressTLS(cr),
		IngressClassName: GetIngressClassName(cr),
		Rules: []netv1.IngressRule{
			{
				Host: GetHost(cr),
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								Path:     GetPath(cr),
								PathType: GetIngressPathType(cr),
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: serviceName(cr),
										Port: netv1.ServiceBackendPort{
											Name: port.StrVal,
										},
									},
									Resource: nil,
								},
							},
						},
					},
				},
			},
		},
	}
}

func GrafanaIngress(cr *v1alpha1.Grafana) *netv1.Ingress {
	return &netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        constants.GrafanaIngressName,
			Namespace:   cr.Namespace,
			Labels:      GetIngressLabels(cr),
			Annotations: GetIngressAnnotations(cr, nil),
		},
		Spec: getIngressSpec(cr),
	}
}

func GrafanaIngressReconciled(cr *v1alpha1.Grafana, currentState *netv1.Ingress) *netv1.Ingress {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = GetIngressLabels(cr)
	reconciled.Annotations = GetIngressAnnotations(cr, currentState.Annotations)
	reconciled.Spec = getIngressSpec(cr)
	return reconciled
}

func GrafanaIngressSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      constants.GrafanaIngressName,
	}
}
