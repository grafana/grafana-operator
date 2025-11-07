package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("Allow use of Ingress on OpenShift", func() {
	It("Creates Ingress on OpenShift when .spec.ingress is defined", func() {
		r := NewIngressReconciler(k8sClient, true)
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		ingress := &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Creates Route on OpenShift when .spec.ingress AND .spec.route is defined", func() {
		r := NewIngressReconciler(k8sClient, true)
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).To(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultFailed), "Route does not exist in Scheme outside of OpenShift")
	})
})

var _ = Describe("HTTPRoute support", func() {
	It("Creates HTTPRoute when .spec.httpRoute is defined", func() {
		r := NewIngressReconciler(k8sClient, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "httproute-test",
				Namespace: "default",
				Labels:    map[string]string{"test": "httproute"},
			},
			Spec: v1beta1.GrafanaSpec{
				HTTPRoute: &v1beta1.HTTPRouteGatewayV1{},
			},
		}

		ctx := context.Background()
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		httpRoute := &gatewayv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, httpRoute)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Creates both Ingress and HTTPRoute when both are defined", func() {
		r := NewIngressReconciler(k8sClient, false)
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-and-httproute",
				Namespace: "default",
				Labels:    map[string]string{"test": "both"},
			},
			Spec: v1beta1.GrafanaSpec{
				Ingress:   &v1beta1.IngressNetworkingV1{},
				HTTPRoute: &v1beta1.HTTPRouteGatewayV1{},
			},
		}

		ctx := context.Background()
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		// Check Ingress was created
		ingress := &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)
		Expect(err).ToNot(HaveOccurred())

		// Check HTTPRoute was created
		httpRoute := &gatewayv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, httpRoute)
		Expect(err).ToNot(HaveOccurred())
	})
})
