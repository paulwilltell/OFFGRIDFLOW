package worker

import (
	"os"
	"strconv"
	"time"
)

// Config controls worker scheduling and retry behavior.
type Config struct {
	IngestionInterval time.Duration
	RecalcInterval    time.Duration
	AlertInterval     time.Duration
	DefaultTimeout    time.Duration
	DefaultRetryLimit int
	DefaultBackoff    time.Duration
	DefaultBackoffMax time.Duration
	DefaultJitter     time.Duration
}

// DefaultConfig provides sensible production defaults.
func DefaultConfig() Config {
	return Config{
		IngestionInterval:  30 * time.Minute,
		RecalcInterval:     24 * time.Hour,
		AlertInterval:      5 * time.Minute,
		DefaultTimeout:     2 * time.Minute,
		DefaultRetryLimit:  3,
		DefaultBackoff:     2 * time.Second,
		DefaultBackoffMax:  30 * time.Second,
		DefaultJitter:      5 * time.Second,
	}
}

// FromEnv populates Config using environment variables if present.
func FromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("WORKER_INGESTION_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.IngestionInterval = d
		}
	}
	if v := os.Getenv("WORKER_RECALC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RecalcInterval = d
		}
	}
	if v := os.Getenv("WORKER_ALERT_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.AlertInterval = d
		}
	}
	if v := os.Getenv("WORKER_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultTimeout = d
		}
	}
	if v := os.Getenv("WORKER_RETRY_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.DefaultRetryLimit = n
		}
	}
	if v := os.Getenv("WORKER_BACKOFF_INITIAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultBackoff = d
		}
	}
	if v := os.Getenv("WORKER_BACKOFF_MAX"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultBackoffMax = d
		}
	}
	if v := os.Getenv("WORKER_JITTER"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultJitter = d
		}
	}

	return cfg
}
