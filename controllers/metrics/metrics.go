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
	}, []string{"instance_name"})

	GrafanaFailedReconciles = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "reconciler",
		Name:      "failed_reconciles",
		Help:      "failed reconciles per Grafana instance and stage",
	}, []string{"instance_name", "stage"})

	GrafanaApiRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "grafana_api",
		Name:      "requests",
		Help:      "requests against the grafana api per instance",
	}, []string{"instance_name", "method", "status"})

	DashboardUrlRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "requests",
		Help:      "requests to fetch dashboards from urls",
	}, []string{"dashboard", "method", "status"})

	GrafanaComApiRevisionRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "revision_requests",
		Help:      "requests to list dashboard revisions on grafana.com/dashboards",
	}, []string{"dashboard", "method", "status"})

	InitialDashboardSyncDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "initial_sync_duration",
		Help:      "time in ms to sync dashboards after operator restart",
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
	metrics.Registry.MustRegister(GrafanaApiRequests)
	metrics.Registry.MustRegister(DashboardUrlRequests)
	metrics.Registry.MustRegister(InitialDashboardSyncDuration)
	metrics.Registry.MustRegister(InitialDatasourceSyncDuration)
	metrics.Registry.MustRegister(InitialFoldersSyncDuration)
}
