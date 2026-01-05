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
	testEnv *envtest.Environment
	cfg     *rest.Config
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

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "tests", "fixtures"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error

	cfg, err = testEnv.Start()
	require.NoError(t, err)
})

var _ = AfterSuite(func() {
	t := GinkgoT()

	By("tearing down the test environment")
	err := testEnv.Stop()
	require.NoError(t, err)
})
