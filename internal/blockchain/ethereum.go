// Package blockchain - Ethereum integration for carbon credit tokenization
package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// =============================================================================
// Ethereum Provider for External Chain Anchoring
// =============================================================================

// EthereumProvider implements AnchorProvider for Ethereum.
type EthereumProvider struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	logger     *slog.Logger

	// Smart contract addresses
	anchorContract common.Address
	tokenContract  common.Address
}

// EthereumConfig configures Ethereum connection.
type EthereumConfig struct {
	RPCURL         string // e.g., "https://mainnet.infura.io/v3/YOUR-PROJECT-ID"
	PrivateKey     string // Hex-encoded private key
	ChainID        int64  // 1=mainnet, 5=goerli, 11155111=sepolia
	AnchorContract string // Address of anchor contract
	TokenContract  string // Address of carbon credit token contract
	Logger         *slog.Logger
}

// NewEthereumProvider creates an Ethereum blockchain provider.
func NewEthereumProvider(cfg EthereumConfig) (*EthereumProvider, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Connect to Ethereum node
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(cfg.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return &EthereumProvider{
		client:         client,
		privateKey:     privateKey,
		chainID:        big.NewInt(cfg.ChainID),
		anchorContract: common.HexToAddress(cfg.AnchorContract),
		tokenContract:  common.HexToAddress(cfg.TokenContract),
		logger:         cfg.Logger.With("component", "ethereum-provider"),
	}, nil
}

// Anchor publishes a hash to Ethereum smart contract.
func (ep *EthereumProvider) Anchor(ctx context.Context, hash Hash, metadata map[string]string) (string, error) {
	// Create transaction auth
	auth, err := ep.createTransactor(ctx)
	if err != nil {
		return "", err
	}

	// Prepare contract call data
	// anchor(bytes32 hash, string metadata)
	contractABI, err := abi.JSON(strings.NewReader(AnchorContractABI))
	if err != nil {
		return "", err
	}

	// Convert metadata to JSON string
	metadataJSON := ""
	for k, v := range metadata {
		metadataJSON += fmt.Sprintf("%s=%s;", k, v)
	}

	data, err := contractABI.Pack("anchor", hash, metadataJSON)
	if err != nil {
		return "", err
	}

	// Estimate gas
	gasLimit, err := ep.client.EstimateGas(ctx, ethereum.CallMsg{
		From: auth.From,
		To:   &ep.anchorContract,
		Data: data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Send transaction
	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		ep.anchorContract,
		big.NewInt(0),
		gasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ep.chainID), ep.privateKey)
	if err != nil {
		return "", err
	}

	if err := ep.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	ep.logger.Info("anchored hash to Ethereum",
		"hash", hash.String()[:16],
		"txHash", signedTx.Hash().Hex())

	return signedTx.Hash().Hex(), nil
}

// Verify checks if a hash is anchored on Ethereum.
func (ep *EthereumProvider) Verify(ctx context.Context, hash Hash, txID string) (bool, error) {
	// Get transaction receipt
	receipt, err := ep.client.TransactionReceipt(ctx, common.HexToHash(txID))
	if err != nil {
		return false, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return false, errors.New("transaction failed")
	}

	// Parse logs to verify hash
	contractABI, err := abi.JSON(strings.NewReader(AnchorContractABI))
	if err != nil {
		return false, err
	}

	for _, log := range receipt.Logs {
		if log.Address != ep.anchorContract {
			continue
		}

		event := struct {
			Hash     [32]byte
			Metadata string
		}{}

		if err := contractABI.UnpackIntoInterface(&event, "Anchored", log.Data); err != nil {
			continue
		}

		if Hash(event.Hash) == hash {
			return true, nil
		}
	}

	return false, nil
}

// createTransactor creates a transaction signer.
func (ep *EthereumProvider) createTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := ep.client.PendingNonceAt(ctx, crypto.PubkeyToAddress(ep.privateKey.PublicKey))
	if err != nil {
		return nil, err
	}

	gasPrice, err := ep.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ep.privateKey, ep.chainID)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice
	auth.Context = ctx

	return auth, nil
}

// =============================================================================
// Carbon Credit Token (ERC-721 NFT)
// =============================================================================

// CarbonCreditToken represents a carbon credit as an NFT.
type CarbonCreditToken struct {
	TokenID     *big.Int         `json:"tokenId"`
	Owner       common.Address   `json:"owner"`
	CO2e        float64          `json:"co2e"`
	Vintage     int              `json:"vintage"` // Year
	ProjectID   string           `json:"projectId"`
	Standard    string           `json:"standard"` // "VCS", "GoldStandard", etc.
	Metadata    string           `json:"metadata"`
	MintedAt    time.Time        `json:"mintedAt"`
	RetiredAt   *time.Time       `json:"retiredAt,omitempty"`
	ChainAnchor Hash             `json:"chainAnchor"`
}

// TokenManager manages carbon credit NFTs.
type TokenManager struct {
	provider *EthereumProvider
	contract common.Address
	logger   *slog.Logger
}

// NewTokenManager creates a token manager.
func NewTokenManager(provider *EthereumProvider, contractAddress string) *TokenManager {
	return &TokenManager{
		provider: provider,
		contract: common.HexToAddress(contractAddress),
		logger:   provider.logger.With("component", "token-manager"),
	}
}

// MintCarbonCredit mints a new carbon credit NFT.
func (tm *TokenManager) MintCarbonCredit(ctx context.Context, credit CarbonCreditToken) (*big.Int, error) {
	auth, err := tm.provider.createTransactor(ctx)
	if err != nil {
		return nil, err
	}

	// Prepare mint call
	contractABI, err := abi.JSON(strings.NewReader(CarbonCreditNFTABI))
	if err != nil {
		return nil, err
	}

	// mint(address to, uint256 co2e, uint256 vintage, string projectId, string metadata)
	data, err := contractABI.Pack("mint",
		credit.Owner,
		big.NewInt(int64(credit.CO2e*1e6)), // Convert to micro-tons
		big.NewInt(int64(credit.Vintage)),
		credit.ProjectID,
		credit.Metadata,
	)
	if err != nil {
		return nil, err
	}

	// Send transaction
	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		tm.contract,
		big.NewInt(0),
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(tm.provider.chainID), tm.provider.privateKey)
	if err != nil {
		return nil, err
	}

	if err := tm.provider.client.SendTransaction(ctx, signedTx); err != nil {
		return nil, err
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, tm.provider.client, signedTx)
	if err != nil {
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("mint transaction failed")
	}

	// Extract token ID from logs
	for _, log := range receipt.Logs {
		if log.Address != tm.contract {
			continue
		}

		event := struct {
			TokenId *big.Int
			To      common.Address
		}{}

		if err := contractABI.UnpackIntoInterface(&event, "Transfer", log.Data); err != nil {
			continue
		}

		tm.logger.Info("minted carbon credit NFT",
			"tokenId", event.TokenId,
			"owner", event.To.Hex(),
			"co2e", credit.CO2e)

		return event.TokenId, nil
	}

	return nil, errors.New("token ID not found in receipt")
}

// RetireCredit permanently retires a carbon credit.
func (tm *TokenManager) RetireCredit(ctx context.Context, tokenID *big.Int) error {
	auth, err := tm.provider.createTransactor(ctx)
	if err != nil {
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(CarbonCreditNFTABI))
	if err != nil {
		return err
	}

	// retire(uint256 tokenId)
	data, err := contractABI.Pack("retire", tokenID)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		tm.contract,
		big.NewInt(0),
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(tm.provider.chainID), tm.provider.privateKey)
	if err != nil {
		return err
	}

	if err := tm.provider.client.SendTransaction(ctx, signedTx); err != nil {
		return err
	}

	tm.logger.Info("retired carbon credit", "tokenId", tokenID)
	return nil
}

// GetCredit retrieves credit information.
func (tm *TokenManager) GetCredit(ctx context.Context, tokenID *big.Int) (*CarbonCreditToken, error) {
	contractABI, err := abi.JSON(strings.NewReader(CarbonCreditNFTABI))
	if err != nil {
		return nil, err
	}

	// credits(uint256) returns (uint256 co2e, uint256 vintage, string projectId, bool retired)
	data, err := contractABI.Pack("credits", tokenID)
	if err != nil {
		return nil, err
	}

	result, err := tm.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &tm.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var credit struct {
		CO2e      *big.Int
		Vintage   *big.Int
		ProjectID string
		Retired   bool
	}

	if err := contractABI.UnpackIntoInterface(&credit, "credits", result); err != nil {
		return nil, err
	}

	token := &CarbonCreditToken{
		TokenID:   tokenID,
		CO2e:      float64(credit.CO2e.Int64()) / 1e6,
		Vintage:   int(credit.Vintage.Int64()),
		ProjectID: credit.ProjectID,
	}

	if credit.Retired {
		now := time.Now()
		token.RetiredAt = &now
	}

	return token, nil
}

// TransferCredit transfers ownership of a credit.
func (tm *TokenManager) TransferCredit(ctx context.Context, tokenID *big.Int, to common.Address) error {
	auth, err := tm.provider.createTransactor(ctx)
	if err != nil {
		return err
	}

	contractABI, err := abi.JSON(strings.NewReader(CarbonCreditNFTABI))
	if err != nil {
		return err
	}

	from := crypto.PubkeyToAddress(tm.provider.privateKey.PublicKey)

	// transferFrom(address from, address to, uint256 tokenId)
	data, err := contractABI.Pack("transferFrom", from, to, tokenID)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		tm.contract,
		big.NewInt(0),
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(tm.provider.chainID), tm.provider.privateKey)
	if err != nil {
		return err
	}

	if err := tm.provider.client.SendTransaction(ctx, signedTx); err != nil {
		return err
	}

	tm.logger.Info("transferred carbon credit",
		"tokenId", tokenID,
		"from", from.Hex(),
		"to", to.Hex())

	return nil
}

// =============================================================================
// Wallet Management
// =============================================================================

// Wallet represents an Ethereum wallet for carbon credits.
type Wallet struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	Balance    *big.Int // ETH balance
	Credits    []*big.Int
}

// WalletManager manages Ethereum wallets.
type WalletManager struct {
	provider *EthereumProvider
	logger   *slog.Logger
}

// NewWalletManager creates a wallet manager.
func NewWalletManager(provider *EthereumProvider) *WalletManager {
	return &WalletManager{
		provider: provider,
		logger:   provider.logger.With("component", "wallet-manager"),
	}
}

// CreateWallet generates a new Ethereum wallet.
func (wm *WalletManager) CreateWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	wm.logger.Info("created new wallet", "address", address.Hex())

	return &Wallet{
		Address:    address,
		PrivateKey: privateKey,
		Balance:    big.NewInt(0),
		Credits:    make([]*big.Int, 0),
	}, nil
}

// ImportWallet imports a wallet from private key.
func (wm *WalletManager) ImportWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &Wallet{
		Address:    address,
		PrivateKey: privateKey,
		Balance:    big.NewInt(0),
		Credits:    make([]*big.Int, 0),
	}, nil
}

// GetBalance retrieves ETH balance.
func (wm *WalletManager) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return wm.provider.client.BalanceAt(ctx, address, nil)
}

// GetCredits retrieves all carbon credits owned by address.
func (wm *WalletManager) GetCredits(ctx context.Context, address common.Address, tokenContract common.Address) ([]*big.Int, error) {
	contractABI, err := abi.JSON(strings.NewReader(CarbonCreditNFTABI))
	if err != nil {
		return nil, err
	}

	// balanceOf(address owner)
	data, err := contractABI.Pack("balanceOf", address)
	if err != nil {
		return nil, err
	}

	result, err := wm.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenContract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	if err := contractABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, err
	}

	// Get token IDs
	credits := make([]*big.Int, 0, balance.Int64())
	for i := int64(0); i < balance.Int64(); i++ {
		// tokenOfOwnerByIndex(address owner, uint256 index)
		data, err := contractABI.Pack("tokenOfOwnerByIndex", address, big.NewInt(i))
		if err != nil {
			continue
		}

		result, err := wm.provider.client.CallContract(ctx, ethereum.CallMsg{
			To:   &tokenContract,
			Data: data,
		}, nil)
		if err != nil {
			continue
		}

		var tokenID *big.Int
		if err := contractABI.UnpackIntoInterface(&tokenID, "tokenOfOwnerByIndex", result); err != nil {
			continue
		}

		credits = append(credits, tokenID)
	}

	return credits, nil
}

// ExportPrivateKey exports wallet private key (use with caution).
func (w *Wallet) ExportPrivateKey() string {
	return hex.EncodeToString(crypto.FromECDSA(w.PrivateKey))
}

// =============================================================================
// Smart Contract ABIs
// =============================================================================

// AnchorContractABI is the ABI for the hash anchoring contract.
const AnchorContractABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "hash", "type": "bytes32"},
			{"indexed": false, "name": "metadata", "type": "string"},
			{"indexed": false, "name": "timestamp", "type": "uint256"}
		],
		"name": "Anchored",
		"type": "event"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "hash", "type": "bytes32"},
			{"name": "metadata", "type": "string"}
		],
		"name": "anchor",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "hash", "type": "bytes32"}],
		"name": "getAnchor",
		"outputs": [
			{"name": "metadata", "type": "string"},
			{"name": "timestamp", "type": "uint256"}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	}
]`

// CarbonCreditNFTABI is the ABI for the carbon credit NFT contract.
const CarbonCreditNFTABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": true, "name": "tokenId", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "to", "type": "address"},
			{"name": "co2e", "type": "uint256"},
			{"name": "vintage", "type": "uint256"},
			{"name": "projectId", "type": "string"},
			{"name": "metadata", "type": "string"}
		],
		"name": "mint",
		"outputs": [{"name": "tokenId", "type": "uint256"}],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [{"name": "tokenId", "type": "uint256"}],
		"name": "retire",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "tokenId", "type": "uint256"}],
		"name": "credits",
		"outputs": [
			{"name": "co2e", "type": "uint256"},
			{"name": "vintage", "type": "uint256"},
			{"name": "projectId", "type": "string"},
			{"name": "retired", "type": "bool"}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{"name": "owner", "type": "address"},
			{"name": "index", "type": "uint256"}
		],
		"name": "tokenOfOwnerByIndex",
		"outputs": [{"name": "tokenId", "type": "uint256"}],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "from", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "tokenId", "type": "uint256"}
		],
		"name": "transferFrom",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
