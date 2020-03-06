package model

import (
	"strings"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getProxyIngressTLS(cr *v1alpha1.GrafanaProxy) []v1beta1.IngressTLS {
	return []v1beta1.IngressTLS{
		{
			Hosts:      []string{cr.Spec.Config.Hostname},
			SecretName: "tls-" + strings.ReplaceAll(cr.Spec.Config.Hostname, ".", "-"),
		},
	}

}

func getProxyIngressSpec(cr *v1alpha1.GrafanaProxy) v1beta1.IngressSpec {
	return v1beta1.IngressSpec{
		TLS: getProxyIngressTLS(cr),
		Rules: []v1beta1.IngressRule{
			{
				Host: cr.Spec.Config.Hostname,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: "/",
								Backend: v1beta1.IngressBackend{
									ServiceName: GrafanaProxyServiceName,
									ServicePort: intstr.FromInt(80),
								},
							},
						},
					},
				},
			},
		},
	}
}

func GrafanaProxyIngress(cr *v1alpha1.GrafanaProxy) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        GrafanaProxyIngressName,
			Namespace:   cr.Namespace,
			Labels:      getProxyPodLabels(cr),
			Annotations: getProxyPodAnnotations(cr),
		},
		Spec: getProxyIngressSpec(cr),
	}
}

func GrafanaProxyIngressReconciled(cr *v1alpha1.GrafanaProxy, currentState *v1beta1.Ingress) *v1beta1.Ingress {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getProxyPodLabels(cr)
	reconciled.Annotations = map[string]string{"vice-president": "true"}
	reconciled.Spec = getProxyIngressSpec(cr)
	return reconciled
}

func GrafanaProxyIngressSelector(cr *v1alpha1.GrafanaProxy) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaProxyIngressName,
	}
}
