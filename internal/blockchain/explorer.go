// Package blockchain - Blockchain explorer integration
package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// =============================================================================
// Explorer Service
// =============================================================================

// ExplorerService provides blockchain explorer integration.
type ExplorerService struct {
	etherscanAPIKey string
	polygonscanAPIKey string
	baseURLs        map[int64]string // chainID -> base URL
	logger          *slog.Logger
	httpClient      *http.Client
}

// ExplorerConfig configures the explorer service.
type ExplorerConfig struct {
	EtherscanAPIKey   string
	PolygonscanAPIKey string
	Logger            *slog.Logger
}

// NewExplorerService creates a new explorer service.
func NewExplorerService(cfg ExplorerConfig) *ExplorerService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &ExplorerService{
		etherscanAPIKey:   cfg.EtherscanAPIKey,
		polygonscanAPIKey: cfg.PolygonscanAPIKey,
		baseURLs: map[int64]string{
			1:        "https://api.etherscan.io/api",           // Mainnet
			5:        "https://api-goerli.etherscan.io/api",    // Goerli
			11155111: "https://api-sepolia.etherscan.io/api",   // Sepolia
			137:      "https://api.polygonscan.com/api",        // Polygon
			80001:    "https://api-testnet.polygonscan.com/api", // Mumbai
		},
		logger:     cfg.Logger.With("component", "explorer-service"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// =============================================================================
// Transaction Information
// =============================================================================

// Transaction represents a blockchain transaction.
type Transaction struct {
	Hash             string    `json:"hash"`
	BlockNumber      string    `json:"blockNumber"`
	TimeStamp        string    `json:"timeStamp"`
	From             string    `json:"from"`
	To               string    `json:"to"`
	Value            string    `json:"value"`
	Gas              string    `json:"gas"`
	GasPrice         string    `json:"gasPrice"`
	GasUsed          string    `json:"gasUsed"`
	IsError          string    `json:"isError"`
	TxReceiptStatus  string    `json:"txreceipt_status"`
	Input            string    `json:"input"`
	ContractAddress  string    `json:"contractAddress"`
	MethodID         string    `json:"methodId"`
	FunctionName     string    `json:"functionName"`
}

// TransactionList represents a list of transactions.
type TransactionList struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []Transaction `json:"result"`
}

// GetTransactions retrieves transactions for an address.
func (es *ExplorerService) GetTransactions(ctx context.Context, chainID int64, address common.Address) ([]Transaction, error) {
	baseURL, ok := es.baseURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	apiKey := es.getAPIKey(chainID)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for chain %d", chainID)
	}

	params := url.Values{}
	params.Set("module", "account")
	params.Set("action", "txlist")
	params.Set("address", address.Hex())
	params.Set("startblock", "0")
	params.Set("endblock", "99999999")
	params.Set("sort", "desc")
	params.Set("apikey", apiKey)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var txList TransactionList
	if err := json.Unmarshal(body, &txList); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if txList.Status != "1" {
		return nil, fmt.Errorf("API error: %s", txList.Message)
	}

	es.logger.Info("fetched transactions",
		"address", address.Hex(),
		"count", len(txList.Result))

	return txList.Result, nil
}

// =============================================================================
// Token Transfers
// =============================================================================

// TokenTransfer represents an ERC-721 token transfer.
type TokenTransfer struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	TokenID           string `json:"tokenID"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
}

// TokenTransferList represents a list of token transfers.
type TokenTransferList struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  []TokenTransfer `json:"result"`
}

// GetNFTTransfers retrieves ERC-721 transfers for an address.
func (es *ExplorerService) GetNFTTransfers(ctx context.Context, chainID int64, address common.Address, contractAddress *common.Address) ([]TokenTransfer, error) {
	baseURL, ok := es.baseURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	apiKey := es.getAPIKey(chainID)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for chain %d", chainID)
	}

	params := url.Values{}
	params.Set("module", "account")
	params.Set("action", "tokennfttx")
	params.Set("address", address.Hex())
	if contractAddress != nil {
		params.Set("contractaddress", contractAddress.Hex())
	}
	params.Set("startblock", "0")
	params.Set("endblock", "99999999")
	params.Set("sort", "desc")
	params.Set("apikey", apiKey)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NFT transfers: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var transferList TokenTransferList
	if err := json.Unmarshal(body, &transferList); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if transferList.Status != "1" {
		return nil, fmt.Errorf("API error: %s", transferList.Message)
	}

	es.logger.Info("fetched NFT transfers",
		"address", address.Hex(),
		"count", len(transferList.Result))

	return transferList.Result, nil
}

// =============================================================================
// Contract Information
// =============================================================================

// ContractInfo represents smart contract information.
type ContractInfo struct {
	SourceCode           string `json:"SourceCode"`
	ABI                  string `json:"ABI"`
	ContractName         string `json:"ContractName"`
	CompilerVersion      string `json:"CompilerVersion"`
	OptimizationUsed     string `json:"OptimizationUsed"`
	Runs                 string `json:"Runs"`
	ConstructorArguments string `json:"ConstructorArguments"`
	EVMVersion           string `json:"EVMVersion"`
	Library              string `json:"Library"`
	LicenseType          string `json:"LicenseType"`
	Proxy                string `json:"Proxy"`
	Implementation       string `json:"Implementation"`
	SwarmSource          string `json:"SwarmSource"`
}

// ContractInfoResponse represents contract info API response.
type ContractInfoResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Result  []ContractInfo `json:"result"`
}

// GetContractInfo retrieves verified contract information.
func (es *ExplorerService) GetContractInfo(ctx context.Context, chainID int64, contractAddress common.Address) (*ContractInfo, error) {
	baseURL, ok := es.baseURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	apiKey := es.getAPIKey(chainID)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for chain %d", chainID)
	}

	params := url.Values{}
	params.Set("module", "contract")
	params.Set("action", "getsourcecode")
	params.Set("address", contractAddress.Hex())
	params.Set("apikey", apiKey)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contract info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var contractResp ContractInfoResponse
	if err := json.Unmarshal(body, &contractResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if contractResp.Status != "1" || len(contractResp.Result) == 0 {
		return nil, fmt.Errorf("contract not verified or not found")
	}

	return &contractResp.Result[0], nil
}

// =============================================================================
// Account Balance
// =============================================================================

// BalanceResponse represents balance API response.
type BalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// GetBalance retrieves ETH balance for an address.
func (es *ExplorerService) GetBalance(ctx context.Context, chainID int64, address common.Address) (*big.Int, error) {
	baseURL, ok := es.baseURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	apiKey := es.getAPIKey(chainID)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for chain %d", chainID)
	}

	params := url.Values{}
	params.Set("module", "account")
	params.Set("action", "balance")
	params.Set("address", address.Hex())
	params.Set("tag", "latest")
	params.Set("apikey", apiKey)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if balanceResp.Status != "1" {
		return nil, fmt.Errorf("API error: %s", balanceResp.Message)
	}

	balance := new(big.Int)
	balance.SetString(balanceResp.Result, 10)

	return balance, nil
}

// =============================================================================
// Gas Prices
// =============================================================================

// GasPrices represents current gas prices.
type GasPrices struct {
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
}

// GasPriceResponse represents gas price API response.
type GasPriceResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Result  GasPrices `json:"result"`
}

// GetGasPrices retrieves current gas prices.
func (es *ExplorerService) GetGasPrices(ctx context.Context, chainID int64) (*GasPrices, error) {
	baseURL, ok := es.baseURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	apiKey := es.getAPIKey(chainID)
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for chain %d", chainID)
	}

	params := url.Values{}
	params.Set("module", "gastracker")
	params.Set("action", "gasoracle")
	params.Set("apikey", apiKey)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gas prices: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gasPriceResp GasPriceResponse
	if err := json.Unmarshal(body, &gasPriceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if gasPriceResp.Status != "1" {
		return nil, fmt.Errorf("API error: %s", gasPriceResp.Message)
	}

	return &gasPriceResp.Result, nil
}

// =============================================================================
// URL Builders
// =============================================================================

// GetTransactionURL returns the explorer URL for a transaction.
func (es *ExplorerService) GetTransactionURL(chainID int64, txHash string) string {
	switch chainID {
	case 1:
		return fmt.Sprintf("https://etherscan.io/tx/%s", txHash)
	case 5:
		return fmt.Sprintf("https://goerli.etherscan.io/tx/%s", txHash)
	case 11155111:
		return fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
	case 137:
		return fmt.Sprintf("https://polygonscan.com/tx/%s", txHash)
	case 80001:
		return fmt.Sprintf("https://mumbai.polygonscan.com/tx/%s", txHash)
	default:
		return ""
	}
}

// GetAddressURL returns the explorer URL for an address.
func (es *ExplorerService) GetAddressURL(chainID int64, address common.Address) string {
	switch chainID {
	case 1:
		return fmt.Sprintf("https://etherscan.io/address/%s", address.Hex())
	case 5:
		return fmt.Sprintf("https://goerli.etherscan.io/address/%s", address.Hex())
	case 11155111:
		return fmt.Sprintf("https://sepolia.etherscan.io/address/%s", address.Hex())
	case 137:
		return fmt.Sprintf("https://polygonscan.com/address/%s", address.Hex())
	case 80001:
		return fmt.Sprintf("https://mumbai.polygonscan.com/address/%s", address.Hex())
	default:
		return ""
	}
}

// GetTokenURL returns the explorer URL for a token.
func (es *ExplorerService) GetTokenURL(chainID int64, contractAddress common.Address, tokenID *big.Int) string {
	switch chainID {
	case 1:
		return fmt.Sprintf("https://etherscan.io/nft/%s/%s", contractAddress.Hex(), tokenID.String())
	case 5:
		return fmt.Sprintf("https://goerli.etherscan.io/nft/%s/%s", contractAddress.Hex(), tokenID.String())
	case 11155111:
		return fmt.Sprintf("https://sepolia.etherscan.io/nft/%s/%s", contractAddress.Hex(), tokenID.String())
	case 137:
		return fmt.Sprintf("https://polygonscan.com/token/%s?a=%s", contractAddress.Hex(), tokenID.String())
	case 80001:
		return fmt.Sprintf("https://mumbai.polygonscan.com/token/%s?a=%s", contractAddress.Hex(), tokenID.String())
	default:
		return ""
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

// getAPIKey returns the appropriate API key for a chain.
func (es *ExplorerService) getAPIKey(chainID int64) string {
	switch chainID {
	case 1, 5, 11155111:
		return es.etherscanAPIKey
	case 137, 80001:
		return es.polygonscanAPIKey
	default:
		return ""
	}
}

// GetChainName returns the human-readable chain name.
func GetChainName(chainID int64) string {
	switch chainID {
	case 1:
		return "Ethereum Mainnet"
	case 5:
		return "Goerli Testnet"
	case 11155111:
		return "Sepolia Testnet"
	case 137:
		return "Polygon"
	case 80001:
		return "Mumbai Testnet"
	default:
		return fmt.Sprintf("Chain %d", chainID)
	}
}
