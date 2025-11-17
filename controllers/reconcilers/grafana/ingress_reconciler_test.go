package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("Allow use of Ingress on OpenShift", func() {
	t := GinkgoT()

	It("Creates Ingress on OpenShift when .spec.ingress is defined", func() {
		r := NewIngressReconciler(k8sClient, true, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-on-openshift",
				Namespace: "default",
				Labels:    map[string]string{"openshift": "ingress"},
			},
			Spec: v1beta1.GrafanaSpec{
				Ingress: &v1beta1.IngressNetworkingV1{},
				Route:   nil,
			},
		}

		ctx := context.Background()

		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		ingress := &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)
		require.NoError(t, err)
	})

	It("Creates Route on OpenShift when .spec.ingress AND .spec.route is defined", func() {
		r := NewIngressReconciler(k8sClient, true, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prefer-route-on-openshift",
				Namespace: "default",
				Labels:    map[string]string{"openshift": "route"},
			},
			Spec: v1beta1.GrafanaSpec{
				Ingress: &v1beta1.IngressNetworkingV1{},
				Route: &v1beta1.RouteOpenshiftV1{
					Spec: &v1beta1.RouteOpenShiftV1Spec{},
				},
			},
		}

		ctx := context.Background()

		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		route := &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)
		require.NoError(t, err)
	})

	It("Removes Route when .spec.route is removed", func() {
		r := NewIngressReconciler(k8sClient, true, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route-nil",
				Namespace: "default",
				Labels:    map[string]string{"openshift": "route"},
			},
			Spec: v1beta1.GrafanaSpec{
				Route: &v1beta1.RouteOpenshiftV1{
					Spec: &v1beta1.RouteOpenShiftV1Spec{},
				},
			},
		}

		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		route := &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)
		require.NoError(t, err)

		cr.Spec.Route = nil

		err = k8sClient.Update(ctx, cr)
		require.NoError(t, err)

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		route = &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)

		assert.True(t, kuberr.IsNotFound(err))
	})

	It("Removes Ingress when .spec.ingress is removed", func() {
		r := NewIngressReconciler(k8sClient, false, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-nil",
				Namespace: "default",
				Labels:    map[string]string{},
			},
			Spec: v1beta1.GrafanaSpec{
				Ingress: &v1beta1.IngressNetworkingV1{},
			},
		}

		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		ingress := &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)
		require.NoError(t, err)

		cr.Spec.Ingress = nil

		err = k8sClient.Update(ctx, cr)
		require.NoError(t, err)

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		ingress = &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)

		assert.True(t, kuberr.IsNotFound(err))
	})
	It("Removes HTTPRoute when .spec.route is removed", func() {
		r := NewIngressReconciler(k8sClient, false, true)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httproute-nil",
				Namespace: "default",
			},
			Spec: v1beta1.GrafanaSpec{
				HTTPRoute: &v1beta1.HTTPRouteV1{
					Spec: gwapiv1.HTTPRouteSpec{},
				},
			},
		}

		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		route := &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, route)
		require.NoError(t, err)

		cr.Spec.HTTPRoute = nil

		err = k8sClient.Update(ctx, cr)
		require.NoError(t, err)

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		route = &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, route)

		assert.True(t, kuberr.IsNotFound(err))
	})
})

var _ = Describe("GatewayAPI support", func() {
	t := GinkgoT()

	It("Creates HTTPRoute when .spec.httpRoute is defined", func() {
		r := NewIngressReconciler(k8sClient, false, true)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httproute-test",
				Namespace: "default",
				Labels:    map[string]string{"test": "httproute"},
			},
			Spec: v1beta1.GrafanaSpec{
				HTTPRoute: &v1beta1.HTTPRouteV1{},
			},
		}

		ctx := context.Background()
		err := k8sClient.Create(ctx, cr)
		require.NoError(t, err)

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		httpRoute := &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, httpRoute)
		require.NoError(t, err)
	})
})
