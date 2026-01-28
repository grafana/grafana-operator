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
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/alecthomas/kong"
	"github.com/go-logr/logr"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
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

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers"
	"github.com/grafana/grafana-operator/v5/controllers/resources"
	"github.com/grafana/grafana-operator/v5/embeds"
	"github.com/grafana/grafana-operator/v5/pkg/autodetect"
	//+kubebuilder:scaffold:imports
)

const (
	// Caching levels enforced by Kong input validation
	cachingLevelAll  = "all"
	cachingLevelOff  = "off"
	cachingLevelSafe = "safe"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup").WithValues("version", embeds.Version)
)

var operatorConfig struct {
	ClusterDomain          string `env:"CLUSTER_DOMAIN"                                              help:"Fully specify the domain to address services with using their FQDNs, e.g. 'cluster.local'"`
	WatchNamespace         string `env:"WATCH_NAMESPACE"                                             help:"Comma separated Namespaces to watch, If empty or undefined, the operator will run in cluster scope."`
	WatchNamespaceSelector string `env:"WATCH_NAMESPACE_SELECTOR"                                    help:"The namespace label and key to watch, e.g. 'environment: dev', If empty or undefined, the operator will run in cluster scope."`
	WatchLabelSelectors    string `env:"WATCH_LABEL_SELECTORS"                                       help:"The resources to watch according to their labels. e.g. 'partition in (customerA, customerB),environment!=qa'. If empty of undefined, the operator will watch all CRs."`
	CachingLevel           string `env:"ENFORCE_CACHE_LABELS"     default:"safe" enum:"all,safe,off" help:"Configure cache limits. Valid values are 'off', 'safe' and 'all'"`

	MetricsAddr             string        `name:"metrics-bind-address"      default:":8080"                                 help:"The address the metric endpoint binds to."`
	ProbeAddr               string        `name:"health-probe-bind-address" default:":8081"                                 help:"The address the probe endpoint binds to."`
	PprofAddr               string        `name:"pprof-addr"                                                                help:"The address to expose the pprof server. Empty string disables the pprof server."`
	EnableLeaderElection    bool          `name:"leader-elect"              default:"false" env:"ENABLE_LEADER_ELECTION"    help:"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager."`
	MaxConcurrentReconciles int           `name:"max-concurrent-reconciles" default:"1"     env:"MAX_CONCURRENT_RECONCILES" help:"Maximum number of concurrent reconciles for dashboard, datasource, folder controllers."`
	ResyncPeriod            time.Duration `name:"default-resync-period"     default:"10m"   env:"DEFAULT_RESYNC_PERIOD"     help:"Controls the default .spec.resyncPeriod when undefined on CRs."`

	ZapDevel           bool   `name:"zap-devel"            default:"false"                                                         help:"Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn)"`
	ZapEncoder         string `name:"zap-encoder"          default:"console" enum:"console,json"                                   help:"Zap log encoding ('json' or 'console')"`
	ZapLogLevel        string `name:"zap-log-level"        default:"info"                                                          help:"Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', 'panic' or any integer value > 0 which corresponds to custom debug levels of increasing verbosity"`
	ZapTimeEncoding    string `name:"zap-time-encoding"    default:"iso8601" enum:"epoch,millis,nanos,iso8601,rfc3339,rfc3339nano" help:"Zap time encoding ('epoch', 'millis', 'nanos', 'iso8601', 'rfc3339' or 'rfc3339nano')."`
	ZapStacktraceLevel string `name:"zap-stacktrace-level" default:"error"   enum:"info,error,panic"                               help:"Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic')."`
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1beta1.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))

	utilruntime.Must(gwapiv1.Install(scheme))
	//+kubebuilder:scaffold:scheme
}

func configureZap() (zap.Options, error) {
	opts := zap.Options{}

	opts.Development = operatorConfig.ZapDevel
	switch operatorConfig.ZapEncoder {
	case "json":
		opts.NewEncoder = func(eco ...zap.EncoderConfigOption) zapcore.Encoder {
			encoderConfig := uberzap.NewProductionEncoderConfig()
			for _, opt := range eco {
				opt(&encoderConfig)
			}

			return zapcore.NewJSONEncoder(encoderConfig)
		}
	case "console":
		opts.NewEncoder = func(eco ...zap.EncoderConfigOption) zapcore.Encoder {
			encoderConfig := uberzap.NewProductionEncoderConfig()
			for _, opt := range eco {
				opt(&encoderConfig)
			}

			return zapcore.NewConsoleEncoder(encoderConfig)
		}
	default:
		return opts, fmt.Errorf("invalid encoder %s", operatorConfig.ZapEncoder)
	}

	numericLevel, err := strconv.Atoi(operatorConfig.ZapLogLevel)
	if err == nil {
		opts.Level = uberzap.NewAtomicLevelAt(zapcore.Level(int8(numericLevel))) // #nosec G115
	} else {
		level, err := zapcore.ParseLevel(operatorConfig.ZapLogLevel)
		if err != nil {
			return opts, fmt.Errorf("invalid log level: %w", err)
		}

		opts.Level = level
	}

	stacktraceLevel, err := zapcore.ParseLevel(operatorConfig.ZapStacktraceLevel)
	if err != nil {
		return opts, fmt.Errorf("invalid log level: %w", err)
	}

	opts.StacktraceLevel = stacktraceLevel

	var timeEncoder zapcore.TimeEncoder

	err = timeEncoder.UnmarshalText([]byte(operatorConfig.ZapTimeEncoding))
	if err != nil {
		return opts, fmt.Errorf("invalid log encoder: %w", err)
	}

	opts.TimeEncoder = timeEncoder

	return opts, nil
}

func main() { //nolint:gocyclo
	kong.Parse(&operatorConfig,
		kong.Name("grafana-operator"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: false,
		}),
	)

	opts, err := configureZap()
	if err != nil {
		fmt.Println(err.Error()) //nolint:forbidigo
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	slogger := slog.New(logr.ToSlogHandler(setupLog))
	slog.SetDefault(slogger)

	// Optimize Go runtime based on CGroup limits (GOMEMLIMIT, sets a soft memory limit for the runtime)
	memlimit.SetGoMemLimitWithOpts(memlimit.WithLogger(slogger)) //nolint:errcheck

	// Determine LeaderElectionID from
	leHash := sha256.New()
	leHash.Write([]byte(operatorConfig.WatchNamespace))
	leHash.Write([]byte(operatorConfig.WatchNamespaceSelector))
	leHash.Write([]byte(operatorConfig.WatchLabelSelectors))

	// Fetch k8s api credentials and detect platform
	restConfig := ctrl.GetConfigOrDie()

	cluster, err := autodetect.NewClusterDiscovery(restConfig)
	if err != nil {
		setupLog.Error(err, "failed to setup auto-detect routine")
		os.Exit(1)
	}

	isOpenShift, err := cluster.IsOpenshift()
	if err != nil {
		setupLog.Error(err, "unable to detect the platform")
		os.Exit(1)
	}

	hasHTTPRouteCRD, err := cluster.HasHTTPRouteCRD()
	if err != nil {
		setupLog.Error(err, "failed to test for HTTPRoute CRD")
		os.Exit(1)
	}

	mgrOptions := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: operatorConfig.MetricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		HealthProbeBindAddress: operatorConfig.ProbeAddr,
		LeaderElection:         operatorConfig.EnableLeaderElection,
		LeaderElectionID:       fmt.Sprintf("grafana-operator-%x", leHash.Sum(nil)),
		PprofBindAddress:       operatorConfig.PprofAddr,
		Controller: config.Controller{
			MaxConcurrentReconciles: operatorConfig.MaxConcurrentReconciles,
		},
	}

	// A non-empty watchLabelSelector will attempt to enable sharding of the operator.
	// An invalid configuration will produce and error and exit early
	// If watchLabelSelector is empty, match any label configuration
	labelSelectors, err := getLabelSelectors(operatorConfig.WatchLabelSelectors)
	if err != nil {
		setupLog.Error(err, "invalid shard selector configuration")
		os.Exit(1)
	}

	if operatorConfig.WatchLabelSelectors != "" {
		setupLog.Info(fmt.Sprintf("sharding enabled via selector '%s'. Beware: Always label Grafana CRs before enabling to ensure labels are inherited. Existing Secrets/ConfigMaps referenced in CRs also need to be labeled to continue working.", operatorConfig.WatchLabelSelectors))
	}

	setupLog.Info("label restrictions for caching resources", "level", operatorConfig.CachingLevel)

	if operatorConfig.CachingLevel != cachingLevelOff {
		var cacheLabelConfig cache.ByObject
		if operatorConfig.WatchLabelSelectors != "" {
			// When sharding, limit cache according to shard labels
			cacheLabelConfig = cache.ByObject{Label: labelSelectors}
		} else {
			// Otherwise limit it to managed-by label
			cacheLabelConfig = cache.ByObject{Label: labels.SelectorFromSet(resources.GetCommonLabels())}
		}

		// ConfigMaps and secrets stay fully cached until we implement support for bypassing the cache for referenced objects
		mgrOptions.Cache.ByObject = map[client.Object]cache.ByObject{
			&appsv1.Deployment{}:            cacheLabelConfig,
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

		if hasHTTPRouteCRD {
			mgrOptions.Cache.ByObject[&gwapiv1.HTTPRoute{}] = cacheLabelConfig
		} else {
			setupLog.Info("skipping cache fine tuning for HTTPRoute resources as GatewayAPI CRDs were not found in the cluster")
		}

		if operatorConfig.CachingLevel == cachingLevelSafe {
			mgrOptions.Client.Cache = &client.CacheOptions{
				DisableFor: []client.Object{&corev1.ConfigMap{}, &corev1.Secret{}},
			}
		}
	}

	ctx := ctrl.SetupSignalHandler()
	ctx = klog.NewContext(ctx, setupLog) // Leader election logger is set through the ctx

	// Determine Operator scope
	switch {
	case strings.Contains(operatorConfig.WatchNamespace, ","):
		// multi namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfig(operatorConfig.WatchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode for multiple namespaces", "namespaces", operatorConfig.WatchNamespace)
	case operatorConfig.WatchNamespace != "":
		// namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfig(operatorConfig.WatchNamespace, labelSelectors)
		setupLog.Info("operator running in namespace scoped mode", "namespace", operatorConfig.WatchNamespace)
	case strings.Contains(operatorConfig.WatchNamespaceSelector, ":"):
		// multi namespace scoped
		mgrOptions.Cache.DefaultNamespaces = getNamespaceConfigSelector(ctx, restConfig, operatorConfig.WatchNamespaceSelector, labelSelectors)

		setupLog.Info("operator running in namespace scoped mode using namespace selector", "selector", operatorConfig.WatchNamespaceSelector)
	case operatorConfig.WatchNamespace == "" && operatorConfig.WatchNamespaceSelector == "":
		// cluster scoped
		mgrOptions.Cache.DefaultLabelSelector = labelSelectors

		setupLog.Info("operator running in cluster scoped mode")
	}

	mgr, err := ctrl.NewManager(restConfig, mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to create new manager")
		os.Exit(1)
	}

	ctrlCfg := &controllers.Config{
		ResyncPeriod: operatorConfig.ResyncPeriod,
	}
	// Register controllers
	if err = (&controllers.GrafanaReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		IsOpenShift:     isOpenShift,
		HasHTTPRouteCRD: hasHTTPRouteCRD,
		ClusterDomain:   operatorConfig.ClusterDomain,
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
		Recorder: mgr.GetEventRecorder("GrafanaNotificationPolicy"),
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

	if err = (&controllers.GrafanaManifestReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Cfg:    ctrlCfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaManifest")
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

	setupLog.Info("starting operator")

	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running operator")
		os.Exit(1)
	}

	setupLog.Info("shutting down operator")
}

func getNamespaceConfig(namespaces string, labelSelectors labels.Selector) map[string]cache.Config {
	defaultNamespaces := map[string]cache.Config{}
	for v := range strings.SplitSeq(namespaces, ",") {
		// Generate a mapping of namespaces to label/field selectors, set to Everything() to enable matching all
		// instances in all namespaces from operatorConfig.WatchNamespace to be controlled by the operator
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

func getNamespaceConfigSelector(ctx context.Context, restConfig *rest.Config, selector string, labelSelectors labels.Selector) map[string]cache.Config {
	kv := strings.Split(selector, ":")
	if len(kv) != 2 {
		err := fmt.Errorf("want pattern 'key:val', got: '%s'", selector)
		setupLog.Error(err, "failed to parse WATCH_NAMESPACE_SELECTOR")
		os.Exit(1)
	}

	cl, err := client.New(restConfig, client.Options{})
	if err != nil {
		setupLog.Error(err, "failed to create a kubernetes client")
		os.Exit(1)
	}

	nsList := &corev1.NamespaceList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(map[string]string{kv[0]: kv[1]}),
	}

	err = cl.List(ctx, nsList, listOpts...)
	if err != nil {
		setupLog.Error(err, "failed to get watch namespaces")
		os.Exit(1)
	}

	defaultNamespaces := map[string]cache.Config{}
	for _, v := range nsList.Items {
		// Generate a mapping of namespaces to label/field selectors, set to Everything() to enable matching all
		// instances in all namespaces from operatorConfig.WatchNamespace to be controlled by the operator
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
			return labelSelectors, fmt.Errorf("unable to parse 'WATCH_LABEL_SELECTOR=%s': %w", watchLabelSelectors, err)
		}
	} else {
		labelSelectors = labels.Everything() // Match any labels
	}

	managedByLabelSelector, _ := labels.SelectorFromSet(resources.GetCommonLabels()).Requirements()
	labelSelectors.Add(managedByLabelSelector...)

	return labelSelectors, nil
}
