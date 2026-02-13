package autodetect

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfgNoCRDs       *rest.Config
	cfgWithCRDs     *rest.Config
	testEnvNoCRDs   *envtest.Environment
	testEnvWithCRDs *envtest.Environment
)

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("-short was passed, skipping Autodetect")
	}

	RunSpecs(t, "Autodetect Suite")
}

var _ = BeforeSuite(func() {
	t := GinkgoT()

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error

	By("bootstrapping test environment")

	testEnvNoCRDs = &envtest.Environment{}

	testEnvWithCRDs = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "tests", "fixtures"),
		},
		ErrorIfCRDPathMissing: true,
	}

	cfgNoCRDs, err = testEnvNoCRDs.Start()
	require.NoError(t, err)

	cfgWithCRDs, err = testEnvWithCRDs.Start()
	require.NoError(t, err)
})

var _ = AfterSuite(func() {
	t := GinkgoT()

	By("tearing down the test environment")

	err := testEnvNoCRDs.Stop()
	require.NoError(t, err)

	err = testEnvWithCRDs.Stop()
	require.NoError(t, err)
})
