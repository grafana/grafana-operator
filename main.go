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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	routev1 "github.com/openshift/api/route/v1"
	discovery2 "k8s.io/client-go/discovery"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers"
	"github.com/grafana/grafana-operator/v5/controllers/autodetect"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

const (
	// watchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE which specifies the Namespace to watch.
	// If empty or undefined, the operator will run in cluster scope.
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
	// watchNamespaceEnvSelector is the constant for env variable WATCH_NAMESPACE_SELECTOR which specifies the Namespace label and key to watch.
	// eg: "environment: dev"
	// If empty or undefined, the operator will run in cluster scope.
	watchNamespaceEnvSelector = "WATCH_NAMESPACE_SELECTOR"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(grafanav1beta1.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	watchNamespace, _ := os.LookupEnv(watchNamespaceEnvVar)
	watchNamespaceSelector, _ := os.LookupEnv(watchNamespaceEnvSelector)

	controllerOptions := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f75f3bba.integreatly.org",
	}

	getNamespaceConfig := func(namespaces string) map[string]cache.Config {
		defaultNamespaces := map[string]cache.Config{}
		for _, v := range strings.Split(namespaces, ",") {
			// Generate a mapping of namespaces to label/field selectors, set to Everything() to enable matching all
			// instances in all namespaces from watchNamespace to be controlled by the operator
			// this is the default behavior of the operator on v5, if you require finer grained control over this
			// please file an issue in the grafana-operator/grafana-operator GH project
			defaultNamespaces[v] = cache.Config{
				LabelSelector:         labels.Everything(), // Match any labels
				FieldSelector:         fields.Everything(), // Match any fields
				Transform:             nil,
				UnsafeDisableDeepCopy: nil,
			}
		}
		return defaultNamespaces
	}
	getNamespaceConfigSelector := func(selector string) map[string]cache.Config {
		cl, err := client.New(config.GetConfigOrDie(), client.Options{})
		if err != nil {
			setupLog.Error(err, "Failed to get watch namespaces")
		}
		nsList := &corev1.NamespaceList{}
		listOpts := []client.ListOption{
			client.MatchingLabels(map[string]string{strings.Split(selector, ":")[0]: strings.Split(selector, ":")[1]}),
		}
		err = cl.List(context.Background(), nsList, listOpts...)
		if err != nil {
			setupLog.Error(err, "Failed to get watch namespaces")
		}
		defaultNamespaces := map[string]cache.Config{}
		for _, v := range nsList.Items {
			// Generate a mapping of namespaces to label/field selectors, set to Everything() to enable matching all
			// instances in all namespaces from watchNamespace to be controlled by the operator
			// this is the default behavior of the operator on v5, if you require finer grained control over this
			// please file an issue in the grafana-operator/grafana-operator GH project
			defaultNamespaces[v.Name] = cache.Config{
				LabelSelector:         labels.Everything(), // Match any labels
				FieldSelector:         fields.Everything(), // Match any fields
				Transform:             nil,
				UnsafeDisableDeepCopy: nil,
			}
		}
		return defaultNamespaces
	}
	switch {
	case strings.Contains(watchNamespace, ","):
		// multi namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace)
		setupLog.Info("manager set up with multiple namespaces", "namespaces", watchNamespace)
	case watchNamespace != "":
		// namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace)
		setupLog.Info("operator running in namespace scoped mode", "namespace", watchNamespace)
	case strings.Contains(watchNamespaceSelector, ":"):
		// namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfigSelector(watchNamespaceSelector)
		setupLog.Info("operator running in namespace scoped mode using namespace selector", "namespace", watchNamespace)

	case watchNamespace == "" && watchNamespaceSelector == "":
		// cluster scoped
		setupLog.Info("operator running in cluster scoped mode")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), controllerOptions)
	if err != nil {
		setupLog.Error(err, "unable to create new manager")
		os.Exit(1) //nolint
	}

	restConfig := ctrl.GetConfigOrDie()
	autodetect, err := autodetect.New(restConfig)
	if err != nil {
		setupLog.Error(err, "failed to setup auto-detect routine")
		os.Exit(1)
	}
	isOpenShift, err := autodetect.IsOpenshift()
	if err != nil {
		setupLog.Error(err, "unable to detect the platform")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaReconciler{
		Client:      mgr.GetClient(),
		Scheme:      mgr.GetScheme(),
		IsOpenShift: isOpenShift,
		Discovery:   discovery2.NewDiscoveryClientForConfigOrDie(ctrl.GetConfigOrDie()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Grafana")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaDashboardReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("DashboardReconciler"),
	}).SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaDatasourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("DatasourceReconciler"),
	}).SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDatasource")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaFolderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaFolder")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaAlertRuleGroupReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaAlertRuleGroup")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaContactPointReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaContactPoint")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	<-ctx.Done()
	setupLog.Info("SIGTERM request gotten, shutting down operator")
}
