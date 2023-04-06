/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

var _ = Describe("Grafana controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		GrafanaName      = "test-grafana"
		GrafanaNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)
	replicas := int32(2)

	Context("When creating Grafana", func() {
		It("Should create expected resources", func() {
			By("By creating a new Grafana resource")
			ctx := context.Background()
			grafana := &v1beta1.Grafana{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "grafana.integreatly.org/v1beta1",
					Kind:       "Grafana",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      GrafanaName,
					Namespace: GrafanaNamespace,
				},
				Spec: v1beta1.GrafanaSpec{
					Deployment: &v1beta1.DeploymentV1{
						Spec: v1beta1.DeploymentV1Spec{
							Replicas: &replicas,
						},
					},
					PersistentVolumeClaim: &v1beta1.PersistentVolumeClaimV1{
						Spec: &v1beta1.PersistentVolumeClaimV1Spec{
							AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
							Resources: &v1.ResourceRequirements{Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("100Mi"),
							}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, grafana)).Should(Succeed())
			grafanaLookupKey := types.NamespacedName{Name: GrafanaName, Namespace: GrafanaNamespace}
			createdGrafana := &v1beta1.Grafana{}

			By("By becoming ready")
			Eventually(func() *metav1.Condition {
				err := k8sClient.Get(ctx, grafanaLookupKey, createdGrafana)
				if err != nil {
					return nil
				}
				return createdGrafana.GetReadyCondition()
			}, timeout, interval).Should(Equal(metav1.Condition{
				Type:    "Ready",
				Message: "grafana api availabel",
				Reason:  v1beta1.GrafanaApiAvailableReason,
			}))

			By("By checking for the created resources")
			configLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-ini", GrafanaName), Namespace: GrafanaNamespace}
			createdConfig := &v1.ConfigMap{}

			secretLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-admin-credentials", GrafanaName), Namespace: GrafanaNamespace}
			createdSecret := &v1.Secret{}

			pvcLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-pvc", GrafanaName), Namespace: GrafanaNamespace}
			createdPvc := &v1.PersistentVolumeClaim{}

			saLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-sa", GrafanaName), Namespace: GrafanaNamespace}
			createdSa := &v1.ServiceAccount{}

			serviceLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-service", GrafanaName), Namespace: GrafanaNamespace}
			createdService := &v1.Service{}

			ingressLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-ingress", GrafanaName), Namespace: GrafanaNamespace}
			createdIngress := &networkingv1.Ingress{}

			deploymentLookupKey := types.NamespacedName{Name: fmt.Sprintf("%s-deployment", GrafanaName), Namespace: GrafanaNamespace}
			createdDeployment := &appsv1.Deployment{}

			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, configLookupKey, createdConfig)
				if err != nil {
					return false, fmt.Errorf("failed config: %w", err)
				}
				err = k8sClient.Get(ctx, secretLookupKey, createdSecret)
				if err != nil {
					return false, fmt.Errorf("failed secret: %w", err)
				}
				err = k8sClient.Get(ctx, pvcLookupKey, createdPvc)
				if err != nil {
					return false, fmt.Errorf("failed pvc: %w", err)
				}
				err = k8sClient.Get(ctx, saLookupKey, createdSa)
				if err != nil {
					return false, fmt.Errorf("failed sa: %w", err)
				}
				err = k8sClient.Get(ctx, serviceLookupKey, createdService)
				if err != nil {
					return false, fmt.Errorf("failed service: %w", err)
				}
				err = k8sClient.Get(ctx, ingressLookupKey, createdIngress)
				if err != nil {
					return false, fmt.Errorf("failed ingress: %w", err)
				}
				err = k8sClient.Get(ctx, deploymentLookupKey, createdDeployment)
				if err != nil {
					return false, fmt.Errorf("failed deployment: %w", err)
				}
				return true, nil
			}, timeout, interval).Should(BeTrue())

			By("By setting the Grafana.Status.PluginList")
			createdGrafana.Status.Plugins = v1beta1.PluginList{
				v1beta1.GrafanaPlugin{
					Name:    "grafana-piechart-panel",
					Version: "1.6.1",
				},
			}
			Expect(k8sClient.Status().Update(ctx, createdGrafana)).Should(Succeed())

			By("By checking that the Deployment has the expected plugin")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, deploymentLookupKey, createdDeployment)
				if err != nil {
					return "", err
				}

				for _, container := range createdDeployment.Spec.Template.Spec.Containers {
					for _, env := range container.Env {
						if env.Name == "GF_INSTALL_PLUGINS" {
							return env.Value, nil
						}
					}
				}
				return "", fmt.Errorf("Missing GF_INSTALL_PLUGINS")
			}, timeout, interval).Should(Equal(createdGrafana.Status.Plugins.String()))
		})
	})

})
