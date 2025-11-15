package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))

		route := &routev1.Route{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-route", cr.Name),
			Namespace: "default",
		}, route)
		Expect(err).ToNot(HaveOccurred())
	})
})
