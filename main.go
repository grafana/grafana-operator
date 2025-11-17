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
	"crypto/sha256"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	uberzap "go.uber.org/zap"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"

	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers"
	"github.com/grafana/grafana-operator/v5/controllers/autodetect"
	"github.com/grafana/grafana-operator/v5/controllers/model"
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
	// Opt in to enable new experimental cache limits by setting this to `safe` or `all`. Valid values are `off`, `safe` and `all`
	enforceCacheLabelsEnvVar = "ENFORCE_CACHE_LABELS"
	cachingLevelAll          = "all"
	cachingLevelOff          = "off"
	cachingLevelSafe         = "safe"
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

	utilruntime.Must(gwapiv1.Install(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() { //nolint:gocyclo
	var (
		metricsAddr             string
		enableLeaderElection    bool
		probeAddr               string
		pprofAddr               string
		maxConcurrentReconciles int
		resyncPeriod            time.Duration
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&pprofAddr, "pprof-addr", ":8888", "The address to expose the pprof server. Empty string disables the pprof server.")
	flag.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 1,
		"Maximum number of concurrent reconciles for dashboard, datasource, folder controllers.")
	flag.DurationVar(&resyncPeriod, "default-resync-period", controllers.DefaultReSyncPeriod, "Controls the default .spec.resyncPeriod when undefined on CRs.")

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

	// Detect environment variables
	watchNamespace, _ := os.LookupEnv(watchNamespaceEnvVar)
	watchNamespaceSelector, _ := os.LookupEnv(watchNamespaceEnvSelector)

	watchLabelSelectors, _ := os.LookupEnv(watchLabelSelectorsEnvVar)
	if watchLabelSelectors != "" {
		setupLog.Info(fmt.Sprintf("sharding is enabled via %s=%s. Beware: Always label Grafana CRs before enabling to ensure labels are inherited. Existing Secrets/ConfigMaps referenced in CRs also need to be labeled to continue working.", watchLabelSelectorsEnvVar, watchLabelSelectors))
	}

	enforceCacheLabelsLevel, _ := os.LookupEnv(enforceCacheLabelsEnvVar)
	if enforceCacheLabelsLevel == "" {
		enforceCacheLabelsLevel = cachingLevelSafe
	}

	enforceCacheLabels := false

	switch enforceCacheLabelsLevel {
	case cachingLevelSafe, cachingLevelAll:
		enforceCacheLabels = true

		setupLog.Info("label restrictions for cached resources are active", "level", enforceCacheLabelsLevel)
	case cachingLevelOff:
	default:
		setupLog.Error(fmt.Errorf("invalid value %s for %s", enforceCacheLabelsLevel, enforceCacheLabelsEnvVar), "falling back to disabling cache enforcement")
	}

	// Determine LeaderElectionID from
	leHash := sha256.New()
	leHash.Write([]byte(watchNamespace))
	leHash.Write([]byte(watchNamespaceSelector))
	leHash.Write([]byte(watchLabelSelectors))

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

	hasGatewayAPI, err := autodetect.HasGatewayAPI()
	if err != nil {
		setupLog.Error(err, "failed to test for GatewayAPI CRDs")
		os.Exit(1)
	}

	mgrOptions := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       fmt.Sprintf("grafana-operator-%x", leHash.Sum(nil)),
		PprofBindAddress:       pprofAddr,
		Controller: config.Controller{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		},
	}

	labelSelectors, err := getLabelSelectors(watchLabelSelectors)
	if err != nil {
		setupLog.Error(err, fmt.Sprintf("unable to parse %s", watchLabelSelectorsEnvVar))
		os.Exit(1)
	}

	if enforceCacheLabels {
		var cacheLabelConfig cache.ByObject
		if watchLabelSelectors != "" {
			// When sharding, limit cache according to shard labels
			cacheLabelConfig = cache.ByObject{Label: labelSelectors}

			setupLog.Info(fmt.Sprintf("sharding is enabled via %s=%s. Beware: Always label Grafana CRs before enabling to ensure labels are inherited. Existing Secrets/ConfigMaps referenced in CRs also need to be labeled to continue working.", watchLabelSelectorsEnvVar, watchLabelSelectors))
		} else {
			// Otherwise limit it to managed-by label
			cacheLabelConfig = cache.ByObject{Label: labels.SelectorFromSet(model.GetCommonLabels())}
		}

		// ConfigMaps and secrets stay fully cached until we implement support for bypassing the cache for referenced objects
		mgrOptions.Cache.ByObject = map[client.Object]cache.ByObject{
			&v1.Deployment{}:                cacheLabelConfig,
			&corev1.Service{}:               cacheLabelConfig,
			&corev1.ServiceAccount{}:        cacheLabelConfig,
			&networkingv1.Ingress{}:         cacheLabelConfig,
			&corev1.PersistentVolumeClaim{}: cacheLabelConfig,
			&corev1.ConfigMap{}:             cacheLabelConfig, // Matching just labeled ConfigMaps and Secrets greatly reduces cache size
			&corev1.Secret{}:                cacheLabelConfig, // Omitting labels or supporting custom labels would require changes in Grafana Reconciler
		}
		if isOpenShift {
			mgrOptions.Cache.ByObject[&routev1.Route{}] = cacheLabelConfig
		}

		if hasGatewayAPI {
			mgrOptions.Cache.ByObject[&gwapiv1.HTTPRoute{}] = cacheLabelConfig
		} else {
			setupLog.Info("skipping cache fine tuning for HTTPRoute resources as GatewayAPI CRDs were not found in the cluster")
		}

		if enforceCacheLabelsLevel == cachingLevelSafe {
			mgrOptions.Client.Cache = &client.CacheOptions{
				DisableFor: []client.Object{&corev1.ConfigMap{}, &corev1.Secret{}},
			}
		}
	}

	// Determine Operator scope
	switch {
	case strings.Contains(watchNamespace, ","):
		// multi namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode for multiple namespaces", "namespaces", watchNamespace)
	case watchNamespace != "":
		// namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfig(watchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode", "namespace", watchNamespace)
	case strings.Contains(watchNamespaceSelector, ":"):
		// multi namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfigSelector(restConfig, watchNamespaceSelector, labelSelectors)

		setupLog.Info("operator running in namespace scoped mode using namespace selector", "namespace", watchNamespace)

	case watchNamespace == "" && watchNamespaceSelector == "":
		// cluster scoped
		mgrOptions.Cache.DefaultLabelSelector = labelSelectors

		setupLog.Info("operator running in cluster scoped mode")
	}

	ctx := ctrl.SetupSignalHandler()

	mgr, err := ctrl.NewManager(restConfig, mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to create new manager")
		os.Exit(1)
	}

	ctrlCfg := &controllers.Config{
		ResyncPeriod: resyncPeriod,
	}
	// Register controllers
	if err = (&controllers.GrafanaReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		IsOpenShift:   isOpenShift,
		HasGatewayAPI: hasGatewayAPI,
		ClusterDomain: clusterDomain,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Grafana")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaDashboardReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaDatasourceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDatasource")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaServiceAccountReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaServiceAccount")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaFolderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaFolder")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaLibraryPanelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaLibraryPanel")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaAlertRuleGroupReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaAlertRuleGroup")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaContactPointReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaContactPoint")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaNotificationPolicyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("GrafanaNotificationPolicy"),
		Cfg:      ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaNotificationPolicy")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaNotificationTemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaNotificationTemplate")
		os.Exit(1)
	}

	if err = (&controllers.GrafanaMuteTimingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
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

	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	<-ctx.Done()
	setupLog.Info("SIGTERM request gotten, shutting down operator")
}

func getNamespaceConfig(namespaces string, labelSelectors labels.Selector) map[string]cache.Config {
	defaultNamespaces := map[string]cache.Config{}
	for v := range strings.SplitSeq(namespaces, ",") {
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

	managedByLabelSelector, _ := labels.SelectorFromSet(model.GetCommonLabels()).Requirements()
	labelSelectors.Add(managedByLabelSelector...)

	return labelSelectors, nil
}
