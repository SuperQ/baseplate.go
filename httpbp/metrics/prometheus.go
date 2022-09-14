package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/reddit/baseplate.go/prometheusbp"
)

const (
	MethodLabel     = "http_method"
	SuccessLabel    = "http_success"
	CodeLabel       = "http_response_code"
	ServerSlugLabel = "http_slug" // Deprecated, will be removed after 2022-09-01
	ClientNameLabel = "http_client_name"
	EndpointLabel   = "http_endpoint"
)

var (
	serverLabels = []string{
		MethodLabel,
		SuccessLabel,
		EndpointLabel,
	}

	ServerLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_latency_seconds",
		Help:    "HTTP server request latencies",
		Buckets: prometheusbp.DefaultBuckets,
	}, serverLabels)

	ServerRequestSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_size_bytes",
		Help:    "Request size",
		Buckets: prometheusbp.DefaultBuckets,
	}, serverLabels)

	ServerResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_response_size_bytes",
		Help:    "Response size",
		Buckets: prometheusbp.DefaultBuckets,
	}, serverLabels)

	ServerTimeToWriteHeader = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_time_to_write_header_seconds",
		Help:    "Request size",
		Buckets: prometheusbp.DefaultBuckets,
	}, serverLabels)

	ServerTimeToFirstByte = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_time_to_first_byte_seconds",
		Help:    "Response size",
		Buckets: prometheusbp.DefaultBuckets,
	}, serverLabels)

	serverTotalRequestLabels = []string{
		MethodLabel,
		SuccessLabel,
		CodeLabel,
		EndpointLabel,
	}

	ServerTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_server_requests_total",
		Help: "Total request count",
	}, serverTotalRequestLabels)

	serverActiveRequestsLabels = []string{
		MethodLabel,
		EndpointLabel,
	}

	ServerActiveRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_server_active_requests",
		Help: "The number of in-flight requests being handled by the service",
	}, serverActiveRequestsLabels)
)

var (
	clientLatencyLabels = []string{
		MethodLabel,
		SuccessLabel,
		ServerSlugLabel,
		ClientNameLabel,
	}

	ClientLatencyDistribution = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_client_latency_seconds",
		Help:    "HTTP client request latencies",
		Buckets: prometheusbp.DefaultBuckets,
	}, clientLatencyLabels)

	clientTotalRequestLabels = []string{
		MethodLabel,
		SuccessLabel,
		CodeLabel,
		ServerSlugLabel,
		ClientNameLabel,
	}

	ClientTotalRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_requests_total",
		Help: "Total request count",
	}, clientTotalRequestLabels)

	clientActiveRequestsLabels = []string{
		MethodLabel,
		ServerSlugLabel,
		ClientNameLabel,
	}

	ClientActiveRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_client_active_requests",
		Help: "The number of in-flight requests",
	}, clientActiveRequestsLabels)
)

const (
	// Note that this is not used by prometheus metrics defined in Baseplate spec.
	promNamespace   = "httpbp"
	subsystemServer = "server"
)

var (
	panicRecoverLabels = []string{
		MethodLabel,
	}

	PanicRecoverCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: subsystemServer,
		Name:      "panic_recover_total",
		Help:      "The number of panics recovered from http server handlers",
	}, panicRecoverLabels)
)

// PerformanceMonitoringMiddleware returns optional Prometheus historgram metrics for monitoring the following:
//  1. http server time to write header in seconds
//  2. http server time to write header in seconds
func PerformanceMonitoringMiddleware() (timeToWriteHeader, timeToFirstByte *prometheus.HistogramVec) {
	return ServerTimeToWriteHeader, ServerTimeToFirstByte
}
