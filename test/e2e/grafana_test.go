package e2e

import (
	"testing"
	"time"

	"github.com/integr8ly/grafana-operator/v3/pkg/apis"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goctx "context"
)

const (
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
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

	// get namespace
	namespaceName, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	err = f.Client.Create(goctx.TODO(), namespace, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 30, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	// create grafana custom resource
	exampleGrafana := &grafanav1alpha1.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-grafana",
			Namespace: namespaceName,
		},
		Spec: grafanav1alpha1.GrafanaSpec{
			Ingress: &grafanav1alpha1.GrafanaIngress{
				Enabled: true,
			},
		},
	}
	err = f.Client.Create(goctx.TODO(), exampleGrafana, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 30, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-grafana to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespaceName, "grafana-deployment", 1, time.Second*5, time.Second*60)
	if err != nil {
		t.Fatal(err)
	}
}
