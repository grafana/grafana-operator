package e2e

import (
	"net/http"
	"os"
	"path"
	"testing"

	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
)

func TestGrafana(t *testing.T) {
	grafanaList := &grafanav1alpha1.GrafanaList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, grafanaList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	ctx := framework.NewContext(t)
	defer ctx.Cleanup()

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global

	// run subtests
	t.Run("grafana", func(t *testing.T) {
		t.Log(f.KubeConfig.BearerToken)
	})
}
