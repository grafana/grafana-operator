package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	GrafanaReconciles = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "reconciler",
		Name:      "reconciles",
		Help:      "reconciles per Grafana instance",
	}, []string{"instance_namespace", "instance_name"})

	GrafanaFailedReconciles = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "reconciler",
		Name:      "failed_reconciles",
		Help:      "failed reconciles per Grafana instance and stage",
	}, []string{"instance_namespace", "instance_name", "stage"})

	GrafanaAPIRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "grafana_api",
		Name:      "requests",
		Help:      "requests against the grafana api per instance",
	}, []string{"instance_namespace", "instance_name", "method", "status"})

	// Deprecated: will be removed in a future version of the operator. Use
	// ContentURLRequests instead, which handles more types of resources that
	// directly utilize Grafana model JSON.
	DashboardURLRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "requests",
		Help:      "requests to fetch dashboards from urls",
	}, []string{"dashboard", "method", "status"})

	ContentURLRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "content",
		Name:      "requests",
		Help:      "requests to fetch model contents from urls",
	}, []string{"kind", "resource", "method", "status"})

	GrafanaComAPIRevisionRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Name:      "revision_requests",
		Help:      "requests to list content revisions on grafana.com",
	}, []string{"kind", "resource", "method", "status"})

	InitialDashboardSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync dashboards after operator restart",
	})

	InitialLibraryPanelSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grafana_operator",
		Subsystem: "librarypanels",
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync library panels after operator restart",
	})

	InitialDatasourceSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grafana_operator",
		Subsystem: "datasources",
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync datasources after operator restart",
	})

	InitialFoldersSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grafana_operator",
		Subsystem: "folders",
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync folders after operator restart",
	})
)

func init() {
	metrics.Registry.MustRegister(GrafanaReconciles)
	metrics.Registry.MustRegister(GrafanaFailedReconciles)
	metrics.Registry.MustRegister(GrafanaAPIRequests)
	metrics.Registry.MustRegister(GrafanaComAPIRevisionRequests)
	metrics.Registry.MustRegister(DashboardURLRequests)
	metrics.Registry.MustRegister(ContentURLRequests)
	metrics.Registry.MustRegister(InitialDashboardSyncDuration)
	metrics.Registry.MustRegister(InitialLibraryPanelSyncDuration)
	metrics.Registry.MustRegister(InitialDatasourceSyncDuration)
	metrics.Registry.MustRegister(InitialFoldersSyncDuration)
}
