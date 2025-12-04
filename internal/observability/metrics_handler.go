package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler provides Prometheus metrics endpoint
type MetricsHandler struct {
	registry *prometheus.Registry
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		registry: prometheus.NewRegistry(),
	}
}

// NewMetricsHandlerWithRegistry creates a metrics handler with a custom registry
func NewMetricsHandlerWithRegistry(registry *prometheus.Registry) *MetricsHandler {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}
	return &MetricsHandler{
		registry: registry,
	}
}

// Handler returns the HTTP handler for the /metrics endpoint
func (h *MetricsHandler) Handler() http.Handler {
	return promhttp.HandlerFor(
		h.registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}

// Registry returns the Prometheus registry
func (h *MetricsHandler) Registry() *prometheus.Registry {
	return h.registry
}

// RegisterCollector registers a Prometheus collector
func (h *MetricsHandler) RegisterCollector(collector prometheus.Collector) error {
	return h.registry.Register(collector)
}
