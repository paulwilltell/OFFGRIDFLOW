package offgrid

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// Configuration Constants
// =============================================================================

const (
	// defaultCheckInterval is how often connectivity is checked.
	defaultCheckInterval = 30 * time.Second

	// defaultCheckTimeout is the maximum time for a single connectivity check.
	defaultCheckTimeout = 5 * time.Second

	// defaultConsecutiveFailures is how many failures before going offline.
	defaultConsecutiveFailures = 2

	// defaultConsecutiveSuccesses is how many successes before going online.
	defaultConsecutiveSuccesses = 1
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrWatcherStopped is returned when operations are attempted on a stopped watcher.
	ErrWatcherStopped = errors.New("offgrid: connectivity watcher stopped")

	// ErrWatcherAlreadyRunning is returned when Start is called on a running watcher.
	ErrWatcherAlreadyRunning = errors.New("offgrid: connectivity watcher already running")
)

// =============================================================================
// Connectivity Checker Interface
// =============================================================================

// ConnectivityChecker defines how to check network connectivity.
// Implementations can use DNS, HTTP, TCP, or any other mechanism.
type ConnectivityChecker interface {
	// Check returns true if network connectivity is available.
	// The context should be respected for cancellation and timeout.
	Check(ctx context.Context) bool
}

// =============================================================================
// Built-in Checker Implementations
// =============================================================================

// DNSChecker verifies connectivity by attempting a UDP connection to a DNS server.
// This is fast and lightweight, suitable for frequent polling.
type DNSChecker struct {
	// Host is the DNS server address (e.g., "8.8.8.8").
	Host string

	// Port is the DNS port (typically "53").
	Port string

	// Timeout is the maximum time for the connection attempt.
	Timeout time.Duration
}

// DefaultDNSChecker returns a DNSChecker configured for Google's public DNS.
func DefaultDNSChecker() *DNSChecker {
	return &DNSChecker{
		Host:    "8.8.8.8",
		Port:    "53",
		Timeout: 2 * time.Second,
	}
}

// Check attempts a UDP connection to the configured DNS server.
func (c *DNSChecker) Check(ctx context.Context) bool {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	// Respect context deadline if shorter than our timeout
	if deadline, ok := ctx.Deadline(); ok {
		if ctxTimeout := time.Until(deadline); ctxTimeout < timeout {
			timeout = ctxTimeout
		}
	}

	// Check context before attempting connection
	if ctx.Err() != nil {
		return false
	}

	addr := net.JoinHostPort(c.Host, c.Port)
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// HTTPChecker verifies connectivity by making an HTTP HEAD request.
// This is more thorough than DNS but slower and uses more resources.
type HTTPChecker struct {
	// URL is the endpoint to check (e.g., "https://www.google.com/generate_204").
	URL string

	// Timeout is the maximum time for the HTTP request.
	Timeout time.Duration

	// ExpectedStatus is the expected HTTP status code (0 means any 2xx).
	ExpectedStatus int

	// client is reused across checks for connection pooling.
	client     *http.Client
	clientOnce sync.Once
}

// DefaultHTTPChecker returns an HTTPChecker configured for Google's connectivity check endpoint.
func DefaultHTTPChecker() *HTTPChecker {
	return &HTTPChecker{
		URL:            "https://www.google.com/generate_204",
		Timeout:        5 * time.Second,
		ExpectedStatus: http.StatusNoContent,
	}
}

// Check performs an HTTP HEAD request to verify connectivity.
func (c *HTTPChecker) Check(ctx context.Context) bool {
	c.clientOnce.Do(func() {
		timeout := c.Timeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		c.client = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 2,
			},
			// Don't follow redirects for connectivity checks
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	})

	if ctx.Err() != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.URL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "OffGridFlow-ConnectivityCheck/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if c.ExpectedStatus > 0 {
		return resp.StatusCode == c.ExpectedStatus
	}
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// MultiChecker combines multiple checkers and requires any/all to pass.
type MultiChecker struct {
	// Checkers is the list of connectivity checkers to use.
	Checkers []ConnectivityChecker

	// RequireAll if true, requires all checkers to pass.
	// If false (default), any single checker passing is sufficient.
	RequireAll bool
}

// Check runs all configured checkers according to the RequireAll policy.
func (c *MultiChecker) Check(ctx context.Context) bool {
	if len(c.Checkers) == 0 {
		return true // No checkers = assume online
	}

	if c.RequireAll {
		for _, checker := range c.Checkers {
			if !checker.Check(ctx) {
				return false
			}
		}
		return true
	}

	// Any checker passing is sufficient
	for _, checker := range c.Checkers {
		if checker.Check(ctx) {
			return true
		}
	}
	return false
}

// =============================================================================
// Watcher Configuration
// =============================================================================

// WatcherConfig holds configuration for the ConnectivityWatcher.
type WatcherConfig struct {
	// Checker is the connectivity checking strategy.
	// If nil, DefaultDNSChecker() is used.
	Checker ConnectivityChecker

	// Interval is how often to check connectivity.
	// Defaults to 30 seconds if zero.
	Interval time.Duration

	// CheckTimeout is the maximum time for each connectivity check.
	// Defaults to 5 seconds if zero.
	CheckTimeout time.Duration

	// ConsecutiveFailures is how many consecutive failures before transitioning offline.
	// Defaults to 2 if zero. This prevents brief network hiccups from triggering offline mode.
	ConsecutiveFailures int

	// ConsecutiveSuccesses is how many consecutive successes before transitioning online.
	// Defaults to 1 if zero.
	ConsecutiveSuccesses int

	// Logger is used for operational logging.
	// If nil, slog.Default() is used.
	Logger *slog.Logger

	// OnCheckComplete is called after each connectivity check with the result.
	// This is useful for metrics and monitoring.
	OnCheckComplete func(online bool, latency time.Duration)
}

// DefaultWatcherConfig returns a WatcherConfig with sensible defaults.
func DefaultWatcherConfig() WatcherConfig {
	return WatcherConfig{
		Checker:              DefaultDNSChecker(),
		Interval:             defaultCheckInterval,
		CheckTimeout:         defaultCheckTimeout,
		ConsecutiveFailures:  defaultConsecutiveFailures,
		ConsecutiveSuccesses: defaultConsecutiveSuccesses,
	}
}

func (c *WatcherConfig) applyDefaults() {
	if c.Checker == nil {
		c.Checker = DefaultDNSChecker()
	}
	if c.Interval <= 0 {
		c.Interval = defaultCheckInterval
	}
	if c.CheckTimeout <= 0 {
		c.CheckTimeout = defaultCheckTimeout
	}
	if c.ConsecutiveFailures <= 0 {
		c.ConsecutiveFailures = defaultConsecutiveFailures
	}
	if c.ConsecutiveSuccesses <= 0 {
		c.ConsecutiveSuccesses = defaultConsecutiveSuccesses
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}

// =============================================================================
// ConnectivityWatcher Implementation
// =============================================================================

// ConnectivityWatcher monitors network connectivity and automatically updates
// a ModeManager when the network state changes. It uses configurable checking
// strategies and hysteresis to avoid flapping.
//
// Thread safety: All methods are safe for concurrent use.
//
// Example:
//
//	mm := offgrid.NewModeManager(offgrid.ModeOnline)
//	watcher := offgrid.NewConnectivityWatcher(mm, offgrid.DefaultWatcherConfig())
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	go watcher.Start(ctx)
type ConnectivityWatcher struct {
	modeManager *ModeManager
	config      WatcherConfig

	// State
	running         atomic.Bool
	consecutiveFail atomic.Int32
	consecutivePass atomic.Int32

	// Metrics
	totalChecks  atomic.Int64
	failedChecks atomic.Int64
	lastCheck    atomic.Value // time.Time
	lastResult   atomic.Bool
}

// NewConnectivityWatcher creates a new watcher with the given configuration.
// The ModeManager must not be nil.
func NewConnectivityWatcher(mm *ModeManager, cfg WatcherConfig) *ConnectivityWatcher {
	if mm == nil {
		panic("offgrid: NewConnectivityWatcher requires non-nil ModeManager")
	}

	cfg.applyDefaults()

	w := &ConnectivityWatcher{
		modeManager: mm,
		config:      cfg,
	}

	w.lastCheck.Store(time.Time{})
	return w
}

// Start begins the connectivity monitoring loop.
// It blocks until the context is canceled.
// Returns an error if the watcher is already running.
func (w *ConnectivityWatcher) Start(ctx context.Context) error {
	// Prevent multiple starts
	if !w.running.CompareAndSwap(false, true) {
		return ErrWatcherAlreadyRunning
	}
	defer w.running.Store(false)

	w.config.Logger.Info("connectivity watcher starting",
		slog.Duration("interval", w.config.Interval),
		slog.Int("consecutive_failures_threshold", w.config.ConsecutiveFailures),
	)

	// Perform initial check immediately
	w.performCheck(ctx)

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.config.Logger.Info("connectivity watcher shutting down",
				slog.String("reason", ctx.Err().Error()),
			)
			return nil
		case <-ticker.C:
			w.performCheck(ctx)
		}
	}
}

// performCheck executes a single connectivity check and updates state.
func (w *ConnectivityWatcher) performCheck(ctx context.Context) {
	// Create timeout context for this check
	checkCtx, cancel := context.WithTimeout(ctx, w.config.CheckTimeout)
	defer cancel()

	startTime := time.Now()
	online := w.config.Checker.Check(checkCtx)
	latency := time.Since(startTime)

	// Update metrics
	w.totalChecks.Add(1)
	w.lastCheck.Store(time.Now())
	w.lastResult.Store(online)

	if !online {
		w.failedChecks.Add(1)
	}

	// Call optional callback
	if w.config.OnCheckComplete != nil {
		w.config.OnCheckComplete(online, latency)
	}

	// Update consecutive counters
	if online {
		w.consecutiveFail.Store(0)
		passes := w.consecutivePass.Add(1)

		// Check if we should transition to online
		if passes >= int32(w.config.ConsecutiveSuccesses) {
			current := w.modeManager.GetMode()
			if current != ModeOnline {
				w.config.Logger.Info("connectivity restored, transitioning to online mode",
					slog.Int("consecutive_successes", int(passes)),
					slog.Duration("check_latency", latency),
				)
				w.modeManager.SetMode(ModeOnline)
			}
		}
	} else {
		w.consecutivePass.Store(0)
		fails := w.consecutiveFail.Add(1)

		// Check if we should transition to offline
		if fails >= int32(w.config.ConsecutiveFailures) {
			current := w.modeManager.GetMode()
			if current != ModeOffline {
				w.config.Logger.Warn("connectivity lost, transitioning to offline mode",
					slog.Int("consecutive_failures", int(fails)),
					slog.Duration("check_latency", latency),
				)
				w.modeManager.SetMode(ModeOffline)
			}
		} else {
			w.config.Logger.Debug("connectivity check failed",
				slog.Int("consecutive_failures", int(fails)),
				slog.Int("threshold", w.config.ConsecutiveFailures),
			)
		}
	}
}

// ForceCheck triggers an immediate connectivity check outside the normal interval.
// This is useful when you have reason to believe connectivity has changed.
func (w *ConnectivityWatcher) ForceCheck(ctx context.Context) {
	w.performCheck(ctx)
}

// IsRunning returns true if the watcher is currently running.
func (w *ConnectivityWatcher) IsRunning() bool {
	return w.running.Load()
}

// Stats returns current watcher statistics.
func (w *ConnectivityWatcher) Stats() WatcherStats {
	lastCheck, _ := w.lastCheck.Load().(time.Time)

	return WatcherStats{
		TotalChecks:          w.totalChecks.Load(),
		FailedChecks:         w.failedChecks.Load(),
		LastCheck:            lastCheck,
		LastResult:           w.lastResult.Load(),
		ConsecutiveFailures:  int(w.consecutiveFail.Load()),
		ConsecutiveSuccesses: int(w.consecutivePass.Load()),
		IsRunning:            w.running.Load(),
	}
}

// WatcherStats contains statistics about the ConnectivityWatcher.
type WatcherStats struct {
	// TotalChecks is the total number of connectivity checks performed.
	TotalChecks int64 `json:"total_checks"`

	// FailedChecks is the number of checks that indicated offline status.
	FailedChecks int64 `json:"failed_checks"`

	// LastCheck is the timestamp of the most recent check.
	LastCheck time.Time `json:"last_check"`

	// LastResult is the result of the most recent check (true = online).
	LastResult bool `json:"last_result"`

	// ConsecutiveFailures is the current streak of failed checks.
	ConsecutiveFailures int `json:"consecutive_failures"`

	// ConsecutiveSuccesses is the current streak of successful checks.
	ConsecutiveSuccesses int `json:"consecutive_successes"`

	// IsRunning indicates whether the watcher is currently active.
	IsRunning bool `json:"is_running"`
}

// SuccessRate returns the percentage of successful checks (0.0 to 1.0).
func (s WatcherStats) SuccessRate() float64 {
	if s.TotalChecks == 0 {
		return 1.0 // No checks = assume healthy
	}
	return float64(s.TotalChecks-s.FailedChecks) / float64(s.TotalChecks)
}

// =============================================================================
// Compile-time interface checks
// =============================================================================

var (
	_ ConnectivityChecker = (*DNSChecker)(nil)
	_ ConnectivityChecker = (*HTTPChecker)(nil)
	_ ConnectivityChecker = (*MultiChecker)(nil)
)

// =============================================================================
// Convenience Functions
// =============================================================================

// CheckConnectivityOnce performs a single connectivity check using the default DNS checker.
// This is useful for one-off checks without setting up a watcher.
func CheckConnectivityOnce(ctx context.Context) bool {
	return DefaultDNSChecker().Check(ctx)
}

// CheckConnectivityWithHTTP performs a single connectivity check using HTTP.
// This is more thorough but slower than DNS checks.
func CheckConnectivityWithHTTP(ctx context.Context) bool {
	return DefaultHTTPChecker().Check(ctx)
}

// MustNewConnectivityWatcher is like NewConnectivityWatcher but panics on nil ModeManager.
// This is useful in initialization code where a nil ModeManager is a programming error.
func MustNewConnectivityWatcher(mm *ModeManager, cfg WatcherConfig) *ConnectivityWatcher {
	if mm == nil {
		panic("offgrid: MustNewConnectivityWatcher requires non-nil ModeManager")
	}
	return NewConnectivityWatcher(mm, cfg)
}
