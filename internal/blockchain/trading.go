// Package blockchain - Carbon credit trading service
package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// =============================================================================
// Trading Service
// =============================================================================

// TradingService provides high-level carbon credit trading operations.
type TradingService struct {
	tokenManager *TokenManager
	walletMgr    *WalletManager
	marketplace  *MarketplaceClient
	logger       *slog.Logger
}

// TradingServiceConfig configures the trading service.
type TradingServiceConfig struct {
	Provider            *EthereumProvider
	NFTContract         string
	MarketplaceContract string
	Logger              *slog.Logger
}

// NewTradingService creates a new trading service.
func NewTradingService(cfg TradingServiceConfig) *TradingService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &TradingService{
		tokenManager: NewTokenManager(cfg.Provider, cfg.NFTContract),
		walletMgr:    NewWalletManager(cfg.Provider),
		marketplace:  NewMarketplaceClient(cfg.Provider, cfg.MarketplaceContract),
		logger:       cfg.Logger.With("component", "trading-service"),
	}
}

// =============================================================================
// Credit Operations
// =============================================================================

// MintCreditRequest represents a request to mint a carbon credit.
type MintCreditRequest struct {
	Owner       common.Address
	CO2e        float64 // in tons
	Vintage     int
	ProjectID   string
	Standard    string // "VCS", "GoldStandard", etc.
	Methodology string
	Region      string
	ChainAnchor Hash
	MetadataURI string
}

// MintCredit mints a new carbon credit NFT.
func (ts *TradingService) MintCredit(ctx context.Context, req MintCreditRequest) (*big.Int, error) {
	ts.logger.Info("minting carbon credit",
		"owner", req.Owner.Hex(),
		"co2e", req.CO2e,
		"vintage", req.Vintage,
		"project", req.ProjectID)

	credit := CarbonCreditToken{
		Owner:       req.Owner,
		CO2e:        req.CO2e,
		Vintage:     req.Vintage,
		ProjectID:   req.ProjectID,
		Standard:    req.Standard,
		Metadata:    req.MetadataURI,
		ChainAnchor: req.ChainAnchor,
	}

	tokenID, err := ts.tokenManager.MintCarbonCredit(ctx, credit)
	if err != nil {
		return nil, fmt.Errorf("failed to mint credit: %w", err)
	}

	ts.logger.Info("successfully minted carbon credit",
		"tokenId", tokenID,
		"owner", req.Owner.Hex(),
		"co2e", req.CO2e)

	return tokenID, nil
}

// RetireCredit permanently retires a carbon credit.
func (ts *TradingService) RetireCredit(ctx context.Context, tokenID *big.Int) error {
	ts.logger.Info("retiring carbon credit", "tokenId", tokenID)

	if err := ts.tokenManager.RetireCredit(ctx, tokenID); err != nil {
		return fmt.Errorf("failed to retire credit: %w", err)
	}

	ts.logger.Info("successfully retired carbon credit", "tokenId", tokenID)
	return nil
}

// TransferCredit transfers a credit to another address.
func (ts *TradingService) TransferCredit(ctx context.Context, tokenID *big.Int, to common.Address) error {
	ts.logger.Info("transferring carbon credit",
		"tokenId", tokenID,
		"to", to.Hex())

	if err := ts.tokenManager.TransferCredit(ctx, tokenID, to); err != nil {
		return fmt.Errorf("failed to transfer credit: %w", err)
	}

	ts.logger.Info("successfully transferred carbon credit",
		"tokenId", tokenID,
		"to", to.Hex())

	return nil
}

// GetCredit retrieves credit information.
func (ts *TradingService) GetCredit(ctx context.Context, tokenID *big.Int) (*CarbonCreditToken, error) {
	return ts.tokenManager.GetCredit(ctx, tokenID)
}

// =============================================================================
// Marketplace Operations
// =============================================================================

// MarketplaceClient provides marketplace interaction.
type MarketplaceClient struct {
	provider *EthereumProvider
	contract common.Address
	logger   *slog.Logger
}

// ErrMarketplaceNotImplemented indicates the marketplace contract integration is unavailable.
var ErrMarketplaceNotImplemented = errors.New("marketplace operations are not implemented")

const carbonCreditMarketplaceABI = `[
  {
    "inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"uint256","name":"price","type":"uint256"},{"internalType":"uint256","name":"co2e","type":"uint256"},{"internalType":"uint256","name":"vintage","type":"uint256"},{"internalType":"string","name":"projectId","type":"string"}],
    "name":"listCredit",
    "outputs":[],
    "stateMutability":"nonpayable",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],
    "name":"unlistCredit",
    "outputs":[],
    "stateMutability":"nonpayable",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],
    "name":"buyCredit",
    "outputs":[],
    "stateMutability":"payable",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"uint256","name":"expiresAt","type":"uint256"}],
    "name":"makeOffer",
    "outputs":[],
    "stateMutability":"payable",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"uint256","name":"offerIndex","type":"uint256"}],
    "name":"acceptOffer",
    "outputs":[],
    "stateMutability":"nonpayable",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"","type":"uint256"}],
    "name":"listings",
    "outputs":[
      {"internalType":"uint256","name":"tokenId","type":"uint256"},
      {"internalType":"address","name":"seller","type":"address"},
      {"internalType":"uint256","name":"price","type":"uint256"},
      {"internalType":"uint256","name":"co2e","type":"uint256"},
      {"internalType":"uint256","name":"vintage","type":"uint256"},
      {"internalType":"string","name":"projectId","type":"string"},
      {"internalType":"bool","name":"active","type":"bool"},
      {"internalType":"uint256","name":"listedAt","type":"uint256"}
    ],
    "stateMutability":"view",
    "type":"function"
  },
  {
    "inputs":[],
    "name":"getActiveListings",
    "outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],
    "stateMutability":"view",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"string","name":"projectId","type":"string"}],
    "name":"getListingsByProject",
    "outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],
    "stateMutability":"view",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"vintage","type":"uint256"}],
    "name":"getListingsByVintage",
    "outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],
    "stateMutability":"view",
    "type":"function"
  },
  {
    "inputs":[],
    "name":"getSalesStats",
    "outputs":[
      {"internalType":"uint256","name":"totalSales","type":"uint256"},
      {"internalType":"uint256","name":"totalVolume","type":"uint256"},
      {"internalType":"uint256","name":"avgPrice","type":"uint256"}
    ],
    "stateMutability":"view",
    "type":"function"
  },
  {
    "inputs":[{"internalType":"uint256","name":"count","type":"uint256"}],
    "name":"getRecentSales",
    "outputs":[
      {"components":[
        {"internalType":"uint256","name":"tokenId","type":"uint256"},
        {"internalType":"address","name":"seller","type":"address"},
        {"internalType":"address","name":"buyer","type":"address"},
        {"internalType":"uint256","name":"price","type":"uint256"},
        {"internalType":"uint256","name":"timestamp","type":"uint256"}
      ],"internalType":"struct CarbonCreditMarketplace.Sale[]","name":"","type":"tuple[]"}
    ],
    "stateMutability":"view",
    "type":"function"
  }
]`

var marketplaceABI abi.ABI
var marketplaceABIOnce sync.Once
var marketplaceABIErr error

// NewMarketplaceClient creates a marketplace client.
func NewMarketplaceClient(provider *EthereumProvider, contractAddress string) *MarketplaceClient {
	return &MarketplaceClient{
		provider: provider,
		contract: common.HexToAddress(contractAddress),
		logger:   provider.logger.With("component", "marketplace-client"),
	}
}

func getMarketplaceABI() (abi.ABI, error) {
	marketplaceABIOnce.Do(func() {
		marketplaceABI, marketplaceABIErr = abi.JSON(strings.NewReader(carbonCreditMarketplaceABI))
	})
	return marketplaceABI, marketplaceABIErr
}

func (mc *MarketplaceClient) sendTx(ctx context.Context, data []byte, value *big.Int) (common.Hash, error) {
	if mc.provider == nil {
		return common.Hash{}, errors.New("marketplace client has no provider")
	}

	auth, err := mc.provider.createTransactor(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	if value == nil {
		value = big.NewInt(0)
	}

	gasLimit, err := mc.provider.client.EstimateGas(ctx, ethereum.CallMsg{
		From:  auth.From,
		To:    &mc.contract,
		Value: value,
		Data:  data,
	})
	if err != nil {
		gasLimit = auth.GasLimit
	}

	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		mc.contract,
		value,
		gasLimit,
		auth.GasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(mc.provider.chainID), mc.provider.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err := mc.provider.client.SendTransaction(ctx, signedTx); err != nil {
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}

func toMicroTons(value float64) *big.Int {
	micro := new(big.Float).Mul(big.NewFloat(value), big.NewFloat(1e6))
	out, _ := micro.Int(nil)
	if out == nil {
		return big.NewInt(0)
	}
	return out
}

func fromMicroTons(value *big.Int) float64 {
	if value == nil {
		return 0
	}
	micro := new(big.Float).SetInt(value)
	tons := new(big.Float).Quo(micro, big.NewFloat(1e6))
	result, _ := tons.Float64()
	return result
}

// Listing represents a marketplace listing.
type Listing struct {
	TokenID   *big.Int
	Seller    common.Address
	Price     *big.Int
	CO2e      float64
	Vintage   int
	ProjectID string
	Active    bool
	ListedAt  time.Time
}

func (mc *MarketplaceClient) fetchListing(ctx context.Context, tokenID *big.Int) (*Listing, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}
	if tokenID == nil {
		return nil, errors.New("token ID is required")
	}

	data, err := contractABI.Pack("listings", tokenID)
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var listing struct {
		TokenID   *big.Int
		Seller    common.Address
		Price     *big.Int
		CO2e      *big.Int
		Vintage   *big.Int
		ProjectID string
		Active    bool
		ListedAt  *big.Int
	}

	if err := contractABI.UnpackIntoInterface(&listing, "listings", result); err != nil {
		return nil, err
	}
	if listing.Vintage == nil {
		return nil, errors.New("marketplace listing missing vintage")
	}

	listedAt := time.Time{}
	if listing.ListedAt != nil {
		listedAt = time.Unix(listing.ListedAt.Int64(), 0)
	}

	return &Listing{
		TokenID:   listing.TokenID,
		Seller:    listing.Seller,
		Price:     listing.Price,
		CO2e:      fromMicroTons(listing.CO2e),
		Vintage:   int(listing.Vintage.Int64()),
		ProjectID: listing.ProjectID,
		Active:    listing.Active,
		ListedAt:  listedAt,
	}, nil
}

func (mc *MarketplaceClient) fetchActiveListingIDs(ctx context.Context) ([]*big.Int, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}

	data, err := contractABI.Pack("getActiveListings")
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var tokenIDs []*big.Int
	if err := contractABI.UnpackIntoInterface(&tokenIDs, "getActiveListings", result); err != nil {
		return nil, err
	}

	return tokenIDs, nil
}

func (mc *MarketplaceClient) listingsFromTokenIDs(ctx context.Context, tokenIDs []*big.Int) ([]Listing, error) {
	listings := make([]Listing, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		listing, err := mc.fetchListing(ctx, tokenID)
		if err != nil {
			mc.logger.Warn("failed to fetch listing", "tokenId", tokenID, "error", err)
			continue
		}
		if listing.Active {
			listings = append(listings, *listing)
		}
	}

	return listings, nil
}

// ListCreditRequest represents a request to list a credit.
type ListCreditRequest struct {
	TokenID   *big.Int
	Price     *big.Int // in wei
	CO2e      float64
	Vintage   int
	ProjectID string
}

// ListCredit lists a carbon credit for sale.
func (mc *MarketplaceClient) ListCredit(ctx context.Context, req ListCreditRequest) error {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return err
	}

	if req.TokenID == nil {
		return errors.New("token ID is required")
	}
	if req.Price == nil || req.Price.Sign() <= 0 {
		return errors.New("price must be positive")
	}
	if req.ProjectID == "" {
		return errors.New("project ID is required")
	}

	data, err := contractABI.Pack(
		"listCredit",
		req.TokenID,
		req.Price,
		toMicroTons(req.CO2e),
		big.NewInt(int64(req.Vintage)),
		req.ProjectID,
	)
	if err != nil {
		return err
	}

	txHash, err := mc.sendTx(ctx, data, big.NewInt(0))
	if err != nil {
		return err
	}

	mc.logger.Info("listed credit",
		"tokenId", req.TokenID,
		"price", req.Price,
		"txHash", txHash.Hex(),
	)
	return nil
}

// UnlistCredit removes a listing.
func (mc *MarketplaceClient) UnlistCredit(ctx context.Context, tokenID *big.Int) error {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return err
	}

	if tokenID == nil {
		return errors.New("token ID is required")
	}

	data, err := contractABI.Pack("unlistCredit", tokenID)
	if err != nil {
		return err
	}

	txHash, err := mc.sendTx(ctx, data, big.NewInt(0))
	if err != nil {
		return err
	}

	mc.logger.Info("unlisted credit", "tokenId", tokenID, "txHash", txHash.Hex())
	return nil
}

// BuyCredit purchases a listed credit.
func (mc *MarketplaceClient) BuyCredit(ctx context.Context, tokenID *big.Int, price *big.Int) error {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return err
	}

	if tokenID == nil {
		return errors.New("token ID is required")
	}
	if price == nil || price.Sign() <= 0 {
		return errors.New("price must be positive")
	}

	data, err := contractABI.Pack("buyCredit", tokenID)
	if err != nil {
		return err
	}

	txHash, err := mc.sendTx(ctx, data, price)
	if err != nil {
		return err
	}

	mc.logger.Info("purchased credit", "tokenId", tokenID, "price", price, "txHash", txHash.Hex())
	return nil
}

// MakeOffer makes an offer on a credit.
func (mc *MarketplaceClient) MakeOffer(ctx context.Context, tokenID *big.Int, price *big.Int, expiresAt time.Time) error {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return err
	}

	if tokenID == nil {
		return errors.New("token ID is required")
	}
	if price == nil || price.Sign() <= 0 {
		return errors.New("price must be positive")
	}
	if expiresAt.Before(time.Now()) {
		return errors.New("expiresAt must be in the future")
	}

	data, err := contractABI.Pack("makeOffer", tokenID, big.NewInt(expiresAt.Unix()))
	if err != nil {
		return err
	}

	txHash, err := mc.sendTx(ctx, data, price)
	if err != nil {
		return err
	}

	mc.logger.Info("offer created", "tokenId", tokenID, "price", price, "txHash", txHash.Hex())
	return nil
}

// AcceptOffer accepts an offer.
func (mc *MarketplaceClient) AcceptOffer(ctx context.Context, tokenID *big.Int, offerIndex int) error {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return err
	}

	if tokenID == nil {
		return errors.New("token ID is required")
	}
	if offerIndex < 0 {
		return errors.New("offer index must be non-negative")
	}

	data, err := contractABI.Pack("acceptOffer", tokenID, big.NewInt(int64(offerIndex)))
	if err != nil {
		return err
	}

	txHash, err := mc.sendTx(ctx, data, big.NewInt(0))
	if err != nil {
		return err
	}

	mc.logger.Info("offer accepted", "tokenId", tokenID, "offerIndex", offerIndex, "txHash", txHash.Hex())
	return nil
}

// GetActiveListings retrieves all active listings.
func (mc *MarketplaceClient) GetActiveListings(ctx context.Context) ([]Listing, error) {
	tokenIDs, err := mc.fetchActiveListingIDs(ctx)
	if err != nil {
		return nil, err
	}

	listings := make([]Listing, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		listing, err := mc.fetchListing(ctx, tokenID)
		if err != nil {
			mc.logger.Warn("failed to fetch listing", "tokenId", tokenID, "error", err)
			continue
		}
		if listing.Active {
			listings = append(listings, *listing)
		}
	}

	return listings, nil
}

// GetListingsByProject retrieves listings for a specific project.
func (mc *MarketplaceClient) GetListingsByProject(ctx context.Context, projectID string) ([]Listing, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(projectID) == "" {
		return nil, errors.New("project ID is required")
	}

	data, err := contractABI.Pack("getListingsByProject", projectID)
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var tokenIDs []*big.Int
	if err := contractABI.UnpackIntoInterface(&tokenIDs, "getListingsByProject", result); err != nil {
		return nil, err
	}

	return mc.listingsFromTokenIDs(ctx, tokenIDs)
}

// GetListingsByVintage retrieves listings for a specific vintage year.
func (mc *MarketplaceClient) GetListingsByVintage(ctx context.Context, vintage int) ([]Listing, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}
	if vintage <= 0 {
		return nil, errors.New("vintage must be positive")
	}

	data, err := contractABI.Pack("getListingsByVintage", big.NewInt(int64(vintage)))
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var tokenIDs []*big.Int
	if err := contractABI.UnpackIntoInterface(&tokenIDs, "getListingsByVintage", result); err != nil {
		return nil, err
	}

	return mc.listingsFromTokenIDs(ctx, tokenIDs)
}

// =============================================================================
// Portfolio Management
// =============================================================================

// Portfolio represents a user's carbon credit portfolio.
type Portfolio struct {
	Owner          common.Address
	TotalCredits   int
	ActiveCredits  int
	RetiredCredits int
	TotalCO2e      float64
	ActiveCO2e     float64
	RetiredCO2e    float64
	Credits        []*CarbonCreditToken
}

// GetPortfolio retrieves a user's complete portfolio.
func (ts *TradingService) GetPortfolio(ctx context.Context, owner common.Address) (*Portfolio, error) {
	ts.logger.Info("fetching portfolio", "owner", owner.Hex())

	// Get all credits owned by address
	creditIDs, err := ts.walletMgr.GetCredits(ctx, owner, common.HexToAddress(ts.tokenManager.contract.Hex()))
	if err != nil {
		return nil, fmt.Errorf("failed to get credits: %w", err)
	}

	portfolio := &Portfolio{
		Owner:   owner,
		Credits: make([]*CarbonCreditToken, 0, len(creditIDs)),
	}

	for _, tokenID := range creditIDs {
		credit, err := ts.tokenManager.GetCredit(ctx, tokenID)
		if err != nil {
			ts.logger.Warn("failed to get credit details",
				"tokenId", tokenID,
				"error", err)
			continue
		}

		portfolio.Credits = append(portfolio.Credits, credit)
		portfolio.TotalCredits++
		portfolio.TotalCO2e += credit.CO2e

		if credit.RetiredAt == nil {
			portfolio.ActiveCredits++
			portfolio.ActiveCO2e += credit.CO2e
		} else {
			portfolio.RetiredCredits++
			portfolio.RetiredCO2e += credit.CO2e
		}
	}

	ts.logger.Info("portfolio fetched",
		"owner", owner.Hex(),
		"totalCredits", portfolio.TotalCredits,
		"activeCO2e", portfolio.ActiveCO2e)

	return portfolio, nil
}

// =============================================================================
// Market Analytics
// =============================================================================

// MarketStats represents marketplace statistics.
type MarketStats struct {
	TotalSales     int
	TotalVolume    *big.Int
	AveragePrice   *big.Int
	ActiveListings int
	FloorPrice     *big.Int
	CeilingPrice   *big.Int
}

// GetMarketStats retrieves marketplace statistics.
func (mc *MarketplaceClient) GetMarketStats(ctx context.Context) (*MarketStats, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}

	data, err := contractABI.Pack("getSalesStats")
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var stats struct {
		TotalSales  *big.Int
		TotalVolume *big.Int
		AvgPrice    *big.Int
	}
	if err := contractABI.UnpackIntoInterface(&stats, "getSalesStats", result); err != nil {
		return nil, err
	}

	activeListings, err := mc.GetActiveListings(ctx)
	if err != nil {
		return nil, err
	}

	totalSales := 0
	if stats.TotalSales != nil {
		totalSales = int(stats.TotalSales.Int64())
	}
	totalVolume := stats.TotalVolume
	if totalVolume == nil {
		totalVolume = big.NewInt(0)
	}
	avgPrice := stats.AvgPrice
	if avgPrice == nil {
		avgPrice = big.NewInt(0)
	}

	var floorPrice *big.Int
	var ceilingPrice *big.Int
	for _, listing := range activeListings {
		if listing.Price == nil {
			continue
		}
		if floorPrice == nil || listing.Price.Cmp(floorPrice) < 0 {
			floorPrice = new(big.Int).Set(listing.Price)
		}
		if ceilingPrice == nil || listing.Price.Cmp(ceilingPrice) > 0 {
			ceilingPrice = new(big.Int).Set(listing.Price)
		}
	}

	if floorPrice == nil {
		floorPrice = big.NewInt(0)
	}
	if ceilingPrice == nil {
		ceilingPrice = big.NewInt(0)
	}

	return &MarketStats{
		TotalSales:     totalSales,
		TotalVolume:    totalVolume,
		AveragePrice:   avgPrice,
		ActiveListings: len(activeListings),
		FloorPrice:     floorPrice,
		CeilingPrice:   ceilingPrice,
	}, nil
}

// PriceHistory represents price history for a project.
type PriceHistory struct {
	ProjectID string
	Prices    []PricePoint
}

// PricePoint represents a price at a specific time.
type PricePoint struct {
	Price     *big.Int
	Timestamp time.Time
	TokenID   *big.Int
}

// GetPriceHistory retrieves price history for a project.
func (mc *MarketplaceClient) GetPriceHistory(ctx context.Context, projectID string, days int) (*PriceHistory, error) {
	contractABI, err := getMarketplaceABI()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(projectID) == "" {
		return nil, errors.New("project ID is required")
	}
	if days <= 0 {
		days = 30
	}

	maxCount := days * 10
	if maxCount < 10 {
		maxCount = 10
	}
	if maxCount > 500 {
		maxCount = 500
	}

	data, err := contractABI.Pack("getRecentSales", big.NewInt(int64(maxCount)))
	if err != nil {
		return nil, err
	}

	result, err := mc.provider.client.CallContract(ctx, ethereum.CallMsg{
		To:   &mc.contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var sales []struct {
		TokenID   *big.Int
		Seller    common.Address
		Buyer     common.Address
		Price     *big.Int
		Timestamp *big.Int
	}
	if err := contractABI.UnpackIntoInterface(&sales, "getRecentSales", result); err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	history := &PriceHistory{
		ProjectID: projectID,
		Prices:    make([]PricePoint, 0, len(sales)),
	}

	for _, sale := range sales {
		if sale.TokenID == nil || sale.Timestamp == nil || sale.Price == nil {
			continue
		}

		timestamp := time.Unix(sale.Timestamp.Int64(), 0)
		if timestamp.Before(cutoff) {
			continue
		}

		listing, err := mc.fetchListing(ctx, sale.TokenID)
		if err != nil {
			continue
		}
		if listing.ProjectID != projectID {
			continue
		}

		history.Prices = append(history.Prices, PricePoint{
			Price:     sale.Price,
			Timestamp: timestamp,
			TokenID:   sale.TokenID,
		})
	}

	return history, nil
}

// =============================================================================
// Batch Operations
// =============================================================================

// BatchMintRequest represents a batch mint request.
type BatchMintRequest struct {
	Owner   common.Address
	Credits []MintCreditRequest
}

// BatchMint mints multiple credits in a single transaction.
func (ts *TradingService) BatchMint(ctx context.Context, req BatchMintRequest) ([]*big.Int, error) {
	ts.logger.Info("batch minting credits",
		"owner", req.Owner.Hex(),
		"count", len(req.Credits))

	tokenIDs := make([]*big.Int, 0, len(req.Credits))

	for i, creditReq := range req.Credits {
		tokenID, err := ts.MintCredit(ctx, creditReq)
		if err != nil {
			ts.logger.Error("failed to mint credit in batch",
				"index", i,
				"error", err)
			return tokenIDs, fmt.Errorf("batch mint failed at index %d: %w", i, err)
		}
		tokenIDs = append(tokenIDs, tokenID)
	}

	ts.logger.Info("batch mint completed",
		"owner", req.Owner.Hex(),
		"minted", len(tokenIDs))

	return tokenIDs, nil
}

// BatchRetire retires multiple credits.
func (ts *TradingService) BatchRetire(ctx context.Context, tokenIDs []*big.Int) error {
	ts.logger.Info("batch retiring credits", "count", len(tokenIDs))

	for i, tokenID := range tokenIDs {
		if err := ts.RetireCredit(ctx, tokenID); err != nil {
			ts.logger.Error("failed to retire credit in batch",
				"index", i,
				"tokenId", tokenID,
				"error", err)
			return fmt.Errorf("batch retire failed at index %d: %w", i, err)
		}
	}

	ts.logger.Info("batch retire completed", "count", len(tokenIDs))
	return nil
}

// =============================================================================
// Utility Functions
// =============================================================================

// WeiToEther converts wei to ether.
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
}

// EtherToWei converts ether to wei.
func EtherToWei(ether float64) *big.Int {
	ethFloat := big.NewFloat(ether)
	weiFloat := new(big.Float).Mul(ethFloat, big.NewFloat(1e18))
	wei, _ := weiFloat.Int(nil)
	return wei
}

// FormatPrice formats a price for display.
func FormatPrice(price *big.Int) string {
	ether := WeiToEther(price)
	return fmt.Sprintf("%.4f ETH", ether)
}
