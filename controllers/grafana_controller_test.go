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
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	grec "github.com/grafana-operator/grafana-operator/v5/controllers/reconcilers/grafana"
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
		It("Should seem to work", func() {
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
					Ingress: &v1beta1.IngressNetworkingV1{
						Spec: &networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "example.com",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, grafana)).Should(Succeed())
			grafanaLookupKey := types.NamespacedName{Name: GrafanaName, Namespace: GrafanaNamespace}
			createdGrafana := &v1beta1.Grafana{}

			By("By checking for the created resources")
			configLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaIniMeta(grafana))
			createdConfig := &v1.ConfigMap{}

			secretLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaAdminSecretMeta(grafana))
			createdSecret := &v1.Secret{}

			pvcLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaDataPVCMeta(grafana))
			createdPvc := &v1.PersistentVolumeClaim{}

			saLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaServiceAccountMeta(grafana))
			createdSa := &v1.ServiceAccount{}

			serviceLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaServiceMeta(grafana))
			createdService := &v1.Service{}

			ingressLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaIngressMeta(grafana))
			createdIngress := &networkingv1.Ingress{}

			deploymentLookupKey := client.ObjectKeyFromObject(grec.GetGrafanaDeploymentMeta(grafana))
			createdDeployment := &appsv1.Deployment{}

			Eventually(func() (bool, error) {
				if err := k8sClient.Get(ctx, grafanaLookupKey, createdGrafana); err != nil {
					return false, fmt.Errorf("failed grafana: %w", err)
				}
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false, fmt.Errorf("failed config: %w", err)
				}
				if err := k8sClient.Get(ctx, secretLookupKey, createdSecret); err != nil {
					return false, fmt.Errorf("failed secret: %w", err)
				}
				if err := k8sClient.Get(ctx, pvcLookupKey, createdPvc); err != nil {
					return false, fmt.Errorf("failed pvc: %w", err)
				}
				if err := k8sClient.Get(ctx, saLookupKey, createdSa); err != nil {
					return false, fmt.Errorf("failed sa: %w", err)
				}
				if err := k8sClient.Get(ctx, serviceLookupKey, createdService); err != nil {
					return false, fmt.Errorf("failed service: %w", err)
				}
				if err := k8sClient.Get(ctx, ingressLookupKey, createdIngress); err != nil {
					return false, fmt.Errorf("failed ingress: %w", err)
				}
				if err := k8sClient.Get(ctx, deploymentLookupKey, createdDeployment); err != nil {
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
						if env.Name == config.GrafanaPluginsEnvVar {
							return env.Value, nil
						}
					}
				}
				return "", fmt.Errorf("Missing %s", config.GrafanaPluginsEnvVar)
			}, timeout, interval).Should(Equal(createdGrafana.Status.Plugins.String()))

			By("By having a status condition")
			Eventually(func() *metav1.Condition {
				err := k8sClient.Get(ctx, grafanaLookupKey, createdGrafana)
				if err != nil {
					return nil
				}
				return createdGrafana.GetReadyCondition()
			}, timeout, interval).Should(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal("Ready"),
				"Status": Equal(metav1.ConditionFalse),
				"Reason": Equal("GrafanaApiUnavailableFailed"),
			})))
		})
	})
})
