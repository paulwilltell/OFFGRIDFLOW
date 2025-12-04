// Package factors provides PostgreSQL-backed emission factor storage.
package factors

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

// =============================================================================
// PostgreSQL Registry Configuration
// =============================================================================

// PostgresConfig configures the PostgreSQL factor registry.
type PostgresConfig struct {
	// DB is the database connection.
	DB *sql.DB

	// Logger for database operations.
	Logger *slog.Logger

	// TableName is the emission factors table name.
	TableName string

	// AutoMigrate creates the table if it doesn't exist.
	AutoMigrate bool

	// SeedDefaults populates default factors on first run.
	SeedDefaults bool
}

// DefaultPostgresConfig returns sensible defaults.
func DefaultPostgresConfig(db *sql.DB) PostgresConfig {
	return PostgresConfig{
		DB:           db,
		TableName:    "emission_factors",
		AutoMigrate:  true,
		SeedDefaults: true,
	}
}

// =============================================================================
// PostgreSQL Registry Implementation
// =============================================================================

// PostgresRegistry implements FactorRegistry backed by PostgreSQL.
//
// The registry stores emission factors in a PostgreSQL table and provides
// efficient lookup by scope, region, source, category, and unit.
//
// Example usage:
//
//	cfg := factors.DefaultPostgresConfig(db)
//	registry, err := factors.NewPostgresRegistry(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	factor, err := registry.FindFactor(ctx, emissions.FactorQuery{
//	    Scope:  emissions.Scope2,
//	    Region: "US-WEST",
//	    Unit:   "kWh",
//	})
type PostgresRegistry struct {
	db        *sql.DB
	tableName string
	logger    *slog.Logger
}

// NewPostgresRegistry creates a new PostgreSQL-backed factor registry.
func NewPostgresRegistry(cfg PostgresConfig) (*PostgresRegistry, error) {
	if cfg.DB == nil {
		return nil, errors.New("database connection is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	tableName := cfg.TableName
	if tableName == "" {
		tableName = "emission_factors"
	}

	r := &PostgresRegistry{
		db:        cfg.DB,
		tableName: tableName,
		logger:    logger,
	}

	if cfg.AutoMigrate {
		if err := r.Migrate(context.Background()); err != nil {
			return nil, fmt.Errorf("auto-migrate failed: %w", err)
		}
	}

	if cfg.SeedDefaults {
		if err := r.SeedDefaultFactors(context.Background()); err != nil {
			logger.Warn("failed to seed default factors",
				"error", err,
			)
		}
	}

	return r, nil
}

// NewPostgresRegistrySimple creates a registry with default configuration.
// Deprecated: Use NewPostgresRegistry with PostgresConfig instead.
func NewPostgresRegistrySimple(db *sql.DB) *PostgresRegistry {
	r, _ := NewPostgresRegistry(DefaultPostgresConfig(db))
	return r
}

// =============================================================================
// FactorRegistry Interface Implementation
// =============================================================================

// GetFactor retrieves a factor by its unique ID.
func (r *PostgresRegistry) GetFactor(ctx context.Context, id string) (emissions.EmissionFactor, error) {
	query := fmt.Sprintf(`
		SELECT id, scope, region, source, category, unit, 
		       value_kg_co2e_per_unit, method, data_source,
		       valid_from, valid_to, uncertainty_percent,
		       notes, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName)

	var f emissions.EmissionFactor
	var method, dataSource, notes sql.NullString
	var validFrom, validTo, createdAt, updatedAt sql.NullTime
	var uncertainty sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID,
		&f.Scope,
		&f.Region,
		&f.Source,
		&f.Category,
		&f.Unit,
		&f.ValueKgCO2ePerUnit,
		&method,
		&dataSource,
		&validFrom,
		&validTo,
		&uncertainty,
		&notes,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return emissions.EmissionFactor{}, fmt.Errorf(
			"factor %q: %w", id, emissions.ErrFactorNotFound,
		)
	}
	if err != nil {
		return emissions.EmissionFactor{}, fmt.Errorf("query factor: %w", err)
	}

	// Map nullable fields
	f.Method = emissions.CalculationMethod(method.String)
	f.DataSource = dataSource.String
	f.Notes = notes.String
	f.ValidFrom = validFrom.Time
	f.ValidTo = validTo.Time
	f.UncertaintyPercent = uncertainty.Float64
	f.CreatedAt = createdAt.Time
	f.UpdatedAt = updatedAt.Time

	return f, nil
}

// FindFactor looks up the best matching factor for an activity.
func (r *PostgresRegistry) FindFactor(ctx context.Context, query emissions.FactorQuery) (emissions.EmissionFactor, error) {
	// Build dynamic query with scoring for best match
	sqlQuery := fmt.Sprintf(`
		SELECT id, scope, region, source, category, unit, 
		       value_kg_co2e_per_unit, method, data_source,
		       valid_from, valid_to, uncertainty_percent,
		       notes, created_at, updated_at,
		       -- Scoring: more specific matches get higher scores
		       (CASE WHEN region = $1 THEN 100 ELSE 0 END) +
		       (CASE WHEN source = $2 THEN 50 ELSE 0 END) +
		       (CASE WHEN category = $3 OR $3 = '' THEN 25 ELSE 0 END) +
		       (CASE WHEN unit = $4 THEN 10 ELSE 0 END) AS score
		FROM %s
		WHERE ($1 = '' OR region = $1 OR region = 'GLOBAL')
		  AND ($2 = '' OR source = $2)
		  AND ($3 = '' OR category = $3 OR category = '' OR category IS NULL)
		  AND ($4 = '' OR unit = $4)
		  AND ($5 = 0 OR scope = $5)
		  AND ($6::timestamp IS NULL OR valid_from IS NULL OR valid_from <= $6)
		  AND ($6::timestamp IS NULL OR valid_to IS NULL OR valid_to >= $6)
		ORDER BY score DESC
		LIMIT 1
	`, r.tableName)

	var validAt interface{} = nil
	if !query.ValidAt.IsZero() {
		validAt = query.ValidAt
	}

	var f emissions.EmissionFactor
	var method, dataSource, notes sql.NullString
	var validFrom, validTo, createdAt, updatedAt sql.NullTime
	var uncertainty sql.NullFloat64
	var score int

	err := r.db.QueryRowContext(ctx, sqlQuery,
		query.Region,
		query.Source,
		query.Category,
		query.Unit,
		query.Scope,
		validAt,
	).Scan(
		&f.ID,
		&f.Scope,
		&f.Region,
		&f.Source,
		&f.Category,
		&f.Unit,
		&f.ValueKgCO2ePerUnit,
		&method,
		&dataSource,
		&validFrom,
		&validTo,
		&uncertainty,
		&notes,
		&createdAt,
		&updatedAt,
		&score,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return emissions.EmissionFactor{}, fmt.Errorf(
			"no factor matching scope=%s region=%q source=%q: %w",
			query.Scope, query.Region, query.Source,
			emissions.ErrFactorNotFound,
		)
	}
	if err != nil {
		return emissions.EmissionFactor{}, fmt.Errorf("find factor: %w", err)
	}

	// Map nullable fields
	f.Method = emissions.CalculationMethod(method.String)
	f.DataSource = dataSource.String
	f.Notes = notes.String
	f.ValidFrom = validFrom.Time
	f.ValidTo = validTo.Time
	f.UncertaintyPercent = uncertainty.Float64
	f.CreatedAt = createdAt.Time
	f.UpdatedAt = updatedAt.Time

	r.logger.Debug("found emission factor",
		"factor_id", f.ID,
		"score", score,
	)

	return f, nil
}

// ListFactors returns all factors matching the given criteria.
func (r *PostgresRegistry) ListFactors(ctx context.Context, query emissions.FactorQuery) ([]emissions.EmissionFactor, error) {
	sqlQuery := fmt.Sprintf(`
		SELECT id, scope, region, source, category, unit, 
		       value_kg_co2e_per_unit, method, data_source,
		       valid_from, valid_to, uncertainty_percent,
		       notes, created_at, updated_at
		FROM %s
		WHERE ($1 = '' OR region = $1)
		  AND ($2 = '' OR source = $2)
		  AND ($3 = '' OR category = $3)
		  AND ($4 = '' OR unit = $4)
		  AND ($5 = 0 OR scope = $5)
		ORDER BY scope, region, source, category
	`, r.tableName)

	rows, err := r.db.QueryContext(ctx, sqlQuery,
		query.Region,
		query.Source,
		query.Category,
		query.Unit,
		query.Scope,
	)
	if err != nil {
		return nil, fmt.Errorf("list factors: %w", err)
	}
	defer rows.Close()

	var factors []emissions.EmissionFactor

	for rows.Next() {
		var f emissions.EmissionFactor
		var method, dataSource, notes sql.NullString
		var validFrom, validTo, createdAt, updatedAt sql.NullTime
		var uncertainty sql.NullFloat64

		err := rows.Scan(
			&f.ID,
			&f.Scope,
			&f.Region,
			&f.Source,
			&f.Category,
			&f.Unit,
			&f.ValueKgCO2ePerUnit,
			&method,
			&dataSource,
			&validFrom,
			&validTo,
			&uncertainty,
			&notes,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan factor row: %w", err)
		}

		f.Method = emissions.CalculationMethod(method.String)
		f.DataSource = dataSource.String
		f.Notes = notes.String
		f.ValidFrom = validFrom.Time
		f.ValidTo = validTo.Time
		f.UncertaintyPercent = uncertainty.Float64
		f.CreatedAt = createdAt.Time
		f.UpdatedAt = updatedAt.Time

		factors = append(factors, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate factors: %w", err)
	}

	return factors, nil
}

// RegisterFactor adds or updates a factor in the registry.
func (r *PostgresRegistry) RegisterFactor(ctx context.Context, factor emissions.EmissionFactor) error {
	if !factor.IsValid() {
		return fmt.Errorf("invalid emission factor %q: missing required fields", factor.ID)
	}

	now := time.Now().UTC()
	if factor.CreatedAt.IsZero() {
		factor.CreatedAt = now
	}
	factor.UpdatedAt = now

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, scope, region, source, category, unit,
			value_kg_co2e_per_unit, method, data_source,
			valid_from, valid_to, uncertainty_percent,
			notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			scope = EXCLUDED.scope,
			region = EXCLUDED.region,
			source = EXCLUDED.source,
			category = EXCLUDED.category,
			unit = EXCLUDED.unit,
			value_kg_co2e_per_unit = EXCLUDED.value_kg_co2e_per_unit,
			method = EXCLUDED.method,
			data_source = EXCLUDED.data_source,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			uncertainty_percent = EXCLUDED.uncertainty_percent,
			notes = EXCLUDED.notes,
			updated_at = EXCLUDED.updated_at
	`, r.tableName)

	var validFrom, validTo interface{}
	if !factor.ValidFrom.IsZero() {
		validFrom = factor.ValidFrom
	}
	if !factor.ValidTo.IsZero() {
		validTo = factor.ValidTo
	}

	_, err := r.db.ExecContext(ctx, query,
		factor.ID,
		factor.Scope,
		factor.Region,
		factor.Source,
		factor.Category,
		factor.Unit,
		factor.ValueKgCO2ePerUnit,
		string(factor.Method),
		factor.DataSource,
		validFrom,
		validTo,
		factor.UncertaintyPercent,
		factor.Notes,
		factor.CreatedAt,
		factor.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert factor: %w", err)
	}

	r.logger.Info("registered emission factor",
		"factor_id", factor.ID,
		"scope", factor.Scope.String(),
		"region", factor.Region,
	)

	return nil
}

// =============================================================================
// Legacy Compatibility Methods
// =============================================================================

// GetScope2Factor returns the Scope 2 emission factor for a given region.
// Deprecated: Use FindFactor with FactorQuery instead.
func (r *PostgresRegistry) GetScope2Factor(region string) (emissions.EmissionFactor, bool) {
	f, err := r.FindFactor(context.Background(), emissions.FactorQuery{
		Scope:  emissions.Scope2,
		Region: region,
		Source: "electricity",
		Unit:   "kWh",
	})
	if err != nil {
		return emissions.EmissionFactor{}, false
	}
	return f, true
}

// Find looks up an emission factor by scope, category, region, and unit.
// Deprecated: Use FindFactor with FactorQuery instead.
func (r *PostgresRegistry) Find(scope, category, region, unit string) (emissions.EmissionFactor, error) {
	var s emissions.Scope
	switch scope {
	case "SCOPE1":
		s = emissions.Scope1
	case "SCOPE2":
		s = emissions.Scope2
	case "SCOPE3":
		s = emissions.Scope3
	}

	return r.FindFactor(context.Background(), emissions.FactorQuery{
		Scope:    s,
		Region:   region,
		Category: category,
		Unit:     unit,
	})
}

// =============================================================================
// Schema Management
// =============================================================================

// Migrate creates the emission_factors table if it doesn't exist.
func (r *PostgresRegistry) Migrate(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id                    VARCHAR(100) PRIMARY KEY,
			scope                 INTEGER NOT NULL,
			region                VARCHAR(50) NOT NULL,
			source                VARCHAR(100) NOT NULL,
			category              VARCHAR(100),
			unit                  VARCHAR(20) NOT NULL,
			value_kg_co2e_per_unit DECIMAL(12, 6) NOT NULL,
			method                VARCHAR(50),
			data_source           VARCHAR(255),
			valid_from            TIMESTAMP WITH TIME ZONE,
			valid_to              TIMESTAMP WITH TIME ZONE,
			uncertainty_percent   DECIMAL(5, 2),
			notes                 TEXT,
			created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_%s_scope_region 
			ON %s (scope, region);
		
		CREATE INDEX IF NOT EXISTS idx_%s_source_category 
			ON %s (source, category);
	`, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create emission_factors table: %w", err)
	}

	r.logger.Info("emission factors table migrated",
		"table", r.tableName,
	)

	return nil
}

// SeedDefaultFactors inserts default emission factors if none exist.
func (r *PostgresRegistry) SeedDefaultFactors(ctx context.Context) error {
	// Check if factors already exist
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE scope = $1", r.tableName)
	err := r.db.QueryRowContext(ctx, query, emissions.Scope2).Scan(&count)
	if err != nil {
		return fmt.Errorf("check existing factors: %w", err)
	}

	if count > 0 {
		r.logger.Debug("emission factors already seeded",
			"existing_count", count,
		)
		return nil
	}

	// Create an in-memory registry to get defaults
	memRegistry := NewInMemoryRegistry(RegistryConfig{
		PreloadDefaults: true,
		Logger:          r.logger,
	})

	// Get all default factors
	factors, err := memRegistry.ListFactors(ctx, emissions.FactorQuery{})
	if err != nil {
		return fmt.Errorf("get default factors: %w", err)
	}

	// Insert each factor
	for _, f := range factors {
		if err := r.RegisterFactor(ctx, f); err != nil {
			r.logger.Warn("failed to seed factor",
				"factor_id", f.ID,
				"error", err,
			)
		}
	}

	r.logger.Info("seeded default emission factors",
		"count", len(factors),
	)

	return nil
}

// Count returns the number of factors in the registry.
func (r *PostgresRegistry) Count(ctx context.Context) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName)
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count factors: %w", err)
	}
	return count, nil
}

// DeleteFactor removes a factor by ID.
func (r *PostgresRegistry) DeleteFactor(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete factor: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("factor %q: %w", id, emissions.ErrFactorNotFound)
	}

	r.logger.Info("deleted emission factor",
		"factor_id", id,
	)

	return nil
}
