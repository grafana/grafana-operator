package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	GrafanaReconciles = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:      "reconciles",
			Namespace: "grafana_operator",
			Subsystem: "controller",
			Help:      "counts the number of reconciles",
		})
)

func init() {
	metrics.Registry.MustRegister(GrafanaReconciles)
}
