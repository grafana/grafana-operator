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
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	routev1 "github.com/openshift/api/route/v1"
	discovery2 "k8s.io/client-go/discovery"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers"
	"github.com/grafana-operator/grafana-operator/v5/controllers/autodetect"
	//+kubebuilder:scaffold:imports
)

const (
	// watchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE which specifies the Namespace to watch.
	// If empty or undefined, the operator will run in cluster scope.
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
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

func getWatchNamespaceFromLabelsInEnv(watchLabelEnvVar string) (string, error) {
	label, found := os.LookupEnv(watchLabelEnvVar)
	if !found {
		return "", fmt.Errorf("%s isn't set", watchLabelEnvVar)
	}

	ns, err := getNamespacesWithLabel(label)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the namespaces for a label: %s", err)
	}

	return ns, nil
}

func getNamespacesWithLabel(label string) (string, error) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	list, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return "", err
	}

	nsStrList := make([]string, 0)
	for _, ns := range list.Items {
		nsStrList = append(nsStrList, ns.Name)
	}

	return strings.Join(nsStrList, ","), nil
}

func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s isn't set", watchNamespaceEnvVar)
	}
	return ns, nil
}

func main() {
	var metricsAddr string
	var metricsAddrCrd string
	var enableLeaderElection bool
	var probeAddr string
	var probeAddrCrd string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddrCrd, "metrics-bind-address-crd", ":8082", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddrCrd, "health-probe-bind-address-crd", ":8084", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Get the namespaces to watch for all objects using the selector label
	watchNamespace, err := getWatchNamespaceFromLabelsInEnv("WATCH_LABEL")
	if err != nil {
		setupLog.Error(err, "unable to get watch namespace from labels, operator running in cluster scoped mode")
	}

	// If we didn't get watchNamespace from label, check if it's specifically set
	if watchNamespace == "" {
		watchNamespace, err = getWatchNamespace()
		if err != nil {
			setupLog.Error(err, "unable to get watch namespace, operator running in cluster scoped mode")
		}
	}

	// Get the namespaces to watch for CRD objects using the selector label
	watchNamespaceCrd, err := getWatchNamespaceFromLabelsInEnv("WATCH_LABEL_CRD")
	if err != nil {
		setupLog.Error(err, "unable to get CRD watch namespace from labels, using generic namespace list")
		watchNamespaceCrd = watchNamespace
	}

	controllerOptions := ctrl.Options{
		Namespace:              watchNamespace,
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f75f3bba.integreatly.org",
	}
	controllerOptionsCrd := ctrl.Options{
		Namespace:              watchNamespaceCrd,
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddrCrd,
		Port:                   9443,
		HealthProbeBindAddress: probeAddrCrd,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f75f3bbb.integreatly.org",
	}

	// Prepare the controllerOptions
	switch {
	case strings.Contains(watchNamespace, ","):
		// multi namespace scoped
		setupLog.Info("manager set up with multiple namespaces", "namespaces", watchNamespace)
		controllerOptions.Namespace = ""
		controllerOptions.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNamespace, ","))
	case watchNamespace != "":
		// namespace scoped
		setupLog.Info("operator running in namespace scoped mode", "namespace", watchNamespace)
	case watchNamespace == "":
		// cluster scoped
		setupLog.Info("operator running in cluster scoped mode")
	}

	// Prepare the controllerOptionsCrd
	switch {
	case strings.Contains(watchNamespaceCrd, ","):
		// multi namespace scoped
		setupLog.Info("CRD manager set up with multiple namespaces", "namespaces", watchNamespaceCrd)
		controllerOptionsCrd.Namespace = ""
		controllerOptionsCrd.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNamespaceCrd, ","))
	case watchNamespaceCrd != "":
		// namespace scoped
		setupLog.Info("CRD operator running in namespace scoped mode", "namespace", watchNamespaceCrd)
		controllerOptionsCrd.Namespace = watchNamespaceCrd
		controllerOptions.NewCache = cache.BuilderWithOptions(cache.Options{Namespace: watchNamespaceCrd})
	case watchNamespaceCrd == "":
		// cluster scoped
		setupLog.Info("CRD operator running in cluster scoped mode")
		controllerOptions.NewCache = cache.BuilderWithOptions(cache.Options{})
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	defer stop()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), controllerOptions)
	if err != nil {
		setupLog.Error(err, "unable to create new manager")
		os.Exit(1) //nolint
	}

	mgrCrd, err := ctrl.NewManager(ctrl.GetConfigOrDie(), controllerOptionsCrd)
	if err != nil {
		setupLog.Error(err, "unable to create new CRD manager")
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
		Client: mgrCrd.GetClient(),
		Scheme: mgrCrd.GetScheme(),
		Log:    ctrl.Log.WithName("DashboardReconciler"),
	}).SetupWithManager(mgrCrd, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDashboard")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaDatasourceReconciler{
		Client: mgrCrd.GetClient(),
		Scheme: mgrCrd.GetScheme(),
		Log:    ctrl.Log.WithName("DatasourceReconciler"),
	}).SetupWithManager(mgrCrd, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaDatasource")
		os.Exit(1)
	}
	if err = (&controllers.GrafanaFolderReconciler{
		Client: mgrCrd.GetClient(),
		Scheme: mgrCrd.GetScheme(),
	}).SetupWithManager(mgrCrd, ctx); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GrafanaFolder")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgrCrd.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	mgrCtx := ctrl.SetupSignalHandler()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		setupLog.Info("starting manager")
		if err := mgr.Start(mgrCtx); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
		<-ctx.Done()
		wg.Done()
		setupLog.Info("SIGTERM request gotten, shutting down operator")
	}()
	go func() {
		setupLog.Info("starting managerCrd")
		if err := mgrCrd.Start(mgrCtx); err != nil {
			setupLog.Error(err, "problem running managerCrd")
			os.Exit(1)
		}
		<-ctx.Done()
		wg.Done()
		setupLog.Info("SIGTERM request gotten, shutting down operator")
	}()

	wg.Wait()
}
