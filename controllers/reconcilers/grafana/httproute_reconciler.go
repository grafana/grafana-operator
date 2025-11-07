package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	protocolHTTP  = "http"
	protocolHTTPS = "https"
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
		httpRoute.Spec = getHTTPRouteSpec(cr, scheme)

		err := v1beta1.Merge(httpRoute, cr.Spec.HTTPRoute)
		if err != nil {
			setInvalidMergeCondition(cr, "HTTPRoute", err)
			return err
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
		return v1beta1.OperatorStageResultFailed, err
	}

	// try to assign the admin url if PreferHTTPRoute is enabled
	if cr.PreferHTTPRoute() {
		adminURL, err := r.getHTTPRouteAdminURL(ctx, httpRoute)
		if err != nil {
			return v1beta1.OperatorStageResultInProgress, fmt.Errorf("httproute is not ready yet: %w", err)
		}

		if adminURL == "" {
			return v1beta1.OperatorStageResultInProgress, fmt.Errorf("httproute spec is incomplete")
		}

		cr.Status.AdminURL = adminURL
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getHTTPRouteSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) gatewayv1.HTTPRouteSpec {
	service := model.GetGrafanaService(cr, scheme)
	port := gatewayv1.PortNumber(GetGrafanaPort(cr)) //nolint:gosec // Port number is always valid Grafana port

	return gatewayv1.HTTPRouteSpec{
		CommonRouteSpec: gatewayv1.CommonRouteSpec{
			ParentRefs: []gatewayv1.ParentReference{},
		},
		Hostnames: []gatewayv1.Hostname{},
		Rules: []gatewayv1.HTTPRouteRule{
			{
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: gatewayv1.ObjectName(service.Name),
								Port: &port,
							},
						},
					},
				},
			},
		},
	}
}

// getHTTPRouteAdminURL returns the admin URL for accessing Grafana via HTTPRoute.
// It performs Gateway lookup to determine the correct protocol and hostname.
func (r *HTTPRouteReconciler) getHTTPRouteAdminURL(ctx context.Context, httpRoute *gatewayv1.HTTPRoute) (string, error) {
	if httpRoute == nil {
		return "", fmt.Errorf("httproute is nil")
	}

	// Try to get hostname from HTTPRoute spec first
	var hostname string
	if len(httpRoute.Spec.Hostnames) > 0 {
		hostname = string(httpRoute.Spec.Hostnames[0])
	}

	// Default protocol
	protocol := protocolHTTP

	// Try to determine protocol by looking up Gateway and its listeners
	if len(httpRoute.Spec.ParentRefs) > 0 {
		parentRef := httpRoute.Spec.ParentRefs[0]

		// Lookup Gateway
		gateway := &gatewayv1.Gateway{}
		gatewayName := types.NamespacedName{
			Namespace: httpRoute.Namespace, // default to HTTPRoute namespace
			Name:      string(parentRef.Name),
		}

		// Use explicit namespace from parentRef if provided
		if parentRef.Namespace != nil {
			gatewayName.Namespace = string(*parentRef.Namespace)
		}

		err := r.client.Get(ctx, gatewayName, gateway)
		if err != nil {
			// Gateway not found or error - fallback to simple logic
			// Assume HTTPS if parentRefs are present (production use case)
			protocol = protocolHTTPS
		} else {
			// Gateway found - try to find matching listener
			listener := r.findMatchingListener(gateway, parentRef.SectionName)
			if listener != nil {
				// Determine protocol from listener
				if listener.Protocol == gatewayv1.HTTPSProtocolType || listener.TLS != nil {
					protocol = protocolHTTPS
				}

				// If hostname is not set in HTTPRoute, try to get it from listener or Gateway status
				if hostname == "" {
					if listener.Hostname != nil {
						hostname = string(*listener.Hostname)
					} else if len(gateway.Status.Addresses) > 0 {
						// Fallback to Gateway address
						addr := gateway.Status.Addresses[0]
						if addr.Type != nil && *addr.Type == gatewayv1.HostnameAddressType {
							hostname = addr.Value
						} else if addr.Type != nil && *addr.Type == gatewayv1.IPAddressType {
							hostname = addr.Value
						}
					}
				}
			} else {
				// No matching listener found, assume HTTPS for security
				protocol = protocolHTTPS
			}
		}
	}

	// Final check: if hostname is still empty, return error
	if hostname == "" {
		return "", fmt.Errorf("no hostname found in HTTPRoute spec or Gateway status")
	}

	return fmt.Sprintf("%s://%s", protocol, hostname), nil
}

// findMatchingListener finds the listener in Gateway that matches the sectionName from parentRef.
// If sectionName is nil, returns the first listener.
func (r *HTTPRouteReconciler) findMatchingListener(gateway *gatewayv1.Gateway, sectionName *gatewayv1.SectionName) *gatewayv1.Listener {
	if gateway == nil || len(gateway.Spec.Listeners) == 0 {
		return nil
	}

	// If no sectionName specified, return first listener
	if sectionName == nil {
		return &gateway.Spec.Listeners[0]
	}

	// Find listener with matching name
	for i := range gateway.Spec.Listeners {
		listener := &gateway.Spec.Listeners[i]
		if listener.Name == gatewayv1.SectionName(*sectionName) {
			return listener
		}
	}

	// No match found, return first listener as fallback
	return &gateway.Spec.Listeners[0]
}
