package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// Config holds observability configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	SampleRate     float64
	EnableMetrics  bool
	EnableTracing  bool
	EnableLogging  bool
}

// DefaultConfig returns default observability configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:    getEnvOrDefault("OTEL_SERVICE_NAME", "offgridflow"),
		ServiceVersion: getEnvOrDefault("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnvOrDefault("ENVIRONMENT", "development"),
		OTLPEndpoint:   getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		SampleRate:     1.0, // Sample all traces in development
		EnableMetrics:  true,
		EnableTracing:  true,
		EnableLogging:  true,
	}
}

// Provider manages all observability components
type Provider struct {
	config          Config
	logger          *slog.Logger
	tracerProvider  *TracerProvider
	metricsProvider *MetricsProvider
	metrics         *Metrics
}

// NewProvider creates and initializes all observability components
func NewProvider(ctx context.Context, config Config) (*Provider, error) {
	provider := &Provider{
		config: config,
	}

	// Initialize logger
	if config.EnableLogging {
		provider.logger = NewStructuredLogger()
		provider.logger.Info("logging initialized",
			slog.String("service", config.ServiceName),
			slog.String("environment", config.Environment),
		)
	} else {
		provider.logger = slog.Default()
	}

	// Initialize tracing
	if config.EnableTracing {
		tracerConfig := TracerConfig{
			ServiceName:    config.ServiceName,
			ServiceVersion: config.ServiceVersion,
			Environment:    config.Environment,
			OTLPEndpoint:   config.OTLPEndpoint,
			SampleRate:     config.SampleRate,
		}

		tp, err := NewTracerProvider(ctx, tracerConfig)
		if err != nil {
			provider.logger.Warn("failed to initialize tracer", slog.String("error", err.Error()))
			// Continue without tracing rather than failing
		} else {
			provider.tracerProvider = tp
			provider.logger.Info("tracing initialized", slog.String("endpoint", config.OTLPEndpoint))
		}
	}

	// Initialize metrics
	if config.EnableMetrics {
		mp, err := NewMetricsProvider(ctx, TracerConfig{
			ServiceName:    config.ServiceName,
			ServiceVersion: config.ServiceVersion,
			Environment:    config.Environment,
			OTLPEndpoint:   config.OTLPEndpoint,
		})
		if err != nil {
			provider.logger.Warn("failed to initialize metrics provider", slog.String("error", err.Error()))
			// Continue without metrics rather than failing
		} else {
			provider.metricsProvider = mp
			provider.logger.Info("metrics provider initialized")

			// Create metrics
			metrics, err := NewMetrics(config.ServiceName)
			if err != nil {
				provider.logger.Warn("failed to create metrics", slog.String("error", err.Error()))
			} else {
				provider.metrics = metrics
				provider.logger.Info("metrics initialized")
			}
		}
	}

	return provider, nil
}

// Shutdown gracefully shuts down all observability components
func (p *Provider) Shutdown(ctx context.Context) error {
	var lastErr error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			p.logger.Error("failed to shutdown tracer", slog.String("error", err.Error()))
			lastErr = err
		}
	}

	if p.metricsProvider != nil {
		if err := p.metricsProvider.Shutdown(ctx); err != nil {
			p.logger.Error("failed to shutdown metrics", slog.String("error", err.Error()))
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("observability shutdown errors occurred: %w", lastErr)
	}

	p.logger.Info("observability shutdown complete")
	return nil
}

// Logger returns the configured logger
func (p *Provider) Logger() *slog.Logger {
	return p.logger
}

// Metrics returns the metrics instance
func (p *Provider) Metrics() *Metrics {
	return p.metrics
}

// HTTPMiddleware returns an HTTP middleware that combines logging, tracing, and metrics
func (p *Provider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		var handler http.Handler = next

		// Apply metrics and tracing middleware if available
		if p.metrics != nil {
			tracingMiddleware := NewHTTPMiddleware(p.config.ServiceName, p.metrics)
			handler = tracingMiddleware.Handler(handler)
		}

		// Apply logging middleware
		if p.config.EnableLogging {
			loggingMiddleware := NewLoggingMiddleware(p.logger)
			handler = loggingMiddleware.Handler(handler)
		}

		return handler
	}
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
