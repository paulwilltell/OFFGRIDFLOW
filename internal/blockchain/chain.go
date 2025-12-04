// Package blockchain provides immutable audit trail for emissions records.
//
// This package creates cryptographic hashes and Merkle trees for emissions data
// to ensure tamper-evident audit trails for regulatory compliance.
package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// Hash represents a cryptographic hash.
type Hash [32]byte

// String returns hex representation of hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// EmissionsRecord represents a verified emissions record.
type EmissionsRecord struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	Scope       int       `json:"scope"`
	Category    string    `json:"category"`
	CO2e        float64   `json:"co2e"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
	Methodology string    `json:"methodology"`
	DataHash    string    `json:"dataHash,omitempty"`
}

// Block represents a block in the audit chain.
type Block struct {
	Index      uint64            `json:"index"`
	Timestamp  time.Time         `json:"timestamp"`
	PrevHash   Hash              `json:"prevHash"`
	MerkleRoot Hash              `json:"merkleRoot"`
	Records    []EmissionsRecord `json:"records"`
	Hash       Hash              `json:"hash"`
	Nonce      uint64            `json:"nonce,omitempty"`
}

// ChainInfo provides information about the audit chain.
type ChainInfo struct {
	Length     uint64    `json:"length"`
	LatestHash string    `json:"latestHash"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// =============================================================================
// Merkle Tree
// =============================================================================

// MerkleTree builds a Merkle tree from records.
type MerkleTree struct {
	Root   Hash
	Leaves []Hash
}

// BuildMerkleTree creates a Merkle tree from emissions records.
func BuildMerkleTree(records []EmissionsRecord) (*MerkleTree, error) {
	if len(records) == 0 {
		return nil, errors.New("no records to build tree")
	}

	// Create leaf hashes
	leaves := make([]Hash, len(records))
	for i, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal record: %w", err)
		}
		leaves[i] = sha256.Sum256(data)
	}

	// Build tree bottom-up
	root := buildTreeLevel(leaves)

	return &MerkleTree{
		Root:   root,
		Leaves: leaves,
	}, nil
}

// buildTreeLevel recursively builds tree levels.
func buildTreeLevel(hashes []Hash) Hash {
	if len(hashes) == 1 {
		return hashes[0]
	}

	// Ensure even number of nodes
	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	// Build next level
	nextLevel := make([]Hash, len(hashes)/2)
	for i := 0; i < len(hashes); i += 2 {
		combined := append(hashes[i][:], hashes[i+1][:]...)
		nextLevel[i/2] = sha256.Sum256(combined)
	}

	return buildTreeLevel(nextLevel)
}

// GenerateProof creates a Merkle proof for a record at index.
func (mt *MerkleTree) GenerateProof(index int) ([]Hash, []bool, error) {
	if index < 0 || index >= len(mt.Leaves) {
		return nil, nil, errors.New("index out of range")
	}

	var proof []Hash
	var directions []bool // true = right, false = left

	leaves := mt.Leaves
	idx := index

	for len(leaves) > 1 {
		// Ensure even number
		if len(leaves)%2 != 0 {
			leaves = append(leaves, leaves[len(leaves)-1])
		}

		// Add sibling to proof
		if idx%2 == 0 {
			if idx+1 < len(leaves) {
				proof = append(proof, leaves[idx+1])
				directions = append(directions, true)
			}
		} else {
			proof = append(proof, leaves[idx-1])
			directions = append(directions, false)
		}

		// Build next level
		nextLevel := make([]Hash, len(leaves)/2)
		for i := 0; i < len(leaves); i += 2 {
			combined := append(leaves[i][:], leaves[i+1][:]...)
			nextLevel[i/2] = sha256.Sum256(combined)
		}
		leaves = nextLevel
		idx = idx / 2
	}

	return proof, directions, nil
}

// VerifyProof verifies a Merkle proof.
func VerifyProof(leaf Hash, proof []Hash, directions []bool, root Hash) bool {
	current := leaf

	for i, sibling := range proof {
		var combined []byte
		if directions[i] {
			combined = append(current[:], sibling[:]...)
		} else {
			combined = append(sibling[:], current[:]...)
		}
		current = sha256.Sum256(combined)
	}

	return current == root
}

// =============================================================================
// Audit Chain
// =============================================================================

// Chain manages the immutable audit blockchain.
type Chain struct {
	blocks  []Block
	pending []EmissionsRecord
	store   ChainStore
	logger  *slog.Logger
	mu      sync.RWMutex

	// Configuration
	blockSize int           // Records per block
	interval  time.Duration // Block creation interval
}

// ChainStore persists the blockchain.
type ChainStore interface {
	SaveBlock(ctx context.Context, block Block) error
	GetBlock(ctx context.Context, index uint64) (*Block, error)
	GetLatestBlock(ctx context.Context) (*Block, error)
	GetBlockCount(ctx context.Context) (uint64, error)
	GetBlocks(ctx context.Context, startIndex, count uint64) ([]Block, error)
}

// ChainConfig configures the audit chain.
type ChainConfig struct {
	Store         ChainStore
	BlockSize     int
	BlockInterval time.Duration
	Logger        *slog.Logger
}

// NewChain creates a new audit blockchain.
func NewChain(cfg ChainConfig) (*Chain, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.BlockSize <= 0 {
		cfg.BlockSize = 100
	}
	if cfg.BlockInterval <= 0 {
		cfg.BlockInterval = 5 * time.Minute
	}

	c := &Chain{
		blocks:    make([]Block, 0),
		pending:   make([]EmissionsRecord, 0),
		store:     cfg.Store,
		logger:    cfg.Logger.With("component", "blockchain"),
		blockSize: cfg.BlockSize,
		interval:  cfg.BlockInterval,
	}

	// Load existing blocks
	if cfg.Store != nil {
		if err := c.loadChain(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to load chain: %w", err)
		}
	}

	return c, nil
}

// loadChain loads existing blocks from storage.
func (c *Chain) loadChain(ctx context.Context) error {
	if c.store == nil {
		return nil
	}

	count, err := c.store.GetBlockCount(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create genesis block
		genesis := c.createGenesisBlock()
		if err := c.store.SaveBlock(ctx, genesis); err != nil {
			return err
		}
		c.blocks = append(c.blocks, genesis)
		return nil
	}

	// Load all blocks
	blocks, err := c.store.GetBlocks(ctx, 0, count)
	if err != nil {
		return err
	}

	c.blocks = blocks
	return nil
}

// createGenesisBlock creates the first block.
func (c *Chain) createGenesisBlock() Block {
	return Block{
		Index:      0,
		Timestamp:  time.Now().UTC(),
		PrevHash:   Hash{},
		MerkleRoot: Hash{},
		Records:    nil,
		Hash:       sha256.Sum256([]byte("OffGridFlow Genesis Block")),
	}
}

// AddRecord adds an emissions record to pending.
func (c *Chain) AddRecord(record EmissionsRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Hash the record data
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(data)
	record.DataHash = hex.EncodeToString(hash[:])

	c.pending = append(c.pending, record)

	// Auto-create block if size reached
	if len(c.pending) >= c.blockSize {
		return c.createBlockLocked(context.Background())
	}

	return nil
}

// CreateBlock creates a new block from pending records.
func (c *Chain) CreateBlock(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.createBlockLocked(ctx)
}

func (c *Chain) createBlockLocked(ctx context.Context) error {
	if len(c.pending) == 0 {
		return nil
	}

	// Build Merkle tree
	tree, err := BuildMerkleTree(c.pending)
	if err != nil {
		return err
	}

	// Get previous hash
	var prevHash Hash
	if len(c.blocks) > 0 {
		prevHash = c.blocks[len(c.blocks)-1].Hash
	}

	// Create block
	block := Block{
		Index:      uint64(len(c.blocks)),
		Timestamp:  time.Now().UTC(),
		PrevHash:   prevHash,
		MerkleRoot: tree.Root,
		Records:    c.pending,
	}

	// Calculate block hash
	blockData, _ := json.Marshal(block)
	block.Hash = sha256.Sum256(blockData)

	// Store block
	if c.store != nil {
		if err := c.store.SaveBlock(ctx, block); err != nil {
			return err
		}
	}

	c.blocks = append(c.blocks, block)
	c.pending = nil

	c.logger.Info("created block",
		"index", block.Index,
		"records", len(block.Records),
		"hash", block.Hash.String()[:16])

	return nil
}

// VerifyChain verifies the integrity of the entire chain.
func (c *Chain) VerifyChain() (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := 1; i < len(c.blocks); i++ {
		block := c.blocks[i]
		prevBlock := c.blocks[i-1]

		// Verify prev hash
		if block.PrevHash != prevBlock.Hash {
			return false, fmt.Errorf("invalid prev hash at block %d", i)
		}

		// Verify Merkle root
		if len(block.Records) > 0 {
			tree, err := BuildMerkleTree(block.Records)
			if err != nil {
				return false, err
			}
			if tree.Root != block.MerkleRoot {
				return false, fmt.Errorf("invalid Merkle root at block %d", i)
			}
		}
	}

	return true, nil
}

// GetRecord finds a record by ID and returns its verification info.
func (c *Chain) GetRecord(recordID string) (*EmissionsRecord, *RecordProof, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, block := range c.blocks {
		for i, record := range block.Records {
			if record.ID == recordID {
				// Generate proof
				tree, err := BuildMerkleTree(block.Records)
				if err != nil {
					return nil, nil, err
				}

				proof, directions, err := tree.GenerateProof(i)
				if err != nil {
					return nil, nil, err
				}

				return &record, &RecordProof{
					BlockIndex:  block.Index,
					RecordIndex: i,
					MerkleProof: proof,
					Directions:  directions,
					MerkleRoot:  block.MerkleRoot,
					BlockHash:   block.Hash,
				}, nil
			}
		}
	}

	return nil, nil, errors.New("record not found")
}

// RecordProof contains proof of a record's inclusion.
type RecordProof struct {
	BlockIndex  uint64 `json:"blockIndex"`
	RecordIndex int    `json:"recordIndex"`
	MerkleProof []Hash `json:"merkleProof"`
	Directions  []bool `json:"directions"`
	MerkleRoot  Hash   `json:"merkleRoot"`
	BlockHash   Hash   `json:"blockHash"`
}

// Verify verifies a record proof.
func (rp *RecordProof) Verify(record EmissionsRecord) (bool, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return false, err
	}

	leaf := sha256.Sum256(data)
	return VerifyProof(leaf, rp.MerkleProof, rp.Directions, rp.MerkleRoot), nil
}

// GetInfo returns chain information.
func (c *Chain) GetInfo() ChainInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var latestHash string
	var updatedAt time.Time

	if len(c.blocks) > 0 {
		latest := c.blocks[len(c.blocks)-1]
		latestHash = latest.Hash.String()
		updatedAt = latest.Timestamp
	}

	var createdAt time.Time
	if len(c.blocks) > 0 {
		createdAt = c.blocks[0].Timestamp
	}

	return ChainInfo{
		Length:     uint64(len(c.blocks)),
		LatestHash: latestHash,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// Start begins periodic block creation.
func (c *Chain) Start(ctx context.Context) {
	c.logger.Info("starting blockchain service",
		"interval", c.interval)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush pending records
			c.CreateBlock(context.Background())
			return
		case <-ticker.C:
			if err := c.CreateBlock(ctx); err != nil {
				c.logger.Error("failed to create block", "error", err)
			}
		}
	}
}

// =============================================================================
// External Chain Integration
// =============================================================================

// AnchorService anchors blockchain hashes to external chains.
type AnchorService struct {
	chain    *Chain
	provider AnchorProvider
	logger   *slog.Logger
	interval time.Duration
}

// AnchorProvider publishes hashes to external blockchain.
type AnchorProvider interface {
	// Anchor publishes a hash and returns transaction ID
	Anchor(ctx context.Context, hash Hash, metadata map[string]string) (string, error)

	// Verify checks if a hash is anchored
	Verify(ctx context.Context, hash Hash, txID string) (bool, error)
}

// Anchor represents an external blockchain anchor.
type Anchor struct {
	BlockIndex uint64    `json:"blockIndex"`
	Hash       Hash      `json:"hash"`
	Chain      string    `json:"chain"` // "ethereum", "bitcoin", etc.
	TxID       string    `json:"txId"`
	AnchoredAt time.Time `json:"anchoredAt"`
}

// NewAnchorService creates a new anchor service.
func NewAnchorService(chain *Chain, provider AnchorProvider, logger *slog.Logger) *AnchorService {
	return &AnchorService{
		chain:    chain,
		provider: provider,
		logger:   logger.With("component", "anchor-service"),
		interval: 1 * time.Hour,
	}
}

// AnchorLatest anchors the latest block to external chain.
func (as *AnchorService) AnchorLatest(ctx context.Context) (*Anchor, error) {
	as.chain.mu.RLock()
	if len(as.chain.blocks) == 0 {
		as.chain.mu.RUnlock()
		return nil, errors.New("no blocks to anchor")
	}
	latest := as.chain.blocks[len(as.chain.blocks)-1]
	as.chain.mu.RUnlock()

	metadata := map[string]string{
		"blockIndex": fmt.Sprintf("%d", latest.Index),
		"timestamp":  latest.Timestamp.Format(time.RFC3339),
		"system":     "OffGridFlow",
	}

	txID, err := as.provider.Anchor(ctx, latest.Hash, metadata)
	if err != nil {
		return nil, err
	}

	return &Anchor{
		BlockIndex: latest.Index,
		Hash:       latest.Hash,
		TxID:       txID,
		AnchoredAt: time.Now(),
	}, nil
}

// =============================================================================
// Certificate Generation
// =============================================================================

// Certificate is a verifiable emissions certificate.
type Certificate struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Period       string    `json:"period"`
	TotalCO2e    float64   `json:"totalCo2e"`
	Scope1       float64   `json:"scope1"`
	Scope2       float64   `json:"scope2"`
	Scope3       float64   `json:"scope3"`
	Methodology  string    `json:"methodology"`
	RecordHashes []string  `json:"recordHashes"`
	MerkleRoot   string    `json:"merkleRoot"`
	ChainAnchor  *Anchor   `json:"chainAnchor,omitempty"`
	GeneratedAt  time.Time `json:"generatedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Signature    string    `json:"signature,omitempty"`
}

// GenerateCertificate creates a verifiable emissions certificate.
func (c *Chain) GenerateCertificate(tenantID, period string) (*Certificate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var records []EmissionsRecord
	var hashes []string

	// Collect all records for tenant
	for _, block := range c.blocks {
		for _, record := range block.Records {
			if record.TenantID == tenantID {
				records = append(records, record)
				hashes = append(hashes, record.DataHash)
			}
		}
	}

	if len(records) == 0 {
		return nil, errors.New("no records found for tenant")
	}

	// Calculate totals
	var scope1, scope2, scope3 float64
	for _, r := range records {
		switch r.Scope {
		case 1:
			scope1 += r.CO2e
		case 2:
			scope2 += r.CO2e
		case 3:
			scope3 += r.CO2e
		}
	}

	// Build Merkle tree
	tree, err := BuildMerkleTree(records)
	if err != nil {
		return nil, err
	}

	cert := &Certificate{
		ID:           fmt.Sprintf("CERT-%s-%s", tenantID[:8], period),
		TenantID:     tenantID,
		Period:       period,
		TotalCO2e:    scope1 + scope2 + scope3,
		Scope1:       scope1,
		Scope2:       scope2,
		Scope3:       scope3,
		Methodology:  "GHG Protocol Corporate Standard",
		RecordHashes: hashes,
		MerkleRoot:   tree.Root.String(),
		GeneratedAt:  time.Now(),
		ExpiresAt:    time.Now().AddDate(1, 0, 0),
	}

	return cert, nil
}
