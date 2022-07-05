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
)

func init() {
	metrics.Registry.MustRegister(GrafanaReconciles)
	metrics.Registry.MustRegister(GrafanaFailedReconciles)
}
