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
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	discovery2 "k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/grafana/grafana-operator/v5/controllers/model"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers"
	"github.com/grafana/grafana-operator/v5/controllers/autodetect"
	"github.com/grafana/grafana-operator/v5/embeds"
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
	// watchLabelSelectorsEnvVar is the constant for env variable WATCH_LABEL_SELECTORS which specifies the resources to watch according to their labels.
	// eg: 'partition in (customerA, customerB),environment!=qa'
	// If empty of undefined, the operator will watch all CRs.
	watchLabelSelectorsEnvVar = "WATCH_LABEL_SELECTORS"
	// clusterDomainEnvVar is the constant for env variable CLUSTER_DOMAIN, which specifies the cluster domain to use for addressing.
	// By default, this is empty, and internal services are addressed without a cluster domain specified, i.e., a
	// relative domain name that will resolve regardless of if a custom domain is configured for the cluster. If you
	// wish to have services addressed using their FQDNs, you can specify the cluster domain explicitly, e.g., "cluster.local"
	// for the default Kubernetes configuration.
	clusterDomainEnvVar = "CLUSTER_DOMAIN"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup").WithValues("version", embeds.Version)
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
	var pprofAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&pprofAddr, "pprof-addr", ":8888", "The address to expose the pprof server. Empty string disables the pprof server.")

	logCfg := uberzap.NewProductionEncoderConfig()
	logCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	opts := zap.Options{
		NewEncoder: func(eco ...zap.EncoderConfigOption) zapcore.Encoder {
			return zapcore.NewConsoleEncoder(logCfg)
		},
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	slogger := slog.New(logr.ToSlogHandler(setupLog))
	slog.SetDefault(slogger)

	// Optimize Go runtime based on CGroup limits (GOMEMLIMIT, sets a soft memory limit for the runtime)
	memlimit.SetGoMemLimitWithOpts(memlimit.WithLogger(slogger)) //nolint:errcheck

	// Optimize Go runtime based on CGroup limits (GOMAXPROCS, limits the number of operating system threads that can execute user-level Go code simultaneously)
	_, err := maxprocs.Set(maxprocs.Logger(log.Printf))
	if err != nil {
		setupLog.Error(err, "failed to adjust GOMAXPROCS")
	}

	watchNamespace, _ := os.LookupEnv(watchNamespaceEnvVar)
	watchNamespaceSelector, _ := os.LookupEnv(watchNamespaceEnvSelector)
	watchLabelSelectors, _ := os.LookupEnv(watchLabelSelectorsEnvVar)
	if watchLabelSelectors != "" {
		setupLog.Info(fmt.Sprintf("sharding is enabled via %s=%s. Beware: Always label Grafana CRs before enabling to ensure labels are inherited. Existing Secrets/ConfigMaps referenced in CRs also need to be labeled to continue working.", watchLabelSelectorsEnvVar, watchLabelSelectors))
	}
	clusterDomain, _ := os.LookupEnv(clusterDomainEnvVar)

	// Fetch k8s api credentials and detect platform
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
		PprofBindAddress:       pprofAddr,
	}

	labelSelectors, err := getLabelSelectors(watchLabelSelectors)
	if err != nil {
		setupLog.Error(err, fmt.Sprintf("unable to parse %s", watchLabelSelectorsEnvVar))
		os.Exit(1) //nolint
	}
	switch {
	case strings.Contains(watchNamespace, ","):
		// multi namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode for multiple namespaces", "namespaces", watchNamespace)
	case watchNamespace != "":
		// namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode", "namespace", watchNamespace)
	case strings.Contains(watchNamespaceSelector, ":"):
		// namespace scoped
		controllerOptions.Cache.DefaultNamespaces = getNamespaceConfigSelector(restConfig, watchNamespaceSelector, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode using namespace selector", "namespace", watchNamespace)

	case watchNamespace == "" && watchNamespaceSelector == "":
		// cluster scoped
		controllerOptions.Cache.DefaultLabelSelector = labelSelectors
		setupLog.Info("operator running in cluster scoped mode")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	mgr, err := ctrl.NewManager(restConfig, controllerOptions)
	if err != nil {
		setupLog.Error(err, "unable to create new manager")
		os.Exit(1) //nolint
	}

	if err = (&controllers.GrafanaReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		IsOpenShift:   isOpenShift,
		Discovery:     discovery2.NewDiscoveryClientForConfigOrDie(restConfig),
		ClusterDomain: clusterDomain,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Grafana")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaDashboardReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaDatasourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
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
	if err = (&controllers.GrafanaLibraryPanelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaLibraryPanel")
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
	if err = (&controllers.GrafanaNotificationPolicyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("GrafanaNotificationPolicy"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaNotificationPolicy")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaNotificationTemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaNotificationTemplate")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaMuteTimingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaMuteTiming")
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

func getNamespaceConfig(namespaces string, labelSelectors labels.Selector) map[string]cache.Config {
	defaultNamespaces := map[string]cache.Config{}
	for _, v := range strings.Split(namespaces, ",") {
		// Generate a mapping of namespaces to label/field selectors, set to Everything() to enable matching all
		// instances in all namespaces from watchNamespace to be controlled by the operator
		// this is the default behavior of the operator on v5, if you require finer grained control over this
		// please file an issue in the grafana-operator/grafana-operator GH project
		defaultNamespaces[v] = cache.Config{
			LabelSelector:         labelSelectors,
			FieldSelector:         fields.Everything(), // Match any fields
			Transform:             nil,
			UnsafeDisableDeepCopy: nil,
		}
	}
	return defaultNamespaces
}

func getNamespaceConfigSelector(restConfig *rest.Config, selector string, labelSelectors labels.Selector) map[string]cache.Config {
	cl, err := client.New(restConfig, client.Options{})
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
			LabelSelector:         labelSelectors,
			FieldSelector:         fields.Everything(), // Match any fields
			Transform:             nil,
			UnsafeDisableDeepCopy: nil,
		}
	}
	return defaultNamespaces
}

func getLabelSelectors(watchLabelSelectors string) (labels.Selector, error) {
	var (
		labelSelectors labels.Selector
		err            error
	)
	if watchLabelSelectors != "" {
		labelSelectors, err = labels.Parse(watchLabelSelectors)
		if err != nil {
			return labelSelectors, fmt.Errorf("unable to parse %s: %w", watchLabelSelectorsEnvVar, err)
		}
	} else {
		labelSelectors = labels.Everything() // Match any labels
	}
	managedByLabelSelector, _ := labels.SelectorFromSet(model.CommonLabels).Requirements()
	labelSelectors.Add(managedByLabelSelector...)
	return labelSelectors, nil
}
