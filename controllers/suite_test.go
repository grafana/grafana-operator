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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	testCtx   context.Context
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

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// NOTE(Baarsgaard) Ensure k8sClient is 100% ready
	// ENVTEST sometimes fail all tests with a 401 Unauthorized
	time.Sleep(100 * time.Millisecond)

	By("Create a dummy 'invalid' instance to provoke conditions")
	intP := 1
	grafanaCr := &v1beta1.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy",
			Namespace: "default",
			Labels: map[string]string{
				"apply-failed":  "test",
				"invalid-spec":  "test",
				"loop-detected": "test",
			},
		},
		Spec: v1beta1.GrafanaSpec{
			Client: &v1beta1.GrafanaClient{TimeoutSeconds: &intP},
			Config: map[string]map[string]string{
				"security": {
					"admin_user":     "root",
					"admin_password": "secret",
				},
			},
		},
	}
	Expect(k8sClient.Create(testCtx, grafanaCr)).NotTo(HaveOccurred())

	grafanaCr.Status = v1beta1.GrafanaStatus{
		Stage:       v1beta1.OperatorStageComplete,
		StageStatus: v1beta1.OperatorStageResultSuccess,
		AdminURL:    fmt.Sprintf("http://%s-service", "invalid"),
		Version:     config.GrafanaVersion,
	}
	Expect(k8sClient.Status().Update(testCtx, grafanaCr)).ToNot(HaveOccurred())

	// Should not be reconciled
	By("Creating folder to use when provoking ApplyFailed conditions")
	folderCr := &v1beta1.GrafanaFolder{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: objectMetaApplyFailed.Namespace,
			Name:      "apply-failed-helper",
		},
		Spec: v1beta1.GrafanaFolderSpec{
			GrafanaCommonSpec: commonSpecApplyFailed,
		},
	}
	Expect(k8sClient.Create(testCtx, folderCr)).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
