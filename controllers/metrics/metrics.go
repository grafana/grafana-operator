package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	instanceName        = "instance_name"
	instanceNamespace   = "instance_namespace"
	method              = "method"
	namespace           = "grafana_operator"
	requests            = "requests"
	status              = "status"
	subsystemReconciler = "reconciler"
)

var (
	GrafanaReconciles = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystemReconciler,
		Name:      "reconciles",
		Help:      "reconciles per Grafana instance",
	}, []string{instanceNamespace, instanceName})

	GrafanaFailedReconciles = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystemReconciler,
		Name:      "failed_reconciles",
		Help:      "failed reconciles per Grafana instance and stage",
	}, []string{instanceNamespace, instanceName, "stage"})

	GrafanaAPIRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "grafana_api",
		Name:      requests,
		Help:      "requests against the grafana api per instance",
	}, []string{instanceNamespace, instanceName, method, status})

	// Deprecated: will be removed in a future version of the operator. Use
	// ContentURLRequests instead, which handles more types of resources that
	// directly utilize Grafana model JSON.
	DashboardURLRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "dashboards",
		Name:      requests,
		Help:      "requests to fetch dashboards from urls",
	}, []string{"dashboard", method, status})

	ContentURLRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "content",
		Name:      requests,
		Help:      "requests to fetch model contents from urls",
	}, []string{"kind", "resource", method, status})

	GrafanaComAPIRevisionRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "revision_requests",
		Help:      "requests to list content revisions on grafana.com",
	}, []string{"kind", "resource", method, status})

	InitialStatusSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemReconciler,
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync statuses after operator restart",
	})
)

func init() {
	metrics.Registry.MustRegister(GrafanaReconciles)
	metrics.Registry.MustRegister(GrafanaFailedReconciles)
	metrics.Registry.MustRegister(GrafanaAPIRequests)
	metrics.Registry.MustRegister(GrafanaComAPIRevisionRequests)
	metrics.Registry.MustRegister(DashboardURLRequests)
	metrics.Registry.MustRegister(ContentURLRequests)
	metrics.Registry.MustRegister(InitialStatusSyncDuration)
}
