package fetchers

import (
	"testing"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("-short was passed, skipping Fetchers")
	}

	RegisterFailHandler(Fail)

	RunSpecs(t, "Fetchers Suite")
}

var _ = BeforeSuite(func() {
	t := GinkgoT()

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	err = grafanav1beta1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)
})

var _ = AfterSuite(func() {
	t := GinkgoT()

	By("tearing down the test environment")
	err := testEnv.Stop()
	require.NoError(t, err)
})
