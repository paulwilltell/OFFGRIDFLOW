// Package blockchain - Polkadot/Substrate integration for carbon credit tokenization
package blockchain

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
)

// =============================================================================
// Polkadot Provider for External Chain Anchoring
// =============================================================================

// PolkadotProvider implements AnchorProvider for Polkadot/Substrate chains.
type PolkadotProvider struct {
	api        *gsrpc.SubstrateAPI
	keyring    signature.KeyringPair
	metadata   *types.Metadata
	runtimeVer *types.RuntimeVersion
	genesisHash types.Hash
	logger     *slog.Logger
	network    string // "polkadot", "kusama", "custom"
}

// PolkadotConfig configures Polkadot connection.
type PolkadotConfig struct {
	RPCURL  string // e.g., "wss://rpc.polkadot.io"
	Seed    string // Seed phrase or private key
	Network string // "polkadot", "kusama", "westend"
	Logger  *slog.Logger
}

// NewPolkadotProvider creates a Polkadot blockchain provider.
func NewPolkadotProvider(cfg PolkadotConfig) (*PolkadotProvider, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Connect to Substrate node
	api, err := gsrpc.NewSubstrateAPI(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Polkadot: %w", err)
	}

	// Get metadata and runtime version
	metadata, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	runtimeVer, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime version: %w", err)
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get genesis hash: %w", err)
	}

	// Create keyring from seed
	keyring, err := signature.KeyringPairFromSecret(cfg.Seed, uint16(cfg.NetworkID()))
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}

	return &PolkadotProvider{
		api:        api,
		keyring:    keyring,
		metadata:   metadata,
		runtimeVer: runtimeVer,
		genesisHash: genesisHash,
		logger:     cfg.Logger.With("component", "polkadot-provider"),
		network:    cfg.Network,
	}, nil
}

// NetworkID returns the network ID for keyring.
func (cfg PolkadotConfig) NetworkID() int {
	switch cfg.Network {
	case "polkadot":
		return 0
	case "kusama":
		return 2
	case "westend":
		return 42
	default:
		return 42 // default to testnet
	}
}

// Anchor publishes a hash to Polkadot using remarks extrinsic.
func (pp *PolkadotProvider) Anchor(ctx context.Context, hash Hash, metadata map[string]string) (string, error) {
	// Build metadata string
	metadataStr := fmt.Sprintf("OffGridFlow|%s", hash.String())
	for k, v := range metadata {
		metadataStr += fmt.Sprintf("|%s=%s", k, v)
	}

	// Create system.remark extrinsic
	call, err := types.NewCall(pp.metadata, "System.remark", []byte(metadataStr))
	if err != nil {
		return "", fmt.Errorf("failed to create call: %w", err)
	}

	// Create extrinsic
	ext := types.NewExtrinsic(call)

	// Get account info for nonce
	key, err := types.CreateStorageKey(pp.metadata, "System", "Account", pp.keyring.PublicKey)
	if err != nil {
		return "", err
	}

	var accountInfo types.AccountInfo
	ok, err := pp.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return "", errors.New("failed to get account info")
	}

	// Sign extrinsic
	signOptions := types.SignatureOptions{
		BlockHash:          pp.genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        pp.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        pp.runtimeVer.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: pp.runtimeVer.TransactionVersion,
	}

	if err := ext.Sign(pp.keyring, signOptions); err != nil {
		return "", fmt.Errorf("failed to sign extrinsic: %w", err)
	}

	// Submit extrinsic
	sub, err := pp.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return "", fmt.Errorf("failed to submit extrinsic: %w", err)
	}
	defer sub.Unsubscribe()

	// Wait for finalization
	timeout := time.After(60 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", errors.New("transaction timeout")
		case status := <-sub.Chan():
			if status.IsInBlock {
				pp.logger.Info("anchored hash to Polkadot",
					"hash", hash.String()[:16],
					"block", status.AsInBlock.Hex())
				return status.AsInBlock.Hex(), nil
			}
			if status.IsFinalized {
				return status.AsFinalized.Hex(), nil
			}
		case err := <-sub.Err():
			return "", fmt.Errorf("subscription error: %w", err)
		}
	}
}

// Verify checks if a hash is anchored on Polkadot.
func (pp *PolkadotProvider) Verify(ctx context.Context, hash Hash, blockHash string) (bool, error) {
	// Get block by hash
	h, err := types.NewHashFromHexString(blockHash)
	if err != nil {
		return false, err
	}

	block, err := pp.api.RPC.Chain.GetBlock(h)
	if err != nil {
		return false, err
	}

	// Search for remark extrinsic with our hash
	hashStr := hash.String()
	for _, ext := range block.Block.Extrinsics {
		// Decode extrinsic
		decoded := ext.Method
		if decoded.CallIndex.SectionIndex == 0 && decoded.CallIndex.MethodIndex == 1 {
			// System.remark
			if len(decoded.Args) > 0 {
				remark := string(decoded.Args)
				if contains(remark, hashStr) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// =============================================================================
// Carbon Credit Pallet Integration
// =============================================================================

// PolkadotCarbonCredit represents a carbon credit on Polkadot.
type PolkadotCarbonCredit struct {
	ID        types.U128
	Owner     types.AccountID
	CO2e      types.U128 // in micro-tons
	Vintage   types.U32
	ProjectID string
	Retired   bool
	MintedAt  types.U64
}

// PolkadotCreditManager manages carbon credits on Polkadot.
type PolkadotCreditManager struct {
	provider *PolkadotProvider
	logger   *slog.Logger
}

// NewPolkadotCreditManager creates a credit manager for Polkadot.
func NewPolkadotCreditManager(provider *PolkadotProvider) *PolkadotCreditManager {
	return &PolkadotCreditManager{
		provider: provider,
		logger:   provider.logger.With("component", "polkadot-credit-manager"),
	}
}

// MintCredit mints a new carbon credit on Polkadot.
func (pcm *PolkadotCreditManager) MintCredit(ctx context.Context, co2e float64, vintage int, projectID string) (string, error) {
	// Convert CO2e to micro-tons
	co2eMicro := uint128FromFloat(co2e * 1e6)
	vintageU32 := types.NewU32(uint32(vintage))

	// Create mint extrinsic
	// CarbonCredits.mint(co2e, vintage, project_id)
	call, err := types.NewCall(
		pcm.provider.metadata,
		"CarbonCredits.mint",
		co2eMicro,
		vintageU32,
		[]byte(projectID),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create mint call: %w", err)
	}

	// Submit and wait
	return pcm.submitExtrinsic(ctx, call)
}

// RetireCredit retires a carbon credit.
func (pcm *PolkadotCreditManager) RetireCredit(ctx context.Context, creditID string) error {
	// Parse credit ID
	id, err := parseU128FromString(creditID)
	if err != nil {
		return err
	}

	// Create retire extrinsic
	call, err := types.NewCall(
		pcm.provider.metadata,
		"CarbonCredits.retire",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to create retire call: %w", err)
	}

	_, err = pcm.submitExtrinsic(ctx, call)
	return err
}

// TransferCredit transfers a credit to another account.
func (pcm *PolkadotCreditManager) TransferCredit(ctx context.Context, creditID string, toAddress string) error {
	// Parse credit ID
	id, err := parseU128FromString(creditID)
	if err != nil {
		return err
	}

	// Parse destination address
	dest, err := parseAccountIDHex(toAddress)
	if err != nil {
		return err
	}

	// Create transfer extrinsic
	call, err := types.NewCall(
		pcm.provider.metadata,
		"CarbonCredits.transfer",
		id,
		dest,
	)
	if err != nil {
		return fmt.Errorf("failed to create transfer call: %w", err)
	}

	_, err = pcm.submitExtrinsic(ctx, call)
	return err
}

// GetCredit retrieves credit information.
func (pcm *PolkadotCreditManager) GetCredit(ctx context.Context, creditID string) (*PolkadotCarbonCredit, error) {
	// Parse credit ID
	id, err := parseU128FromString(creditID)
	if err != nil {
		return nil, err
	}

	// Create storage key
	idBytes, err := encodeU128(id)
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(pcm.provider.metadata, "CarbonCredits", "Credits", idBytes)
	if err != nil {
		return nil, err
	}

	// Query storage
	var credit PolkadotCarbonCredit
	ok, err := pcm.provider.api.RPC.State.GetStorageLatest(key, &credit)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("credit not found")
	}

	return &credit, nil
}

// GetOwnerCredits retrieves all credits owned by an address.
func (pcm *PolkadotCreditManager) GetOwnerCredits(ctx context.Context, ownerAddress string) ([]string, error) {
	// Parse owner address
	owner, err := parseAccountIDHex(ownerAddress)
	if err != nil {
		return nil, err
	}

	// Create storage key
	key, err := types.CreateStorageKey(pcm.provider.metadata, "CarbonCredits", "CreditsByOwner", owner.ToBytes())
	if err != nil {
		return nil, err
	}

	// Query storage
	var creditIDs []types.U128
	ok, err := pcm.provider.api.RPC.State.GetStorageLatest(key, &creditIDs)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []string{}, nil
	}

	// Convert to strings
	result := make([]string, len(creditIDs))
	for i, id := range creditIDs {
		result[i] = id.String()
	}

	return result, nil
}

// submitExtrinsic submits and waits for extrinsic finalization.
func (pcm *PolkadotCreditManager) submitExtrinsic(ctx context.Context, call types.Call) (string, error) {
	// Create extrinsic
	ext := types.NewExtrinsic(call)

	// Get account info for nonce
	key, err := types.CreateStorageKey(pcm.provider.metadata, "System", "Account", pcm.provider.keyring.PublicKey)
	if err != nil {
		return "", err
	}

	var accountInfo types.AccountInfo
	ok, err := pcm.provider.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return "", errors.New("failed to get account info")
	}

	// Sign extrinsic
	signOptions := types.SignatureOptions{
		BlockHash:          pcm.provider.genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        pcm.provider.genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
		SpecVersion:        pcm.provider.runtimeVer.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: pcm.provider.runtimeVer.TransactionVersion,
	}

	if err := ext.Sign(pcm.provider.keyring, signOptions); err != nil {
		return "", fmt.Errorf("failed to sign extrinsic: %w", err)
	}

	// Submit extrinsic
	sub, err := pcm.provider.api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return "", fmt.Errorf("failed to submit extrinsic: %w", err)
	}
	defer sub.Unsubscribe()

	// Wait for finalization
	timeout := time.After(60 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timeout:
			return "", errors.New("transaction timeout")
		case status := <-sub.Chan():
			if status.IsFinalized {
				pcm.logger.Info("extrinsic finalized", "block", status.AsFinalized.Hex())
				return status.AsFinalized.Hex(), nil
			}
		case err := <-sub.Err():
			return "", fmt.Errorf("subscription error: %w", err)
		}
	}
}

// =============================================================================
// Polkadot Wallet Management
// =============================================================================

// PolkadotWallet represents a Polkadot wallet.
type PolkadotWallet struct {
	Address string
	Keyring signature.KeyringPair
	Balance types.U128
}

// PolkadotWalletManager manages Polkadot wallets.
type PolkadotWalletManager struct {
	provider *PolkadotProvider
	logger   *slog.Logger
}

// NewPolkadotWalletManager creates a wallet manager.
func NewPolkadotWalletManager(provider *PolkadotProvider) *PolkadotWalletManager {
	return &PolkadotWalletManager{
		provider: provider,
		logger:   provider.logger.With("component", "polkadot-wallet-manager"),
	}
}

// CreateWallet generates a new Polkadot wallet.
func (pwm *PolkadotWalletManager) CreateWallet(seed string) (*PolkadotWallet, error) {
	keyring, err := signature.KeyringPairFromSecret(seed, uint16(0))
	if err != nil {
		return nil, err
	}

	address := keyring.Address

	pwm.logger.Info("created new Polkadot wallet", "address", address)

	return &PolkadotWallet{
		Address: address,
		Keyring: keyring,
		Balance: types.NewU128(*big.NewInt(0)),
	}, nil
}

// GetBalance retrieves DOT balance.
func (pwm *PolkadotWalletManager) GetBalance(ctx context.Context, address string) (types.U128, error) {
	// Parse address
	accountID, err := parseAccountIDHex(address)
	if err != nil {
		return types.U128{}, err
	}

	// Create storage key
	key, err := types.CreateStorageKey(pwm.provider.metadata, "System", "Account", accountID.ToBytes())
	if err != nil {
		return types.U128{}, err
	}

	// Query storage
	var accountInfo types.AccountInfo
	ok, err := pwm.provider.api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return types.U128{}, err
	}
	if !ok {
		return types.NewU128(*big.NewInt(0)), nil
	}

	return types.NewU128(*accountInfo.Data.Free.Int), nil
}

// ExportSeed exports wallet seed (use with caution).
func (pw *PolkadotWallet) ExportSeed() string {
	return hex.EncodeToString(pw.Keyring.PublicKey)
}

func parseU128FromString(value string) (types.U128, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return types.U128{}, fmt.Errorf("empty U128 value")
	}

	bi := new(big.Int)
	if _, ok := bi.SetString(trimmed, 10); !ok {
		return types.U128{}, fmt.Errorf("invalid U128 value: %q", value)
	}

	return types.NewU128(*bi), nil
}

func encodeU128(value types.U128) ([]byte, error) {
	if value.Int == nil {
		value.Int = big.NewInt(0)
	}

	return types.BigIntToUintBytes(value.Int, 16)
}

func parseAccountIDHex(address string) (*types.AccountID, error) {
	cleaned := strings.TrimSpace(address)
	if cleaned == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if !strings.HasPrefix(cleaned, "0x") {
		cleaned = "0x" + cleaned
	}

	return types.NewAccountIDFromHexString(cleaned)
}

// =============================================================================
// Utility Functions
// =============================================================================

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && substr != "" && s[0:len(substr)] == substr
}

// uint128FromFloat converts a float to a U128 representation.
func uint128FromFloat(f float64) types.U128 {
	return types.NewU128(*big.NewInt(int64(f)))
}
