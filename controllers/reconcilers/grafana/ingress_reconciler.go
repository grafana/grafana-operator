package grafana

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type IngressReconciler struct {
	client client.Client
}

func NewIngressReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &IngressReconciler{
		client: client,
	}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx)

	if r.isOpenShift() {
		logger.Info("platform is OpenShift, creating Route")
		return r.reconcileRoute()
	} else {
		logger.Info("platform is Kubernetes, creating Ingress")
		return r.reconcileIngress(ctx, cr, status, vars, scheme)
	}
}

func (r *IngressReconciler) reconcileIngress(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	ingress := model.GetGrafanaIngress(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, ingress, func() error {
		ingress.Spec = getIngressSpec(cr, scheme)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) reconcileRoute() (v1beta1.OperatorStageStatus, error) {
	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *IngressReconciler) isOpenShift() bool {
	return false
}

func getIngressTLS(cr *v1beta1.Grafana) []v1.IngressTLS {
	if cr.Spec.Ingress == nil {
		return nil
	}

	if cr.Spec.Ingress.TLSEnabled {
		return []v1.IngressTLS{
			{
				Hosts:      []string{cr.Spec.Ingress.Hostname},
				SecretName: cr.Spec.Ingress.TLSSecretName,
			},
		}
	}
	return nil
}

func GetIngressPathType(cr *v1beta1.Grafana) *v1.PathType {

	if cr.Spec.Ingress == nil {
		return nil
	}

	t := v1.PathType(cr.Spec.Ingress.PathType)
	switch t {
	case v1.PathTypeExact, v1.PathTypePrefix:
		return &t
	case v1.PathTypeImplementationSpecific:
		t = v1.PathTypeImplementationSpecific
		return &t
	}

	d := v1.PathTypeExact
	return &d
}

func GetIngressClassName(cr *v1beta1.Grafana) *string {
	if cr.Spec.Ingress == nil || cr.Spec.Ingress.IngressClassName == "" {
		return nil
	}

	return &cr.Spec.Ingress.IngressClassName
}

func GetIngressTargetPort(cr *v1beta1.Grafana) intstr.IntOrString {
	defaultPort := intstr.FromInt(GetGrafanaPort(cr))

	if cr.Spec.Ingress == nil {
		return defaultPort
	}

	if cr.Spec.Ingress.TargetPort == "" {
		return defaultPort
	}

	return intstr.FromString(cr.Spec.Ingress.TargetPort)
}

func GetHost(cr *v1beta1.Grafana) string {
	if cr.Spec.Ingress == nil {
		return ""
	}
	return cr.Spec.Ingress.Hostname
}

func GetPath(cr *v1beta1.Grafana) string {
	if cr.Spec.Ingress == nil {
		return "/"
	}

	if cr.Spec.Ingress.Path == "" {
		return "/"
	}

	return cr.Spec.Ingress.Path
}

func getIngressSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) v1.IngressSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetIngressTargetPort(cr)

	if port.IntVal != 0 {
		return v1.IngressSpec{
			TLS:              getIngressTLS(cr),
			IngressClassName: GetIngressClassName(cr),
			Rules: []v1.IngressRule{
				{
					Host: GetHost(cr),
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									Path:     GetPath(cr),
									PathType: GetIngressPathType(cr),
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: service.Name,
											Port: v1.ServiceBackendPort{
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
	return v1.IngressSpec{
		TLS:              getIngressTLS(cr),
		IngressClassName: GetIngressClassName(cr),
		Rules: []v1.IngressRule{
			{
				Host: GetHost(cr),
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path:     GetPath(cr),
								PathType: GetIngressPathType(cr),
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: service.Name,
										Port: v1.ServiceBackendPort{
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
