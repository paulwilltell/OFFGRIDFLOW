package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PostgresActivityStore implements ActivityStore backed by PostgreSQL.
type PostgresActivityStore struct {
	db *sql.DB
}

// NewPostgresActivityStore creates a new PostgreSQL-backed activity store.
func NewPostgresActivityStore(db *sql.DB) *PostgresActivityStore {
	return &PostgresActivityStore{db: db}
}

// Save stores a single activity in the database.
func (s *PostgresActivityStore) Save(ctx context.Context, activity Activity) error {
	if activity.ID == "" {
		activity.ID = uuid.NewString()
	}
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}

	metaJSON, _ := json.Marshal(activity.Metadata)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO activities (
			id, source, category, meter_id, location,
			period_start, period_end, quantity, unit,
			org_id, metadata, workspace_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET
			source = EXCLUDED.source,
			category = EXCLUDED.category,
			meter_id = EXCLUDED.meter_id,
			location = EXCLUDED.location,
			period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			quantity = EXCLUDED.quantity,
			unit = EXCLUDED.unit,
			org_id = EXCLUDED.org_id,
			metadata = EXCLUDED.metadata,
			workspace_id = EXCLUDED.workspace_id,
			created_at = EXCLUDED.created_at
	`,
		activity.ID,
		activity.Source,
		activity.Category,
		activity.MeterID,
		activity.Location,
		activity.PeriodStart,
		activity.PeriodEnd,
		activity.Quantity,
		activity.Unit,
		activity.OrgID,
		metaJSON,
		activity.WorkspaceID,
		activity.CreatedAt,
	)
	return err
}

// SaveBatch stores multiple activities in a transaction.
func (s *PostgresActivityStore) SaveBatch(ctx context.Context, activities []Activity) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for i := range activities {
		if activities[i].ID == "" {
			activities[i].ID = uuid.NewString()
		}
		if activities[i].CreatedAt.IsZero() {
			activities[i].CreatedAt = time.Now()
		}
		metaJSON, _ := json.Marshal(activities[i].Metadata)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO activities (
				id, source, category, meter_id, location,
				period_start, period_end, quantity, unit,
				org_id, metadata, workspace_id, created_at
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id) DO UPDATE SET
				source = EXCLUDED.source,
				category = EXCLUDED.category,
				meter_id = EXCLUDED.meter_id,
				location = EXCLUDED.location,
				period_start = EXCLUDED.period_start,
				period_end = EXCLUDED.period_end,
				quantity = EXCLUDED.quantity,
				unit = EXCLUDED.unit,
				org_id = EXCLUDED.org_id,
				metadata = EXCLUDED.metadata,
				workspace_id = EXCLUDED.workspace_id,
				created_at = EXCLUDED.created_at
		`,
			activities[i].ID,
			activities[i].Source,
			activities[i].Category,
			activities[i].MeterID,
			activities[i].Location,
			activities[i].PeriodStart,
			activities[i].PeriodEnd,
			activities[i].Quantity,
			activities[i].Unit,
			activities[i].OrgID,
			metaJSON,
			activities[i].WorkspaceID,
			activities[i].CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// List returns all activities (limited to 1000 most recent).
func (s *PostgresActivityStore) List(ctx context.Context) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, category, meter_id, location,
		       period_start, period_end, quantity, unit,
		       org_id, metadata, workspace_id, created_at
		FROM activities
		ORDER BY created_at DESC
		LIMIT 1000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivities(rows)
}

// ListBySource returns activities filtered by source type.
func (s *PostgresActivityStore) ListBySource(ctx context.Context, source string) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, category, meter_id, location,
		       period_start, period_end, quantity, unit,
		       org_id, metadata, workspace_id, created_at
		FROM activities
		WHERE source = $1
		ORDER BY created_at DESC
		LIMIT 1000
	`, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivities(rows)
}

// ListByOrgAndSource returns activities filtered by organization and source type.
func (s *PostgresActivityStore) ListByOrgAndSource(ctx context.Context, orgID, source string) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, category, meter_id, location,
		       period_start, period_end, quantity, unit,
		       org_id, metadata, workspace_id, created_at
		FROM activities
		WHERE org_id = $1 AND source = $2
		ORDER BY created_at DESC
		LIMIT 1000
	`, orgID, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivities(rows)
}

// ListByOrg returns activities filtered by organization.
func (s *PostgresActivityStore) ListByOrg(ctx context.Context, orgID string) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, category, meter_id, location,
		       period_start, period_end, quantity, unit,
		       org_id, metadata, workspace_id, created_at
		FROM activities
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT 1000
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivities(rows)
}

// ListRecent returns activities created after the given time.
func (s *PostgresActivityStore) ListRecent(ctx context.Context, since time.Time) ([]Activity, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, category, meter_id, location,
		       period_start, period_end, quantity, unit,
		       org_id, metadata, workspace_id, created_at
		FROM activities
		WHERE created_at > $1
		ORDER BY created_at DESC
		LIMIT 1000
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanActivities(rows)
}

// scanActivities is a helper to scan activity rows into a slice.
func scanActivities(rows *sql.Rows) ([]Activity, error) {
	var result []Activity
	for rows.Next() {
		var a Activity
		var metaJSON []byte
		var category, meterID, location, unit, orgID, workspaceID sql.NullString
		var periodStart, periodEnd sql.NullTime
		var quantity sql.NullFloat64

		if err := rows.Scan(
			&a.ID,
			&a.Source,
			&category,
			&meterID,
			&location,
			&periodStart,
			&periodEnd,
			&quantity,
			&unit,
			&orgID,
			&metaJSON,
			&workspaceID,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}

		a.Category = category.String
		a.MeterID = meterID.String
		a.Location = location.String
		a.Unit = unit.String
		a.OrgID = orgID.String
		a.WorkspaceID = workspaceID.String
		if periodStart.Valid {
			a.PeriodStart = periodStart.Time
		}
		if periodEnd.Valid {
			a.PeriodEnd = periodEnd.Time
		}
		if quantity.Valid {
			a.Quantity = quantity.Float64
		}

		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &a.Metadata)
		}

		result = append(result, a)
	}
	return result, rows.Err()
}
