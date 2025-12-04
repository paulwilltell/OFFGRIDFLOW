package connectors

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Note: This is a minimal smoke test using an in-memory SQLite substitute is not available;
// We only ensure the interface works with a nil DB using expected errors. Full integration
// would require a Postgres test harness.

func TestPostgresStore_ListWithNilDB(t *testing.T) {
	store := NewPostgresStore(&sql.DB{})
	// Expect a panic or error; ensure we don't silently succeed
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic with nil db")
		}
	}()
	_, _ = store.List(context.Background(), "")
}
