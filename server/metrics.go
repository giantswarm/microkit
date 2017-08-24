package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "microkit"
	prometheusSubsystem = "endpoint"
)

var (
	endpointTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "duration_seconds",
			Help:      "Time taken to execute the HTTP handler of an endpoint, in seconds.",
		},
		[]string{"method", "name"},
	)

	endpointTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "total",
			Help:      "Total count of all HTTP requests, with response codes.",
		},
		[]string{"code", "method", "name"},
	)
)

func init() {
	prometheus.MustRegister(endpointTime)
}
