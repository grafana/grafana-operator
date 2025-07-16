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
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/docker/go-connections/nat"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	grafanaName = "external-grafana"
	grafanaUser = "root"
	grafanaPass = "secret"
)

var (
	k8sClient         client.Client
	testEnv           *envtest.Environment
	testCtx           context.Context
	grafanaContainer  testcontainers.Container
	externalGrafanaCr *v1beta1.Grafana

	grafanaPort        = nat.Port(fmt.Sprint(config.GrafanaHTTPPort)) //nolint
	grafanaCredentials = map[string]string{
		"GF_SECURITY_ADMIN_USER":     grafanaUser,
		"GF_SECURITY_ADMIN_PASSWORD": grafanaPass,
	}
)

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("-short was passed, skipping Controllers")
	}

	RegisterFailHandler(Fail)

	RunSpecs(t, "Controllers Suite")
}

var _ = BeforeSuite(func() {
	testCtx = context.Background()

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx := context.Background()
	log := logf.FromContext(ctx).WithName("ControllerTests")
	testCtx = logf.IntoContext(ctx, log)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	Expect(v1beta1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	By("Instantiating k8sClient")
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Starting Grafana TestContainer")
	grafanaContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         fmt.Sprintf("%s-%d", grafanaName, GinkgoRandomSeed()),
			Image:        fmt.Sprintf("%s:%s", config.GrafanaImage, config.GrafanaVersion),
			ExposedPorts: []string{grafanaPort.Port()},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort(grafanaPort),
				wait.ForHTTP("/api/frontend/settings").
					WithPort(grafanaPort).
					WithBasicAuth(grafanaUser, grafanaPass).
					WithStartupTimeout(8*time.Second),
			),
			Env: grafanaCredentials,
		},
	})
	Expect(err).NotTo(HaveOccurred())

	port, err := grafanaContainer.MappedPort(testCtx, grafanaPort)
	Expect(err).NotTo(HaveOccurred())

	createSharedTestCRs(port.Port())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	testcontainers.CleanupContainer(GinkgoTB(), grafanaContainer)
	Expect(testEnv.Stop()).To(Succeed())
})

func createSharedTestCRs(port string) {
	GinkgoHelper()

	By("Creating Configmaps and GrafanaFolder for testing")
	secretCR := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "external-credentials",
		},
		StringData: grafanaCredentials,
	}
	Expect(k8sClient.Create(testCtx, secretCR)).ToNot(HaveOccurred())
	folderCR := &v1beta1.GrafanaFolder{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "apply-failed-helper",
		},
		Spec: v1beta1.GrafanaFolderSpec{
			GrafanaCommonSpec: commonSpecApplyFailed,
		},
	}
	Expect(k8sClient.Create(testCtx, folderCR)).ToNot(HaveOccurred())

	By("Creating Grafana CRs. One Fake and one External")
	intP := 1
	dummy := &v1beta1.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "dummy",
			Labels: map[string]string{
				"apply-failed":  "test",
				"invalid-spec":  "test",
				"loop-detected": "test",
			},
		},
		Spec: v1beta1.GrafanaSpec{
			Client: &v1beta1.GrafanaClient{TimeoutSeconds: &intP},
		},
	}
	external := &v1beta1.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      grafanaName,
			Labels: map[string]string{
				"synchronized":       "test",
				"matching-instances": "test",
				"dashboards":         "grafana",
			},
		},
		Spec: v1beta1.GrafanaSpec{
			External: &v1beta1.External{
				URL: fmt.Sprintf("http://localhost:%s", port),
				AdminUser: &corev1.SecretKeySelector{
					Key: "GF_SECURITY_ADMIN_USER",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretCR.Name,
					},
				},
				AdminPassword: &corev1.SecretKeySelector{
					Key: "GF_SECURITY_ADMIN_PASSWORD",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretCR.Name,
					},
				},
			},
			Client: &v1beta1.GrafanaClient{TimeoutSeconds: &intP},
		},
	}
	Expect(k8sClient.Create(testCtx, dummy)).Should(Succeed())
	Expect(k8sClient.Create(testCtx, external)).Should(Succeed())

	dummy.Status = v1beta1.GrafanaStatus{
		Stage:       v1beta1.OperatorStageComplete,
		StageStatus: v1beta1.OperatorStageResultSuccess,
		AdminURL:    fmt.Sprintf("http://%s-service", "invalid"),
		Version:     config.GrafanaVersion,
	}
	Expect(k8sClient.Status().Update(testCtx, dummy)).ToNot(HaveOccurred())

	By("Reconciling External Grafana")
	r := GrafanaReconciler{
		Client:      k8sClient,
		Scheme:      k8sClient.Scheme(),
		IsOpenShift: false,
	}
	reg := requestFromMeta(external.ObjectMeta)
	_, err := r.Reconcile(testCtx, reg)
	Expect(err).ToNot(HaveOccurred())

	By("Get External Grafana")
	externalGrafanaCr = &v1beta1.Grafana{}
	Expect(k8sClient.Get(testCtx, types.NamespacedName{
		Namespace: external.Namespace,
		Name:      external.Name,
	}, externalGrafanaCr)).Should(Succeed())
}
