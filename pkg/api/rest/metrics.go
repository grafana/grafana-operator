package rest

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.Register(HTTPRequestsTotal)
}

var HTTPRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	},
	[]string{"code", "method"},
)
