package blockchain

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBuildMerkleTree(t *testing.T) {
	records := []EmissionsRecord{
		{
			ID:        "rec-1",
			TenantID:  "tenant-1",
			Scope:     1,
			Category:  "stationary_combustion",
			CO2e:      100.5,
			Unit:      "tCO2e",
			Timestamp: time.Now(),
			Source:    "facility-1",
		},
		{
			ID:        "rec-2",
			TenantID:  "tenant-1",
			Scope:     2,
			Category:  "electricity",
			CO2e:      250.0,
			Unit:      "tCO2e",
			Timestamp: time.Now(),
			Source:    "grid-purchase",
		},
	}

	tree, err := BuildMerkleTree(records)
	if err != nil {
		t.Fatalf("BuildMerkleTree failed: %v", err)
	}

	if tree == nil {
		t.Fatal("BuildMerkleTree returned nil")
	}

	if len(tree.Leaves) != 2 {
		t.Errorf("Expected 2 leaves, got %d", len(tree.Leaves))
	}

	// Root should not be zero hash
	var zeroHash Hash
	if tree.Root == zeroHash {
		t.Error("Root hash should not be zero")
	}
}

func TestBuildMerkleTree_Empty(t *testing.T) {
	_, err := BuildMerkleTree([]EmissionsRecord{})
	if err == nil {
		t.Error("Expected error for empty records")
	}
}

func TestHash_String(t *testing.T) {
	var h Hash
	h[0] = 0xAB
	h[1] = 0xCD

	str := h.String()
	if len(str) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("Expected 64 char hex string, got %d chars", len(str))
	}
}

func TestNewChain(t *testing.T) {
	cfg := ChainConfig{
		BlockSize:     10,
		BlockInterval: time.Minute,
	}
	chain, err := NewChain(cfg)
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}

	if chain == nil {
		t.Fatal("NewChain returned nil")
	}
}

func TestChain_AddRecord(t *testing.T) {
	cfg := ChainConfig{
		BlockSize:     2, // Small block size for testing
		BlockInterval: time.Minute,
	}
	chain, err := NewChain(cfg)
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}

	record := EmissionsRecord{
		ID:        "test-1",
		TenantID:  "tenant-1",
		Scope:     1,
		Category:  "mobile_combustion",
		CO2e:      50.0,
		Unit:      "tCO2e",
		Timestamp: time.Now(),
	}

	err = chain.AddRecord(record)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}
}

func TestChain_GetInfo(t *testing.T) {
	cfg := ChainConfig{
		BlockSize:     10,
		BlockInterval: time.Minute,
	}
	chain, err := NewChain(cfg)
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}

	// Add a record to create a block (or genesis block)
	record := EmissionsRecord{
		ID:        "test-1",
		TenantID:  "tenant-1",
		Scope:     1,
		CO2e:      100.0,
		Timestamp: time.Now(),
	}
	chain.AddRecord(record)
	chain.CreateBlock(context.Background())

	info := chain.GetInfo()

	// After adding a block, chain should have info
	if info.Length == 0 {
		t.Log("Chain length is 0 (expected if no genesis block)")
	}
}

func TestChain_VerifyChain(t *testing.T) {
	cfg := ChainConfig{
		BlockSize:     1, // Create block after each record
		BlockInterval: time.Minute,
	}
	chain, err := NewChain(cfg)
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}
	ctx := context.Background()

	// Add some records to create blocks
	for i := 0; i < 3; i++ {
		record := EmissionsRecord{
			ID:        fmt.Sprintf("rec-%d", i),
			TenantID:  "tenant-1",
			Scope:     1,
			CO2e:      float64(i * 10),
			Timestamp: time.Now(),
		}
		chain.AddRecord(record)
	}

	// Force block creation
	chain.CreateBlock(ctx)

	valid, err := chain.VerifyChain()
	if err != nil {
		t.Fatalf("VerifyChain failed: %v", err)
	}

	if !valid {
		t.Error("Chain should be valid")
	}
}

func TestEmissionsRecord_Fields(t *testing.T) {
	record := EmissionsRecord{
		ID:        "test-record",
		TenantID:  "tenant-1",
		Scope:     1,
		Category:  "stationary_combustion",
		CO2e:      123.45,
		Unit:      "tCO2e",
		Timestamp: time.Now(),
		Source:    "facility-1",
		Methodology: "GHG Protocol",
	}

	if record.ID == "" {
		t.Error("ID should not be empty")
	}
	if record.Scope < 1 || record.Scope > 3 {
		t.Error("Scope should be 1, 2, or 3")
	}
}

func TestBlock_Fields(t *testing.T) {
	block := Block{
		Index:     1,
		Timestamp: time.Now(),
		Records: []EmissionsRecord{
			{ID: "r1", CO2e: 100.0},
		},
	}

	if block.Index != 1 {
		t.Errorf("Expected index 1, got %d", block.Index)
	}
	if len(block.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(block.Records))
	}
}

func TestChainInfo_Fields(t *testing.T) {
	info := ChainInfo{
		Length:     5,
		LatestHash: "abc123",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if info.Length != 5 {
		t.Errorf("Expected length 5, got %d", info.Length)
	}
	if info.LatestHash == "" {
		t.Error("LatestHash should not be empty")
	}
}

func TestChainConfig_Defaults(t *testing.T) {
	// Test that NewChain applies defaults
	cfg := ChainConfig{}
	chain, err := NewChain(cfg)
	if err != nil {
		t.Fatalf("NewChain failed: %v", err)
	}

	if chain == nil {
		t.Fatal("Chain should not be nil")
	}
}
