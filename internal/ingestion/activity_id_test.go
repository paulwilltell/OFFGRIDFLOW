package ingestion

import (
	"testing"

	"github.com/google/uuid"
)

// coerceActivityID must turn parser-generated human-readable IDs into valid,
// deterministic UUIDs (the activities table id column is UUID) — otherwise every
// CSV upload fails against Postgres.
func TestCoerceActivityID(t *testing.T) {
	// Non-UUID -> valid UUID, deterministic, original preserved.
	a := Activity{ID: "waste-2"}
	coerceActivityID(&a)
	if _, err := uuid.Parse(a.ID); err != nil {
		t.Fatalf("expected valid UUID, got %q", a.ID)
	}
	if a.Metadata["source_id"] != "waste-2" {
		t.Errorf("expected source_id preserved, got %q", a.Metadata["source_id"])
	}
	b := Activity{ID: "waste-2"}
	coerceActivityID(&b)
	if a.ID != b.ID {
		t.Errorf("expected deterministic UUID for same input: %q vs %q", a.ID, b.ID)
	}

	// Different inputs -> different UUIDs.
	c := Activity{ID: "csv-METER-1-3"}
	coerceActivityID(&c)
	if c.ID == a.ID {
		t.Error("different source IDs must map to different UUIDs")
	}

	// Already a UUID -> unchanged.
	existing := uuid.NewString()
	d := Activity{ID: existing}
	coerceActivityID(&d)
	if d.ID != existing {
		t.Errorf("valid UUID must be preserved, got %q", d.ID)
	}

	// Empty -> random UUID.
	e := Activity{}
	coerceActivityID(&e)
	if _, err := uuid.Parse(e.ID); err != nil {
		t.Errorf("empty ID must become a UUID, got %q", e.ID)
	}
}
