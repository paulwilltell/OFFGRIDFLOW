package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/offgrid"
)

// =============================================================================
// Router Sentinel Errors
// =============================================================================

var (
	// ErrNoLocalProvider is returned when offline mode requires a local provider
	// but none was configured.
	ErrNoLocalProvider = errors.New("ai/router: no local provider configured")

	// ErrNoCloudProvider is returned when online mode requires a cloud provider
	// but none was configured or available.
	ErrNoCloudProvider = errors.New("ai/router: no cloud provider configured")

	// ErrNoProvidersConfigured is returned when neither cloud nor local
	// providers are available.
	ErrNoProvidersConfigured = errors.New("ai/router: no providers configured")

	// ErrNilModeManager is returned when the Router's ModeManager is nil.
	ErrNilModeManager = errors.New("ai/router: nil mode manager")

	// ErrNilRouter is returned when attempting to use a nil Router.
	ErrNilRouter = errors.New("ai/router: nil router")
)

// =============================================================================
// Router Configuration
// =============================================================================

// RouterConfig holds configuration for the AI router.
type RouterConfig struct {
	// ModeManager provides the current online/offline mode.
	// Required.
	ModeManager *offgrid.ModeManager

	// Cloud is the cloud provider (e.g., OpenAI).
	// Optional; if nil, only local provider will be used.
	Cloud CloudProvider

	// Local is the local/offline provider.
	// Optional; if nil, offline mode won't work.
	Local LocalProvider

	// Logger is used for operational logging.
	// If nil, a default slog logger is used.
	Logger *slog.Logger

	// EnableFallback controls whether the router falls back to Local
	// when Cloud fails in online mode. Defaults to true.
	EnableFallback *bool

	// CloudTimeout is an additional timeout applied to cloud requests
	// beyond any context deadline. If zero, no additional timeout is applied.
	CloudTimeout time.Duration

	// RetryConfig configures retry behavior for transient failures.
	RetryConfig *RetryConfig
}

// RetryConfig controls retry behavior.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including the first).
	// Defaults to 1 (no retries) if zero.
	MaxAttempts int

	// InitialBackoff is the delay before the first retry.
	// Defaults to 100ms if zero.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum delay between retries.
	// Defaults to 5s if zero.
	MaxBackoff time.Duration

	// BackoffMultiplier is applied after each retry.
	// Defaults to 2.0 if zero.
	BackoffMultiplier float64
}

func (c *RetryConfig) applyDefaults() {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 1
	}
	if c.InitialBackoff <= 0 {
		c.InitialBackoff = 100 * time.Millisecond
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = 5 * time.Second
	}
	if c.BackoffMultiplier <= 0 {
		c.BackoffMultiplier = 2.0
	}
}

// =============================================================================
// Router Implementation
// =============================================================================

// Router intelligently routes AI requests between cloud and local providers
// based on the current connectivity mode. It provides automatic fallback,
// metrics collection, and graceful degradation.
//
// Routing logic:
//   - Offline mode: Always use Local provider
//   - Online mode: Try Cloud first, fall back to Local on failure (if enabled)
//
// Thread safety: Router is safe for concurrent use.
type Router struct {
	modeManager    *offgrid.ModeManager
	cloud          CloudProvider
	local          LocalProvider
	logger         *slog.Logger
	enableFallback bool
	cloudTimeout   time.Duration
	retryConfig    RetryConfig

	// Metrics
	mu             sync.RWMutex
	cloudRequests  int64
	localRequests  int64
	fallbackCount  int64
	totalCloudTime time.Duration
	totalLocalTime time.Duration
}

// NewRouter creates a new Router with the given configuration.
// Returns an error if required configuration is missing.
func NewRouter(cfg RouterConfig) (*Router, error) {
	if cfg.ModeManager == nil {
		return nil, ErrNilModeManager
	}

	// At least one provider must be configured
	if cfg.Cloud == nil && cfg.Local == nil {
		return nil, ErrNoProvidersConfigured
	}

	// Default to enabling fallback
	enableFallback := true
	if cfg.EnableFallback != nil {
		enableFallback = *cfg.EnableFallback
	}

	// Default logger
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Apply retry defaults
	retryConfig := RetryConfig{}
	if cfg.RetryConfig != nil {
		retryConfig = *cfg.RetryConfig
	}
	retryConfig.applyDefaults()

	return &Router{
		modeManager:    cfg.ModeManager,
		cloud:          cfg.Cloud,
		local:          cfg.Local,
		logger:         logger,
		enableFallback: enableFallback,
		cloudTimeout:   cfg.CloudTimeout,
		retryConfig:    retryConfig,
	}, nil
}

// Chat routes the request based on the current connectivity mode.
//
// Behavior:
//   - Offline mode: Routes directly to Local provider
//   - Online mode: Routes to Cloud provider with optional fallback to Local
//
// The response.Source field indicates which provider handled the request.
func (r *Router) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if r == nil {
		return ChatResponse{}, ErrNilRouter
	}

	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		return ChatResponse{}, fmt.Errorf("ai/router: %w", err)
	}

	// Validate request early
	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	mode := r.modeManager.GetMode()

	r.logger.Debug("routing chat request",
		slog.String("mode", string(mode)),
		slog.Bool("cloud_available", r.cloud != nil && r.cloud.IsConfigured()),
		slog.Bool("local_available", r.local != nil),
	)

	// Offline mode: force local
	if mode == offgrid.ModeOffline {
		return r.routeToLocal(ctx, req, "offline mode")
	}

	// Online mode: try cloud with optional fallback
	return r.routeToCloudWithFallback(ctx, req)
}

// routeToLocal sends the request to the local provider.
func (r *Router) routeToLocal(ctx context.Context, req ChatRequest, reason string) (ChatResponse, error) {
	if r.local == nil {
		r.logger.Warn("local provider not available",
			slog.String("reason", reason),
		)
		return ChatResponse{}, ErrNoLocalProvider
	}

	// Check if local provider is available
	if !r.local.IsAvailable() {
		r.logger.Warn("local provider not ready",
			slog.String("reason", reason),
		)
		return ChatResponse{}, ErrProviderUnavailable
	}

	r.logger.Debug("routing to local provider",
		slog.String("reason", reason),
	)

	startTime := time.Now()
	resp, err := r.local.Chat(ctx, req)
	elapsed := time.Since(startTime)

	// Update metrics
	r.mu.Lock()
	r.localRequests++
	r.totalLocalTime += elapsed
	r.mu.Unlock()

	if err != nil {
		r.logger.Error("local provider failed",
			slog.String("error", err.Error()),
			slog.Duration("latency", elapsed),
		)
		return ChatResponse{}, fmt.Errorf("ai/router: local provider: %w", err)
	}

	r.logger.Debug("local provider succeeded",
		slog.Duration("latency", elapsed),
	)

	return resp, nil
}

// routeToCloudWithFallback attempts the cloud provider and falls back to local on failure.
func (r *Router) routeToCloudWithFallback(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// If no cloud provider, go directly to local
	if r.cloud == nil || !r.cloud.IsConfigured() {
		r.logger.Debug("cloud provider not configured, using local")
		return r.routeToLocal(ctx, req, "cloud not configured")
	}

	// Apply cloud-specific timeout if configured
	cloudCtx := ctx
	var cloudCancel context.CancelFunc
	if r.cloudTimeout > 0 {
		cloudCtx, cloudCancel = context.WithTimeout(ctx, r.cloudTimeout)
		defer cloudCancel()
	}

	// Attempt cloud with retries
	resp, err := r.executeWithRetry(cloudCtx, req, r.cloud.Chat)

	if err == nil {
		return resp, nil
	}

	r.logger.Warn("cloud provider failed",
		slog.String("error", err.Error()),
	)

	// Check if we should fall back
	if !r.enableFallback {
		return ChatResponse{}, fmt.Errorf("ai/router: cloud provider: %w", err)
	}

	// Don't fall back if the original context is canceled
	if ctx.Err() != nil {
		return ChatResponse{}, fmt.Errorf("ai/router: %w", ctx.Err())
	}

	// Check if local is available for fallback
	if r.local == nil || !r.local.IsAvailable() {
		return ChatResponse{}, fmt.Errorf("ai/router: cloud provider failed and no local fallback: %w", err)
	}

	r.logger.Info("falling back to local provider after cloud failure")

	// Update fallback metric
	r.mu.Lock()
	r.fallbackCount++
	r.mu.Unlock()

	// Attempt local fallback
	localResp, localErr := r.routeToLocal(ctx, req, "cloud fallback")
	if localErr != nil {
		// Return original cloud error if local also fails
		return ChatResponse{}, fmt.Errorf("ai/router: cloud failed (%w) and local fallback failed (%v)", err, localErr)
	}

	// Mark response as coming from fallback
	localResp.Source = ChatSourceFallback
	return localResp, nil
}

// executeWithRetry executes a chat function with retry logic for transient errors.
func (r *Router) executeWithRetry(
	ctx context.Context,
	req ChatRequest,
	fn func(context.Context, ChatRequest) (ChatResponse, error),
) (ChatResponse, error) {
	var lastErr error
	backoff := r.retryConfig.InitialBackoff

	for attempt := 0; attempt < r.retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			r.logger.Debug("retrying request",
				slog.Int("attempt", attempt+1),
				slog.Duration("backoff", backoff),
			)

			// Wait for backoff or context cancellation
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ChatResponse{}, ctx.Err()
			}

			// Increase backoff for next attempt
			backoff = time.Duration(float64(backoff) * r.retryConfig.BackoffMultiplier)
			if backoff > r.retryConfig.MaxBackoff {
				backoff = r.retryConfig.MaxBackoff
			}
		}

		startTime := time.Now()
		resp, err := fn(ctx, req)
		elapsed := time.Since(startTime)

		if err == nil {
			// Update metrics
			r.mu.Lock()
			r.cloudRequests++
			r.totalCloudTime += elapsed
			r.mu.Unlock()

			r.logger.Debug("cloud provider succeeded",
				slog.Duration("latency", elapsed),
				slog.Int("attempt", attempt+1),
			)

			return resp, nil
		}

		lastErr = err

		// Only retry on transient errors
		if !IsRetryableError(err) {
			break
		}
	}

	return ChatResponse{}, lastErr
}

// Metrics returns current router metrics.
func (r *Router) Metrics() RouterMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var avgCloudLatency, avgLocalLatency time.Duration
	if r.cloudRequests > 0 {
		avgCloudLatency = r.totalCloudTime / time.Duration(r.cloudRequests)
	}
	if r.localRequests > 0 {
		avgLocalLatency = r.totalLocalTime / time.Duration(r.localRequests)
	}

	return RouterMetrics{
		CloudRequests:       r.cloudRequests,
		LocalRequests:       r.localRequests,
		FallbackCount:       r.fallbackCount,
		AverageCloudLatency: avgCloudLatency,
		AverageLocalLatency: avgLocalLatency,
	}
}

// RouterMetrics contains runtime metrics for the router.
type RouterMetrics struct {
	CloudRequests       int64
	LocalRequests       int64
	FallbackCount       int64
	AverageCloudLatency time.Duration
	AverageLocalLatency time.Duration
}

// =============================================================================
// Simple Local Provider (Default Offline Implementation)
// =============================================================================

// SimpleLocalProvider is a basic LocalProvider implementation that provides
// offline responses. It's suitable for development, testing, and scenarios
// where sophisticated local inference isn't available.
//
// In production, consider implementing a more capable LocalProvider using
// Ollama, llama.cpp, or another local inference solution.
type SimpleLocalProvider struct {
	// Prefix is prepended to offline responses.
	// Defaults to "[OFFLINE] " if empty.
	Prefix string

	// Available controls whether this provider reports as available.
	// Defaults to true.
	available *bool

	mu        sync.RWMutex
	callCount int64
}

// Chat returns a simple offline acknowledgment response.
func (s *SimpleLocalProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	s.mu.Lock()
	s.callCount++
	s.mu.Unlock()

	// Check context
	if err := ctx.Err(); err != nil {
		return ChatResponse{}, err
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ChatResponse{}, err
	}

	// Check availability
	if !s.IsAvailable() {
		return ChatResponse{}, ErrProviderUnavailable
	}

	prefix := s.Prefix
	if prefix == "" {
		prefix = "[OFFLINE] "
	}

	return ChatResponse{
		Output:       prefix + "Currently operating in offline mode. Your request has been received: " + truncate(req.Prompt, 100),
		Source:       ChatSourceLocal,
		Model:        "offline-simple-v1",
		FinishReason: "stop",
		CreatedAt:    time.Now(),
	}, nil
}

// IsAvailable returns whether the provider is ready to serve requests.
func (s *SimpleLocalProvider) IsAvailable() bool {
	if s.available == nil {
		return true
	}
	return *s.available
}

// SetAvailable configures the availability state.
func (s *SimpleLocalProvider) SetAvailable(available bool) {
	s.available = &available
}

// CallCount returns the number of times Chat has been called.
func (s *SimpleLocalProvider) CallCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.callCount
}

// truncate shortens a string to the specified maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// =============================================================================
// Compile-time interface checks
// =============================================================================

var (
	_ LocalProvider = (*SimpleLocalProvider)(nil)
)
