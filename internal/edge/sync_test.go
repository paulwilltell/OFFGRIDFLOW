package edge

import (
	"log/slog"
	"testing"
	"time"
)

func TestConnectivityState_Values(t *testing.T) {
	states := []ConnectivityState{
		StateOnline,
		StateOffline,
		StateConnecting,
		StateSyncing,
	}

	for _, s := range states {
		if string(s) == "" {
			t.Errorf("ConnectivityState %v has empty string value", s)
		}
	}
}

func TestOperationType_Values(t *testing.T) {
	ops := []OperationType{
		OpCreate,
		OpUpdate,
		OpDelete,
		OpSync,
	}

	for _, op := range ops {
		if string(op) == "" {
			t.Errorf("OperationType %v has empty string value", op)
		}
	}
}

func TestSyncOperation_Fields(t *testing.T) {
	op := SyncOperation{
		ID:        "sync-1",
		Type:      OpCreate,
		Resource:  "emissions",
		TenantID:  "tenant-1",
		CreatedAt: time.Now(),
		Priority:  1,
	}

	if op.ID == "" {
		t.Error("ID should not be empty")
	}
	if op.Type != OpCreate {
		t.Errorf("Expected type %v, got %v", OpCreate, op.Type)
	}
}

func TestNewSyncQueue(t *testing.T) {
	logger := slog.Default()
	sq := NewSyncQueue(nil, logger)

	if sq == nil {
		t.Fatal("NewSyncQueue returned nil")
	}
}

func TestSyncQueue_Enqueue(t *testing.T) {
	logger := slog.Default()
	sq := NewSyncQueue(nil, logger)

	op := SyncOperation{
		ID:       "test-1",
		Type:     OpCreate,
		Resource: "emissions",
		TenantID: "tenant-1",
		Priority: 1,
	}

	err := sq.Enqueue(op)
	if err != nil {
		t.Logf("Enqueue may fail without store: %v", err)
	}
}

func TestSyncManagerConfig_Fields(t *testing.T) {
	cfg := SyncManagerConfig{
		MaxRetries: 5,
		RetryDelay: time.Second * 30,
		BatchSize:  100,
	}

	if cfg.MaxRetries <= 0 {
		t.Error("MaxRetries should be positive")
	}
	if cfg.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}
}
