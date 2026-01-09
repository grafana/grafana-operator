package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("Ingress Reconciler", func() {
	t := GinkgoT()

	Context("on Openshift", func() {
		const isOpenshift = true
		const hasHTTPRouteCRD = false

		It("creates Ingress when only .spec.ingress is defined", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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

			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}

			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			ingress := &networkingv1.Ingress{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-ingress", cr.Name),
				Namespace: "default",
			}, ingress)
			require.NoError(t, err)
		})

		It("creates Route when .spec.ingress AND .spec.route are defined", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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

			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}
			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			route := &routev1.Route{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-route", cr.Name),
				Namespace: "default",
			}, route)
			require.NoError(t, err)
		})

		It("removes Route when .spec.route is removed", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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
			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}
			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			route := &routev1.Route{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-route", cr.Name),
				Namespace: "default",
			}, route)
			require.NoError(t, err)

			cr.Spec.Route = nil

			err = cl.Update(ctx, cr)
			require.NoError(t, err)

			status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			route = &routev1.Route{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-route", cr.Name),
				Namespace: "default",
			}, route)

			assert.True(t, apierrors.IsNotFound(err))
		})
	})

	Context("on Kubernetes", func() {
		const isOpenshift = false
		const hasHTTPRouteCRD = true

		It("creates Ingress when .spec.ingress is defined", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
			cr := &v1beta1.Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress-on-k8s",
					Namespace: "default",
					Labels:    map[string]string{},
				},
				Spec: v1beta1.GrafanaSpec{
					Ingress: &v1beta1.IngressNetworkingV1{},
					Route:   nil,
				},
			}

			ctx := context.Background()

			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}

			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			ingress := &networkingv1.Ingress{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-ingress", cr.Name),
				Namespace: "default",
			}, ingress)
			require.NoError(t, err)
		})

		It("removes Ingress when .spec.ingress is removed", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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
			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}
			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			ingress := &networkingv1.Ingress{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-ingress", cr.Name),
				Namespace: "default",
			}, ingress)
			require.NoError(t, err)

			cr.Spec.Ingress = nil

			err = cl.Update(ctx, cr)
			require.NoError(t, err)

			status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			ingress = &networkingv1.Ingress{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-ingress", cr.Name),
				Namespace: "default",
			}, ingress)

			assert.True(t, apierrors.IsNotFound(err))
		})

		It("creates HTTPRoute when .spec.httpRoute is defined", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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
			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}
			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			httpRoute := &gwapiv1.HTTPRoute{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-httproute", cr.Name),
				Namespace: "default",
			}, httpRoute)
			require.NoError(t, err)
		})

		It("removes HTTPRoute when .spec.httpRoute is removed", func() {
			r := NewIngressReconciler(cl, isOpenshift, hasHTTPRouteCRD)
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
			err := cl.Create(ctx, cr)
			require.NoError(t, err)

			vars := &v1beta1.OperatorReconcileVars{}
			status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			route := &gwapiv1.HTTPRoute{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-httproute", cr.Name),
				Namespace: "default",
			}, route)
			require.NoError(t, err)

			cr.Spec.HTTPRoute = nil

			err = cl.Update(ctx, cr)
			require.NoError(t, err)

			status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

			require.NoError(t, err)
			assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

			route = &gwapiv1.HTTPRoute{}
			err = cl.Get(ctx, types.NamespacedName{
				Name:      fmt.Sprintf("%s-httproute", cr.Name),
				Namespace: "default",
			}, route)

			assert.True(t, apierrors.IsNotFound(err))
		})
	})
})
