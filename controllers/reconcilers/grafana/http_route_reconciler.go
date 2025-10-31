package grafana

import (
	"context"
	"fmt"
	"slices"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	v2 "sigs.k8s.io/gateway-api/apis/v1"
)

// HTTPRouteReconciler is responsible for managing and reconciling
// the Gateway API HTTPRoute resource associated with a Grafana instance.
type HTTPRouteReconciler struct {
	client client.Client
}

// NewHTTPRouteReconciler returns a new instance of HTTPRouteReconciler.
func NewHTTPRouteReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &HTTPRouteReconciler{
		client: client,
	}
}

// Reconcile runs the main reconciliation logic for Grafana HTTPRoute.
// It ensures that the desired HTTPRoute exists and is configured correctly.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("HTTPRouteReconciler")

	log.Info("reconciling http route")

	return r.reconcileHTTPRoute(ctx, cr, vars, scheme)
}

// reconcileHTTPRoute ensures the HTTPRoute object for Grafana matches the desired spec.
// It creates or updates the HTTPRoute, merges configurations, and updates Grafana’s AdminURL if needed.
func (r *HTTPRouteReconciler) reconcileHTTPRoute(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	if cr.Spec.HTTPRoute == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	httpRoute := model.GetGrafanaHTTPRoute(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, httpRoute, func() error {
		httpRoute.Spec = getHTTPRouteSpec(cr, scheme)

		// Merge the CR-defined HTTPRoute into the generated object
		err := v1beta1.Merge(cr.Spec.HTTPRoute, httpRoute)
		if err != nil {
			setInvalidMergeCondition(cr, "HTTPRoute", err)
			return err
		}

		removeInvalidMergeCondition(cr, "HTTPRoute")

		// Set Grafana as the controller owner
		err = controllerutil.SetControllerReference(cr, httpRoute, scheme)
		if err != nil {
			return err
		}

		// Propagate labels from Grafana CR to the HTTPRoute
		model.SetInheritedLabels(httpRoute, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	// Assign AdminURL if ingress is preferred
	if cr.PreferIngress() {
		adminURL := r.getHTTPRouteAdminURL(ctx, httpRoute)

		// Wait until route has parents (attached to Gateway)
		if len(httpRoute.Status.Parents) == 0 {
			return v1beta1.OperatorStageResultInProgress, fmt.Errorf("http route is not ready yet")
		}

		if adminURL == "" {
			return v1beta1.OperatorStageResultFailed, fmt.Errorf("http route spec is incomplete")
		}

		cr.Status.AdminURL = adminURL
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

// getHTTPRouteAdminURL builds the external access URL for Grafana based on the
// associated Gateway’s status and listener configuration. Returns empty string
// if the Gateway is not ready or missing an address.
func (r *HTTPRouteReconciler) getHTTPRouteAdminURL(ctx context.Context, httpRoute *v2.HTTPRoute) (adminURL string) {
	log := logf.FromContext(ctx)

	if httpRoute == nil {
		return ""
	}

	if len(httpRoute.Spec.ParentRefs) == 0 {
		return ""
	}

	var (
		protocol  = "http"
		namespace = "default"
		hostname  string
		port      int
	)

	// Fetch the Gateway referenced by the HTTPRoute
	gw := &v2.Gateway{}
	pr := httpRoute.Spec.ParentRefs[0]

	if pr.Namespace != nil {
		namespace = string(*pr.Namespace)
	}

	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      string(pr.Name),
	}, gw)
	if err != nil {
		log.Error(err, "error synchronizing grafana statuses")
		return ""
	}

	// Match appropriate listener from Gateway
	listener := r.getMatchListener(ctx, httpRoute, gw)

	if listener != nil {
		if listener.TLS != nil {
			protocol = "https"
		}

		port = int(listener.Port)

		if listener.Hostname != nil {
			hostname = string(*listener.Hostname)
		}

		if hostname == "" {
			for _, address := range gw.Status.Addresses {
				if address.Value != "" {
					hostname = address.Value
					break
				}
			}
		}
	}

	// Wait until Gateway has an assigned address
	if hostname == "" {
		log.Info("gateway has no assigned address yet; waiting for it to become ready",
			"gateway", gw.Name, "namespace", gw.Namespace)

		return ""
	}

	return fmt.Sprintf("%v://%v:%v", protocol, hostname, port)
}

// getMatchListener tries to find a Gateway listener that matches the HTTPRoute’s
// namespace and hostname constraints, according to the Gateway API rules.
func (r *HTTPRouteReconciler) getMatchListener(ctx context.Context, httpRoute *v2.HTTPRoute, gw *v2.Gateway) *v2.Listener {
	log := logf.FromContext(ctx)
	hostnames := httpRoute.Spec.Hostnames

	for _, listener := range gw.Spec.Listeners {
		if listener.AllowedRoutes != nil && listener.AllowedRoutes.Namespaces != nil {
			switch *listener.AllowedRoutes.Namespaces.From {
			case v2.NamespacesFromAll:
				if len(hostnames) == 0 || slices.Contains(hostnames, *listener.Hostname) {
					return &listener
				}
			case v2.NamespacesFromSelector:
				var (
					nsList = corev1.NamespaceList{}

					opts = []client.ListOption{
						client.MatchingLabels(listener.AllowedRoutes.Namespaces.Selector.MatchLabels),
					}
				)

				if err := r.client.List(ctx, &nsList, opts...); err != nil {
					log.Error(err, "error fetching namespace for http route")
					continue
				}

				if len(nsList.Items) > 0 {
					for _, item := range nsList.Items {
						if item.Namespace == httpRoute.Namespace && labelsSatisfyMatchExpressions(item.Labels,
							listener.AllowedRoutes.Namespaces.Selector.MatchExpressions) {
							if len(hostnames) == 0 || slices.Contains(hostnames, *listener.Hostname) {
								return &listener
							}
						}
					}
				}
			case v2.NamespacesFromSame:
				if gw.Namespace == httpRoute.Namespace {
					if len(hostnames) == 0 || slices.Contains(hostnames, *listener.Hostname) {
						return &listener
					}
				}
			case v2.NamespacesFromNone:
				continue
			}
		}
	}

	return nil
}

// GetHTTPRouteTargetPort returns the target port number that should be used
// in HTTPRoute backend references for Grafana’s service.
func GetHTTPRouteTargetPort(cr *v1beta1.Grafana) intstr.IntOrString {
	return intstr.FromInt(GetGrafanaPort(cr))
}

// getHTTPRouteSpec builds the desired HTTPRouteSpec for Grafana,
// wiring its rules to the underlying Service backend.
func getHTTPRouteSpec(cr *v1beta1.Grafana, scheme *runtime.Scheme) v2.HTTPRouteSpec {
	service := model.GetGrafanaService(cr, scheme)

	port := GetHTTPRouteTargetPort(cr)
	serviceName := v2.ObjectName(service.GetName())
	serviceNamespace := v2.Namespace(service.GetNamespace())
	servicePort := v2.PortNumber(port.IntVal)

	return v2.HTTPRouteSpec{
		CommonRouteSpec: v2.CommonRouteSpec{
			ParentRefs: cr.Spec.HTTPRoute.Spec.ParentRefs,
		},
		Hostnames: cr.Spec.HTTPRoute.Spec.Hostnames,
		Rules: func(cr *v1beta1.Grafana) []v2.HTTPRouteRule {
			{
				rules := make([]v2.HTTPRouteRule, 0, len(cr.Spec.HTTPRoute.Spec.Rules))
				for _, rule := range cr.Spec.HTTPRoute.Spec.Rules {
					rule.BackendRefs = []v2.HTTPBackendRef{
						{
							BackendRef: v2.BackendRef{
								BackendObjectReference: v2.BackendObjectReference{
									Name:      serviceName,
									Namespace: &serviceNamespace,
									Port:      &servicePort,
								},
							},
						},
					}
					rules = append(rules, rule)
				}
				return rules
			}
		}(cr),
	}
}

// labelsSatisfyMatchExpressions checks if a given label set satisfies
// a list of Kubernetes label selector requirements.
func labelsSatisfyMatchExpressions(labels map[string]string, matchExpressions []metav1.LabelSelectorRequirement) bool {
	// Support for empty selector: treat as match-all
	if len(labels) == 0 {
		return true
	}

	for _, matchExpression := range matchExpressions {
		selected := false

		if label, ok := labels[matchExpression.Key]; ok {
			switch matchExpression.Operator {
			case metav1.LabelSelectorOpDoesNotExist:
				selected = false
			case metav1.LabelSelectorOpExists:
				selected = true
			case metav1.LabelSelectorOpIn:
				selected = slices.Contains(matchExpression.Values, label)
			case metav1.LabelSelectorOpNotIn:
				selected = !slices.Contains(matchExpression.Values, label)
			}
		}

		// All matchExpressions must evaluate to true
		if !selected {
			return false
		}
	}

	return true
}
