package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	endpointTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "endpoint_total",
			Help: "Number of times we have execute the HTTP handler of an endpoint",
		},
		[]string{"method", "endpoint", "code"},
	)
	endpointTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "endpoint_milliseconds",
			Help: "Time taken to execute the HTTP handler of an endpoint, in milliseconds",
		},
		[]string{"method", "endpoint", "code"},
	)
	errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "error_total",
			Help: "Number of times we have seen a specific error",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(endpointTotal)
	prometheus.MustRegister(endpointTime)
	prometheus.MustRegister(errorTotal)
}
