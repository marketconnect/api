package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func init() {
	// Register Go runtime metrics safely - only register if not already registered
	// This approach won't panic if the collector is already registered
	registry := prometheus.DefaultRegisterer
	if registry != nil {
		// Try to register, but don't panic if already registered
		err := registry.Register(collectors.NewGoCollector())
		if err != nil {
			// Log error or ignore - this is expected if already registered
			// In production, you might want to log this
		}
	}
}

var (
	// HTTPMetrics are general HTTP request metrics.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"code", "method", "handler"},
	)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "method", "handler"},
	)

	// AppCardCreationsTotal is a counter for successful card creations (core logic).
	AppCardCreationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_card_creations_total",
			Help: "Total number of successful card creations (core CardCraftAI content generation).",
		},
	)
	// AppExternalAPIErrorsTotal is a counter for errors from external APIs.
	AppExternalAPIErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_external_api_errors_total",
			Help: "Total number of errors from external APIs.",
		},
		[]string{"api_name"}, // e.g., "card_craft_ai_session", "wb_card_upload"
	)
	// AppWBMediaOperationsTotal is a counter for Wildberries media operations.
	AppWBMediaOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_wb_media_operations_total",
			Help: "Total number of Wildberries media operations attempted.",
		},
		[]string{"operation_type"}, // "upload_file", "save_by_link"
	)
	// AppWBMediaOperationErrorsTotal is a counter for errors during Wildberries media operations.
	AppWBMediaOperationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_wb_media_operation_errors_total",
			Help: "Total number of errors during Wildberries media operations.",
		},
		[]string{"operation_type"}, // "upload_file", "save_by_link"
	)
)
