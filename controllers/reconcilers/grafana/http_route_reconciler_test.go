package grafana

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("IngressReconciler utils", Ordered, func() {
	var (
		ctx        context.Context
		reconciler *IngressReconciler
	)

	BeforeAll(func() {
		ctx = context.Background()
		reconciler = &IngressReconciler{
			client: k8sClient,
		}

		// init namespace
		namespaces := []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ns1",
					Labels: map[string]string{"label1": "value1"},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ns2",
					Labels: map[string]string{"label2": "value2"},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ns-prod",
					Labels: map[string]string{"label2": "value3"},
				},
			},
		}
		for _, ns := range namespaces {
			err := k8sClient.Create(ctx, &ns)

			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterAll(func() {
		// clean namespace
		namespaces := []string{"ns1", "ns2", "ns-prod"}
		for _, n := range namespaces {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: n}}
			err := k8sClient.Delete(ctx, ns)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("IngressReconciler getMatchListener", func() {
		hostname := gwapiv1.Hostname("test.com")

		It("matches NamespacesFromAll listener", func() {
			httpRoute := &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "route1", Namespace: "ns1"},
				Spec:       gwapiv1.HTTPRouteSpec{Hostnames: []gwapiv1.Hostname{hostname}},
			}
			gw := &gwapiv1.Gateway{
				Spec: gwapiv1.GatewaySpec{
					Listeners: []gwapiv1.Listener{
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromSame),
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromAll),
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
					},
				},
			}

			l := reconciler.getMatchListener(ctx, httpRoute, gw)
			Expect(l).NotTo(BeNil())
			Expect(*l.Hostname).To(Equal(hostname))
			Expect(*l).To(Equal(gw.Spec.Listeners[1]))
		})

		It("matches NamespacesFromSame listener", func() {
			httpRoute := &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "route2", Namespace: "ns1"},
				Spec:       gwapiv1.HTTPRouteSpec{Hostnames: []gwapiv1.Hostname{hostname}},
			}
			gw := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Namespace: "ns1"},
				Spec: gwapiv1.GatewaySpec{
					Listeners: []gwapiv1.Listener{
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromSelector),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"label2": "value2"},
									},
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromSame),
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
					},
				},
			}

			l := reconciler.getMatchListener(ctx, httpRoute, gw)
			Expect(l).NotTo(BeNil())
			Expect(*l).To(Equal(gw.Spec.Listeners[1]))
		})

		It("matches NamespacesFromSelector listener", func() {
			httpRoute := &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "route3", Namespace: "ns2"},
				Spec:       gwapiv1.HTTPRouteSpec{Hostnames: []gwapiv1.Hostname{hostname}},
			}
			gw := &gwapiv1.Gateway{
				Spec: gwapiv1.GatewaySpec{
					Listeners: []gwapiv1.Listener{
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromSelector),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"label1": "value1"},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "tier",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"frontend"},
											},
										},
									},
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromSelector),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"label2": "value2"},
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "label2",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"value2"},
											},
										},
									},
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
					},
				},
			}

			l := reconciler.getMatchListener(ctx, httpRoute, gw)
			Expect(l).NotTo(BeNil())
			Expect(*l).To(Equal(gw.Spec.Listeners[1]))
		})

		It("returns nil if no listener matches", func() {
			httpRoute := &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "route4", Namespace: "ns1"},
				Spec:       gwapiv1.HTTPRouteSpec{Hostnames: []gwapiv1.Hostname{"other.com"}},
			}
			gw := &gwapiv1.Gateway{
				Spec: gwapiv1.GatewaySpec{
					Listeners: []gwapiv1.Listener{
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: ptr(gwapiv1.NamespacesFromAll),
								},
							},
							Protocol: "http",
							Hostname: ptr(hostname),
						},
					},
				},
			}

			l := reconciler.getMatchListener(ctx, httpRoute, gw)
			Expect(l).To(BeNil())
		})
	})
})

// helper to get pointer
func ptr[T any](v T) *T { return &v }
