package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/go-openapi/loads"
	apipkg "github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	restConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	DefaultPolicyFile string
	pathConverter     = regexp.MustCompile(`{(.+?)}`)
	namespace         string
	log               = logf.Log.WithName("apiserver_cmd")
	opts              config.Options
	metricsPort       int
)

func init() {
	flag.StringVar(&opts.ConfigFilePath, "config-file-path", "./etc/apiserver/config/config.yaml", "To set apiserver config file path")
	flag.StringVar(&DefaultPolicyFile, "policy", "etc/policy.json", "API authorization policy file")
	flag.StringVar(&namespace, "namespace", "grafana-operator", "k8s Namespace")
	flag.IntVar(&opts.MetricPort, "metrics-port", 9100, "Lister port for metric exposition")
	flag.IntVar(&opts.APIPort, "api-port", 8080, "Lister port for api exposition")

	flag.Parse()

	logf.SetLogger(logf.ZapLogger(false))
}

func main() {
	var server *rest.Server // make sure init is called

	//If --kubeconfig is set, will use the kubeconfig file at that location.
	cfg := restConfig.GetConfigOrDie()

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	apiCfg, err := config.GetConfig(opts)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	swaggerSpec, err := loads.Embedded(rest.SwaggerJSON, rest.FlatSwaggerJSON)
	if err != nil {
		log.Error(err, "Could not load Swagger json")
		os.Exit(1)

	}

	cl, err := client.New(cfg, client.Options{})
	if err != nil {
		log.Error(err, "failed to create client")
		os.Exit(1)
	}

	api := operations.NewGrafanaOperatorAPI(swaggerSpec)
	rt := &apipkg.Runtime{
		Namespace: types.NamespacedName{Namespace: namespace, Name: namespace},
		Client:    cl,
		Config:    apiCfg,
	}
	if err := rest.ConfigureGrafanaAPI(api, rt); err != nil {
		log.Error(err, "failed to configure API server")
		os.Exit(1)
	}

	server = rest.NewServer(api)
	defer server.Shutdown()

	//Setup metrics listener
	metricsHost := "0.0.0.0"
	metricsListener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", metricsHost, opts.MetricPort))
	log.Info(
		"msg", "Exposing metrics",
		"host", metricsHost,
		"port", opts.MetricPort,
		"err", err)
	if err == nil {
		go http.Serve(metricsListener, promhttp.Handler())
		api.ServerShutdown = func() {
			metricsListener.Close()
		}
	}
	server.Port = opts.APIPort
	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Error(err, "failed to start API server")
		os.Exit(1)
	}

}
