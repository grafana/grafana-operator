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
	}, []string{"instance_name", "path", "method", "status"})

	DashboardUrlRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "grafana_operator",
		Subsystem: "dashboards",
		Name:      "requests",
		Help:      "requests to fetch dashboards from urls",
	}, []string{"dashboard", "path", "method", "status"})
)

func init() {
	metrics.Registry.MustRegister(GrafanaReconciles)
	metrics.Registry.MustRegister(GrafanaFailedReconciles)
	metrics.Registry.MustRegister(GrafanaApiRequests)
	metrics.Registry.MustRegister(DashboardUrlRequests)
}
