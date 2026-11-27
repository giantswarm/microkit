package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelCode = "code"
)

var (
	endpointTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "endpoint_total",
			Help: "Number of times we have execute the HTTP handler of an endpoint.",
		},
		[]string{labelCode, "method", "name"},
	)
	endpointTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "endpoint_milliseconds",
			Help: "Time taken to execute the HTTP handler of an endpoint, in milliseconds.",
		},
		[]string{labelCode, "method", "name"},
	)
	errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "error_total",
			Help: "Number of times we have seen a specific error within a specific error domain.",
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(endpointTotal)
	prometheus.MustRegister(endpointTime)
	prometheus.MustRegister(errorTotal)
}
