package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type HTTPRouteReconciler struct {
	client client.Client
}

func NewHTTPRouteReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &HTTPRouteReconciler{
		client: client,
	}
}

func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("HTTPRouteReconciler")

	if cr.Spec.HTTPRoute == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	log.Info("reconciling httproute")

	httpRoute := model.GetGrafanaHTTPRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, httpRoute, func() error {
		// Get default backendRefs that point to Grafana service
		service := model.GetGrafanaService(cr, scheme)
		port := gatewayv1.PortNumber(GetGrafanaPort(cr)) //nolint:gosec // Port number is always valid Grafana port

		defaultBackendRefs := []gatewayv1.HTTPBackendRef{
			{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: gatewayv1.ObjectName(service.Name),
						Port: &port,
					},
				},
			},
		}

		// Start with base spec (empty parentRefs, hostnames, default backendRefs)
		httpRoute.Spec = gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{},
			},
			Hostnames: []gatewayv1.Hostname{},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					BackendRefs: defaultBackendRefs,
				},
			},
		}

		// Merge user overrides (parentRefs, hostnames, custom rules if specified)
		err := v1beta1.Merge(httpRoute, cr.Spec.HTTPRoute)
		if err != nil {
			setInvalidMergeCondition(cr, "HTTPRoute", err)
			return err
		}

		// Ensure backendRefs are set if user didn't provide custom rules or backendRefs
		if len(httpRoute.Spec.Rules) == 0 {
			httpRoute.Spec.Rules = []gatewayv1.HTTPRouteRule{{BackendRefs: defaultBackendRefs}}
		} else if len(httpRoute.Spec.Rules[0].BackendRefs) == 0 {
			httpRoute.Spec.Rules[0].BackendRefs = defaultBackendRefs
		}

		removeInvalidMergeCondition(cr, "HTTPRoute")

		err = controllerutil.SetControllerReference(cr, httpRoute, scheme)
		if err != nil {
			return err
		}

		model.SetInheritedLabels(httpRoute, cr.Labels)

		return nil
	})
	if err != nil {
		// If Gateway API CRDs are not installed, gracefully skip HTTPRoute reconciliation
		if meta.IsNoMatchError(err) {
			log.Info("Gateway API CRDs not found, skipping HTTPRoute reconciliation. Install Gateway API CRDs to enable HTTPRoute support.")
			return v1beta1.OperatorStageResultSuccess, nil
		}

		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

