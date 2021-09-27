/*
Copyright 2021.

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
	"github.com/integr8ly/grafana-operator/controllers/grafananotificationchannel"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"

	apis "github.com/integr8ly/grafana-operator/api"
	"github.com/integr8ly/grafana-operator/controllers/common"
	grafanaconfig "github.com/integr8ly/grafana-operator/controllers/config"
	"github.com/integr8ly/grafana-operator/controllers/grafana"
	"github.com/integr8ly/grafana-operator/controllers/grafanadashboard"
	"github.com/integr8ly/grafana-operator/controllers/grafanadatasource"
	"github.com/integr8ly/grafana-operator/internal/k8sutil"
	"github.com/integr8ly/grafana-operator/version"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-lib/leader"
	"k8s.io/client-go/rest"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	integreatlyorgv1alpha1 "github.com/integr8ly/grafana-operator/api/integreatly/v1alpha1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	// +kubebuilder:scaffold:imports
)

var (
	scheme                        = k8sruntime.NewScheme()
	setupLog                      = ctrl.Log.WithName("setup")
	flagImage                     string
	flagImageTag                  string
	flagPluginsInitContainerImage string
	flagPluginsInitContainerTag   string
	flagNamespaces                string
	scanAll                       bool
	flagJsonnetLocation           string
	metricsAddr                   string
	enableLeaderElection          bool
	probeAddr                     string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(integreatlyorgv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	log.Log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Log.Info(fmt.Sprintf("operator-sdk Version: %v", "v1.3.0"))
	log.Log.Info(fmt.Sprintf("operator Version: %v", version.Version))
}

func assignOpts() {
	flag.StringVar(&flagImage, "grafana-image", "", "Overrides the default Grafana image")
	flag.StringVar(&flagImageTag, "grafana-image-tag", "", "Overrides the default Grafana image tag")
	flag.StringVar(&flagPluginsInitContainerImage, "grafana-plugins-init-container-image", "", "Overrides the default Grafana Plugins Init Container image")
	flag.StringVar(&flagPluginsInitContainerTag, "grafana-plugins-init-container-tag", "", "Overrides the default Grafana Plugins Init Container tag")
	flag.StringVar(&flagNamespaces, "namespaces", "", "Namespaces to scope the interaction of the Grafana operator. Mutually exclusive with --scan-all")
	flag.StringVar(&flagJsonnetLocation, "jsonnet-location", "", "Overrides the base path of the jsonnet libraries")
	flag.BoolVar(&scanAll, "scan-all", false, "Scans all namespaces for dashboards")

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
}

func main() { // nolint

	printVersion()
	assignOpts()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Log.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Namespace:              namespace,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "2c0156f0.integreatly.org",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if scanAll && flagNamespaces != "" {
		fmt.Fprint(os.Stderr, "--scan-all and --namespaces both set. Please provide only one")
		os.Exit(1)
	}

	// Controller configuration
	controllerConfig := grafanaconfig.GetControllerConfig()
	controllerConfig.AddConfigItem(grafanaconfig.ConfigGrafanaImage, flagImage)
	controllerConfig.AddConfigItem(grafanaconfig.ConfigGrafanaImageTag, flagImageTag)
	controllerConfig.AddConfigItem(grafanaconfig.ConfigPluginsInitContainerImage, flagPluginsInitContainerImage)
	controllerConfig.AddConfigItem(grafanaconfig.ConfigPluginsInitContainerTag, flagPluginsInitContainerTag)
	controllerConfig.AddConfigItem(grafanaconfig.ConfigOperatorNamespace, namespace)
	controllerConfig.AddConfigItem(grafanaconfig.ConfigDashboardLabelSelector, "")
	controllerConfig.AddConfigItem(grafanaconfig.ConfigJsonnetBasePath, flagJsonnetLocation)

	// Get the namespaces to scan for dashboards
	// It's either the same namespace as the controller's or it's all namespaces if the
	// --scan-all flag has been passed
	var dashboardNamespaces = []string{namespace}
	if scanAll {
		dashboardNamespaces = []string{""}
		log.Log.Info("Scanning for dashboards in all namespaces")
	}

	if flagNamespaces != "" {
		dashboardNamespaces = getSanitizedNamespaceList()
		if len(dashboardNamespaces) == 0 {
			fmt.Fprint(os.Stderr, "--namespaces provided but no valid namespaces in list")
			os.Exit(1)
		}
		log.Log.Info(fmt.Sprintf("Scanning for dashboards in the following namespaces: [%s]", strings.Join(dashboardNamespaces, ",")))
	}

	// Get a config to talk to the apiserver
	cfg, err := k8sconfig.GetConfig()
	if err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Become the leader before proceeding
	err = leader.Become(context.TODO(), "grafana-operator-lock")
	if err != nil {
		log.Log.Error(err, "")
	}

	log.Log.Info("Registering Components.")

	// Starting the resource auto-detection for the grafana controller
	autodetect, err := common.NewAutoDetect(mgr)
	if err != nil {
		log.Log.Error(err, "failed to start the background process to auto-detect the operator capabilities")
	} else {
		autodetect.Start()
		defer autodetect.Stop()
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift routes
	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	if err != nil {
		log.Log.Error(err, "error starting metrics service")
	}

	log.Log.Info("Starting the Cmd.")

	// Start one dashboard controller per watch namespace
	for _, ns := range dashboardNamespaces {
		startDashboardController(ns, cfg, context.Background())
		startNotificationChannelController(ns, cfg, context.Background())
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	if err = (&grafana.ReconcileGrafana{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Plugins:  grafana.NewPluginsHelper(),
		Context:  ctx,
		Cancel:   cancel,
		Config:   grafanaconfig.GetControllerConfig(),
		Recorder: mgr.GetEventRecorderFor("GrafanaDashboard"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Grafana")
		os.Exit(1)
	}
	if err = (&grafanadashboard.GrafanaDashboardReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("GrafanaDashboard"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}
	if err = (&grafanadatasource.GrafanaDatasourceReconciler{
		Client:  mgr.GetClient(),
		Context: ctx,
		Cancel:  cancel,
		Logger:  ctrl.Log.WithName("controllers").WithName("GrafanaDatasource"),
		Scheme:  mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDatasource")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// Starts a separate controller for the dashboard reconciliation in the background
func startDashboardController(ns string, cfg *rest.Config, ctx context.Context) {
	// Create a new Cmd to provide shared dependencies and start components
	dashboardMgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
		Namespace:          ns,
	})
	if err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for the dashboard resource
	if err := apis.AddToScheme(dashboardMgr.GetScheme()); err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Use a separate manager for the dashboard controller
	err = grafanadashboard.Add(dashboardMgr, ns)
	if err != nil {
		log.Log.Error(err, "")
	}

	go func() {
		if err := dashboardMgr.Start(ctx); err != nil {
			log.Log.Error(err, "dashboard manager exited non-zero")
			os.Exit(1)
		}
	}()
}

// Starts a separate controller for the notification channels reconciliation in the background
func startNotificationChannelController(ns string, cfg *rest.Config, ctx context.Context) {
	// Create a new Cmd to provide shared dependencies and start components
	channelMgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
		Namespace:          ns,
	})
	if err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for the notification channel resource
	if err := apis.AddToScheme(channelMgr.GetScheme()); err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	// Use a separate manager for the dashboard controller
	err = grafananotificationchannel.Add(channelMgr, ns)
	if err != nil {
		log.Log.Error(err, "")
		os.Exit(1)
	}

	go func() {
		if err := channelMgr.Start(ctx); err != nil {
			log.Log.Error(err, "notification channel manager exited non-zero")
			os.Exit(1)
		}
	}()
}

// Get the trimmed and sanitized list of namespaces (if --namespaces was provided)
func getSanitizedNamespaceList() []string {
	provided := strings.Split(flagNamespaces, ",")
	var selected []string

	for _, v := range provided {
		v = strings.TrimSpace(v)

		if v != "" {
			selected = append(selected, v)
		}
	}

	return selected
}
