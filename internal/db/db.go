// Package db provides PostgreSQL database connectivity and migration support
// for OffGridFlow. It wraps the standard database/sql package with connection
// pooling, health checks, and migration management.
//
// Features:
//   - Connection pooling with configurable limits
//   - Automatic connection health verification
//   - Embedded schema migrations
//   - Context-aware query execution
//   - Structured error handling
//
// Usage:
//
//	db, err := db.Connect(ctx, db.Config{
//	    DSN: "postgres://user:pass@localhost:5432/offgridflow",
//	})
//	if err != nil {
//	    log.Fatalf("database connection failed: %v", err)
//	}
//	defer db.Close()
//
//	if err := db.RunMigrations(ctx); err != nil {
//	    log.Fatalf("migration failed: %v", err)
//	}
package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// =============================================================================
// Embedded Schema
// =============================================================================

//go:embed schema.sql
var schemaSQL string

// =============================================================================
// Configuration Constants
// =============================================================================

const (
	// defaultMaxOpenConns is the default maximum number of open connections.
	defaultMaxOpenConns = 25

	// defaultMaxIdleConns is the default maximum number of idle connections.
	defaultMaxIdleConns = 10

	// defaultConnMaxLifetime is the default maximum connection lifetime.
	defaultConnMaxLifetime = 45 * time.Minute

	// defaultConnMaxIdleTime is the default maximum idle time for a connection.
	defaultConnMaxIdleTime = 15 * time.Minute

	// defaultConnectTimeout is the default timeout for initial connection.
	defaultConnectTimeout = 10 * time.Second

	// defaultPingTimeout is the default timeout for health checks.
	defaultPingTimeout = 5 * time.Second
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrEmptyDSN is returned when the DSN is empty or whitespace-only.
	ErrEmptyDSN = errors.New("db: empty DSN")

	// ErrNilConnection is returned when a nil connection is passed.
	ErrNilConnection = errors.New("db: nil connection")

	// ErrEmptySchema is returned when the embedded schema is empty.
	ErrEmptySchema = errors.New("db: empty schema SQL")

	// ErrConnectionFailed is returned when the database connection fails.
	ErrConnectionFailed = errors.New("db: connection failed")

	// ErrMigrationFailed is returned when schema migration fails.
	ErrMigrationFailed = errors.New("db: migration failed")

	// ErrAlreadyClosed is returned when operating on a closed connection pool.
	ErrAlreadyClosed = errors.New("db: connection pool already closed")
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds database connection configuration.
type Config struct {
	// DSN is the PostgreSQL connection string.
	// Format: postgres://user:pass@host:port/database?sslmode=disable
	DSN string

	// MaxOpenConns is the maximum number of open connections.
	// Defaults to 25 if zero.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections.
	// Defaults to 10 if zero.
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	// Defaults to 45 minutes if zero.
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// Defaults to 15 minutes if zero.
	ConnMaxIdleTime time.Duration

	// ConnectTimeout is the maximum time to wait for initial connection.
	// Defaults to 10 seconds if zero.
	ConnectTimeout time.Duration

	// PingTimeout is the timeout for health check pings.
	// Defaults to 5 seconds if zero.
	PingTimeout time.Duration
}

// applyDefaults fills in default values for unset fields.
func (c *Config) applyDefaults() {
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = defaultMaxOpenConns
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = defaultMaxIdleConns
	}
	if c.ConnMaxLifetime <= 0 {
		c.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = defaultConnMaxIdleTime
	}
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = defaultPingTimeout
	}
}

// validate checks the configuration for errors.
func (c *Config) validate() error {
	if strings.TrimSpace(c.DSN) == "" {
		return ErrEmptyDSN
	}

	// Ensure MaxIdleConns doesn't exceed MaxOpenConns
	if c.MaxIdleConns > c.MaxOpenConns {
		c.MaxIdleConns = c.MaxOpenConns
	}

	return nil
}

// =============================================================================
// Database Connection
// =============================================================================

// DB wraps sql.DB with additional functionality for OffGridFlow.
type DB struct {
	*sql.DB
	config Config

	mu     sync.RWMutex
	closed bool
	stats  ConnectionStats
}

// ConnectionStats tracks connection pool statistics.
type ConnectionStats struct {
	// ConnectTime is when the connection was established.
	ConnectTime time.Time

	// LastPingTime is when the last successful health check occurred.
	LastPingTime time.Time

	// MigrationsRun indicates whether migrations have been applied.
	MigrationsRun bool

	// MigrationTime is when migrations were last applied.
	MigrationTime time.Time
}

// Connect opens a PostgreSQL connection pool with the given configuration.
// It verifies connectivity before returning and applies sensible defaults.
func Connect(ctx context.Context, cfg Config) (*DB, error) {
	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Apply connect timeout to context
	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	// Open connection pool
	sqlDB, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connectivity
	if err := sqlDB.PingContext(connectCtx); err != nil {
		_ = sqlDB.Close() // Clean up on failure
		return nil, fmt.Errorf("%w: ping failed: %v", ErrConnectionFailed, err)
	}

	now := time.Now()
	return &DB{
		DB:     sqlDB,
		config: cfg,
		stats: ConnectionStats{
			ConnectTime:  now,
			LastPingTime: now,
		},
	}, nil
}

// ConnectWithDSN is a convenience function for simple DSN-only connections.
// It uses default configuration for all other settings.
func ConnectWithDSN(ctx context.Context, dsn string) (*DB, error) {
	return Connect(ctx, Config{DSN: dsn})
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	if db == nil {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrAlreadyClosed
	}

	db.closed = true
	return db.DB.Close()
}

// HealthCheck performs a lightweight database health check.
// Returns an error if the database is unreachable or the check times out.
func (db *DB) HealthCheck(ctx context.Context) error {
	if db == nil {
		return ErrNilConnection
	}

	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return ErrAlreadyClosed
	}
	db.mu.RUnlock()

	pingCtx, cancel := context.WithTimeout(ctx, db.config.PingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("db: health check failed: %w", err)
	}

	db.mu.Lock()
	db.stats.LastPingTime = time.Now()
	db.mu.Unlock()

	return nil
}

// Stats returns connection pool statistics.
func (db *DB) Stats() (ConnectionStats, sql.DBStats) {
	if db == nil {
		return ConnectionStats{}, sql.DBStats{}
	}

	db.mu.RLock()
	connStats := db.stats
	db.mu.RUnlock()

	return connStats, db.DB.Stats()
}

// IsClosed returns true if the connection pool has been closed.
func (db *DB) IsClosed() bool {
	if db == nil {
		return true
	}

	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.closed
}

// =============================================================================
// Migrations
// =============================================================================

// RunMigrations executes the embedded SQL schema.
// For production-grade migrations, consider using golang-migrate or Atlas,
// but this approach keeps development setups simple and deterministic.
//
// Note: This function is idempotent if the schema uses IF NOT EXISTS clauses.
func (db *DB) RunMigrations(ctx context.Context) error {
	if db == nil {
		return ErrNilConnection
	}

	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return ErrAlreadyClosed
	}
	db.mu.RUnlock()

	schema := strings.TrimSpace(schemaSQL)
	if schema == "" {
		return ErrEmptySchema
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	db.mu.Lock()
	db.stats.MigrationsRun = true
	db.stats.MigrationTime = time.Now()
	db.mu.Unlock()

	return nil
}

// RunMigrationsLegacy applies migrations using a raw sql.DB connection.
// Deprecated: Use DB.RunMigrations instead.
func RunMigrationsLegacy(ctx context.Context, sqlDB *sql.DB) error {
	if sqlDB == nil {
		return ErrNilConnection
	}

	schema := strings.TrimSpace(schemaSQL)
	if schema == "" {
		return ErrEmptySchema
	}

	if _, err := sqlDB.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	return nil
}

// =============================================================================
// Transaction Helpers
// =============================================================================

// TxFunc is a function that runs within a transaction.
type TxFunc func(tx *sql.Tx) error

// WithTx executes a function within a database transaction.
// The transaction is committed if the function returns nil,
// otherwise it is rolled back.
func (db *DB) WithTx(ctx context.Context, fn TxFunc) error {
	if db == nil {
		return ErrNilConnection
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("db: rollback failed after error (%v): %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit: %w", err)
	}

	return nil
}

// WithTxOptions executes a function within a transaction with custom options.
func (db *DB) WithTxOptions(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error {
	if db == nil {
		return ErrNilConnection
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("db: begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("db: rollback failed after error (%v): %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit: %w", err)
	}

	return nil
}

// =============================================================================
// Query Helpers
// =============================================================================

// QueryRowContext is a convenience wrapper that scans a single row result.
// The scanner function receives the row and should call row.Scan(...).
func (db *DB) QueryRowFunc(ctx context.Context, query string, scanner func(*sql.Row) error, args ...any) error {
	if db == nil {
		return ErrNilConnection
	}

	row := db.QueryRowContext(ctx, query, args...)
	return scanner(row)
}

// Exists checks if at least one row matches the query.
func (db *DB) Exists(ctx context.Context, query string, args ...any) (bool, error) {
	if db == nil {
		return false, ErrNilConnection
	}

	var exists bool
	err := db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("db: exists query: %w", err)
	}
	return exists, nil
}

// =============================================================================
// Error Helpers
// =============================================================================

// IsNotFound returns true if the error indicates no rows were found.
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 23505 is unique_violation
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "unique constraint")
}

// IsForeignKeyViolation checks if the error is a PostgreSQL foreign key violation.
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 23503 is foreign_key_violation
	return strings.Contains(err.Error(), "23503") ||
		strings.Contains(err.Error(), "foreign key constraint")
}

// =============================================================================
// Legacy Compatibility
// =============================================================================

// LegacyConnect opens a PostgreSQL connection using the old interface.
// Deprecated: Use Connect with Config instead.
func LegacyConnect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := ConnectWithDSN(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return db.DB, nil
}

// RunMigrations is a legacy function for backward compatibility.
// Deprecated: Use DB.RunMigrations instead.
func RunMigrations(ctx context.Context, dbConn *sql.DB) error {
	return RunMigrationsLegacy(ctx, dbConn)
}
