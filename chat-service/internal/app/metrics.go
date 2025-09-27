package app

import "github.com/prometheus/client_golang/prometheus"

// ChatLatencyHistogram is a Prometheus histogram for chat response latency
var ChatLatencyHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "chat_response_latency_seconds",
		Help:    "Latency of chat responses in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"endpoint"},
)
