package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("Allow use of Ingress on OpenShift", func() {
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route := &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)
		Expect(err).ToNot(HaveOccurred())
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route := &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)
		Expect(err).ToNot(HaveOccurred())

		cr.Spec.Route = nil

		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route = &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)

		Expect(kuberr.IsNotFound(err)).To(BeTrue())
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

		cr.Spec.Ingress = nil

		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		ingress = &networkingv1.Ingress{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-ingress", cr.Name),
			Namespace: "default",
		}, ingress)

		Expect(kuberr.IsNotFound(err)).To(BeTrue())
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route := &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, route)
		Expect(err).ToNot(HaveOccurred())

		cr.Spec.HTTPRoute = nil

		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route = &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, route)

		Expect(kuberr.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("GatewayAPI support", func() {
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
		Expect(k8sClient.Create(ctx, cr)).To(Succeed())

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		httpRoute := &gwapiv1.HTTPRoute{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-httproute", cr.Name),
			Namespace: "default",
		}, httpRoute)
		Expect(err).ToNot(HaveOccurred())
	})
})
