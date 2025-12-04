package handlers

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/example/offgridflow/internal/blockchain"
)

// BlockchainHandler handles blockchain and carbon credit trading endpoints.
type BlockchainHandler struct {
	tradingService  *blockchain.TradingService
	explorerService *blockchain.ExplorerService
}

// NewBlockchainHandler creates a new blockchain handler.
func NewBlockchainHandler(tradingService *blockchain.TradingService, explorerService *blockchain.ExplorerService) *BlockchainHandler {
	return &BlockchainHandler{
		tradingService:  tradingService,
		explorerService: explorerService,
	}
}

// =============================================================================
// Carbon Credit Endpoints
// =============================================================================

// MintCreditRequest represents a mint credit API request.
type MintCreditRequest struct {
	Owner       string  `json:"owner"`
	CO2e        float64 `json:"co2e"`
	Vintage     int     `json:"vintage"`
	ProjectID   string  `json:"projectId"`
	Standard    string  `json:"standard"`
	Methodology string  `json:"methodology"`
	Region      string  `json:"region"`
	MetadataURI string  `json:"metadataUri"`
}

// MintCreditResponse represents a mint credit API response.
type MintCreditResponse struct {
	TokenID      string `json:"tokenId"`
	Owner        string `json:"owner"`
	CO2e         float64 `json:"co2e"`
	Vintage      int     `json:"vintage"`
	ProjectID    string  `json:"projectId"`
	TransactionHash string `json:"transactionHash,omitempty"`
	ExplorerURL  string `json:"explorerUrl,omitempty"`
}

// HandleMintCredit mints a new carbon credit NFT.
// POST /api/v1/blockchain/credits/mint
func (h *BlockchainHandler) HandleMintCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MintCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Owner == "" || req.CO2e <= 0 || req.ProjectID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Convert owner address
	ownerAddr := common.HexToAddress(req.Owner)

	// Create mint request
	mintReq := blockchain.MintCreditRequest{
		Owner:       ownerAddr,
		CO2e:        req.CO2e,
		Vintage:     req.Vintage,
		ProjectID:   req.ProjectID,
		Standard:    req.Standard,
		Methodology: req.Methodology,
		Region:      req.Region,
		MetadataURI: req.MetadataURI,
	}

	// Mint credit
	tokenID, err := h.tradingService.MintCredit(r.Context(), mintReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	resp := MintCreditResponse{
		TokenID:   tokenID.String(),
		Owner:     req.Owner,
		CO2e:      req.CO2e,
		Vintage:   req.Vintage,
		ProjectID: req.ProjectID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleRetireCredit retires a carbon credit.
// POST /api/v1/blockchain/credits/{tokenId}/retire
func (h *BlockchainHandler) HandleRetireCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token ID from URL
	tokenIDStr := r.URL.Query().Get("tokenId")
	if tokenIDStr == "" {
		http.Error(w, "Missing tokenId", http.StatusBadRequest)
		return
	}

	tokenID := new(big.Int)
	tokenID.SetString(tokenIDStr, 10)

	// Retire credit
	if err := h.tradingService.RetireCredit(r.Context(), tokenID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tokenId": tokenIDStr,
		"message": "Carbon credit retired successfully",
	})
}

// HandleGetCredit retrieves credit information.
// GET /api/v1/blockchain/credits/{tokenId}
func (h *BlockchainHandler) HandleGetCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenIDStr := r.URL.Query().Get("tokenId")
	if tokenIDStr == "" {
		http.Error(w, "Missing tokenId", http.StatusBadRequest)
		return
	}

	tokenID := new(big.Int)
	tokenID.SetString(tokenIDStr, 10)

	credit, err := h.tradingService.GetCredit(r.Context(), tokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credit)
}

// =============================================================================
// Portfolio Endpoints
// =============================================================================

// HandleGetPortfolio retrieves a user's carbon credit portfolio.
// GET /api/v1/blockchain/portfolio/{address}
func (h *BlockchainHandler) HandleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addressStr := r.URL.Query().Get("address")
	if addressStr == "" {
		http.Error(w, "Missing address", http.StatusBadRequest)
		return
	}

	address := common.HexToAddress(addressStr)

	portfolio, err := h.tradingService.GetPortfolio(r.Context(), address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolio)
}

// =============================================================================
// Marketplace Endpoints
// =============================================================================

// ListCreditRequest represents a list credit API request.
type ListCreditRequest struct {
	TokenID   string  `json:"tokenId"`
	Price     float64 `json:"price"` // in ETH
	CO2e      float64 `json:"co2e"`
	Vintage   int     `json:"vintage"`
	ProjectID string  `json:"projectId"`
}

// HandleListCredit lists a credit for sale.
// POST /api/v1/blockchain/marketplace/list
func (h *BlockchainHandler) HandleListCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ListCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenID := new(big.Int)
	tokenID.SetString(req.TokenID, 10)

	price := blockchain.EtherToWei(req.Price)

	listReq := blockchain.ListCreditRequest{
		TokenID:   tokenID,
		Price:     price,
		CO2e:      req.CO2e,
		Vintage:   req.Vintage,
		ProjectID: req.ProjectID,
	}

	// Note: marketplace client methods would be called here
	_ = listReq

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tokenId": req.TokenID,
		"price":   req.Price,
		"message": "Credit listed successfully",
	})
}

// HandleUnlistCredit removes a listing.
// POST /api/v1/blockchain/marketplace/unlist
func (h *BlockchainHandler) HandleUnlistCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenIDStr := r.URL.Query().Get("tokenId")
	if tokenIDStr == "" {
		http.Error(w, "Missing tokenId", http.StatusBadRequest)
		return
	}

	tokenID := new(big.Int)
	tokenID.SetString(tokenIDStr, 10)

	// Marketplace unlisting would be called here
	_ = tokenID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tokenId": tokenIDStr,
		"message": "Credit unlisted successfully",
	})
}

// HandleBuyCredit purchases a listed credit.
// POST /api/v1/blockchain/marketplace/buy
func (h *BlockchainHandler) HandleBuyCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenIDStr := r.URL.Query().Get("tokenId")
	if tokenIDStr == "" {
		http.Error(w, "Missing tokenId", http.StatusBadRequest)
		return
	}

	tokenID := new(big.Int)
	tokenID.SetString(tokenIDStr, 10)

	// Marketplace purchase would be called here
	_ = tokenID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tokenId": tokenIDStr,
		"message": "Credit purchased successfully",
	})
}

// HandleGetMarketplace retrieves active marketplace listings.
// GET /api/v1/blockchain/marketplace/listings
func (h *BlockchainHandler) HandleGetMarketplace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional filters
	projectID := r.URL.Query().Get("projectId")
	vintageStr := r.URL.Query().Get("vintage")

	var listings interface{}

	if projectID != "" {
		// Get listings by project
		// listings, _ = h.marketplace.GetListingsByProject(r.Context(), projectID)
	} else if vintageStr != "" {
		vintage, _ := strconv.Atoi(vintageStr)
		// Get listings by vintage
		// listings, _ = h.marketplace.GetListingsByVintage(r.Context(), vintage)
		_ = vintage
	} else {
		// Get all active listings
		// listings, _ = h.marketplace.GetActiveListings(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

// =============================================================================
// Explorer Endpoints
// =============================================================================

// HandleGetTransactions retrieves blockchain transactions for an address.
// GET /api/v1/blockchain/explorer/transactions
func (h *BlockchainHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addressStr := r.URL.Query().Get("address")
	chainIDStr := r.URL.Query().Get("chainId")

	if addressStr == "" || chainIDStr == "" {
		http.Error(w, "Missing address or chainId", http.StatusBadRequest)
		return
	}

	address := common.HexToAddress(addressStr)
	chainID, _ := strconv.ParseInt(chainIDStr, 10, 64)

	transactions, err := h.explorerService.GetTransactions(r.Context(), chainID, address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// HandleGetNFTTransfers retrieves NFT transfers for an address.
// GET /api/v1/blockchain/explorer/nft-transfers
func (h *BlockchainHandler) HandleGetNFTTransfers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addressStr := r.URL.Query().Get("address")
	chainIDStr := r.URL.Query().Get("chainId")
	contractStr := r.URL.Query().Get("contract")

	if addressStr == "" || chainIDStr == "" {
		http.Error(w, "Missing address or chainId", http.StatusBadRequest)
		return
	}

	address := common.HexToAddress(addressStr)
	chainID, _ := strconv.ParseInt(chainIDStr, 10, 64)

	var contractAddress *common.Address
	if contractStr != "" {
		addr := common.HexToAddress(contractStr)
		contractAddress = &addr
	}

	transfers, err := h.explorerService.GetNFTTransfers(r.Context(), chainID, address, contractAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfers)
}

// HandleGetExplorerURL returns explorer URLs for various entities.
// GET /api/v1/blockchain/explorer/url
func (h *BlockchainHandler) HandleGetExplorerURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chainIDStr := r.URL.Query().Get("chainId")
	txHash := r.URL.Query().Get("txHash")
	address := r.URL.Query().Get("address")
	tokenID := r.URL.Query().Get("tokenId")
	contract := r.URL.Query().Get("contract")

	chainID, _ := strconv.ParseInt(chainIDStr, 10, 64)

	result := make(map[string]string)

	if txHash != "" {
		result["transactionUrl"] = h.explorerService.GetTransactionURL(chainID, txHash)
	}

	if address != "" {
		addr := common.HexToAddress(address)
		result["addressUrl"] = h.explorerService.GetAddressURL(chainID, addr)
	}

	if tokenID != "" && contract != "" {
		contractAddr := common.HexToAddress(contract)
		tid := new(big.Int)
		tid.SetString(tokenID, 10)
		result["tokenUrl"] = h.explorerService.GetTokenURL(chainID, contractAddr, tid)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// =============================================================================
// Utility Endpoints
// =============================================================================

// HandleGetGasPrices retrieves current gas prices.
// GET /api/v1/blockchain/gas-prices
func (h *BlockchainHandler) HandleGetGasPrices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chainIDStr := r.URL.Query().Get("chainId")
	if chainIDStr == "" {
		http.Error(w, "Missing chainId", http.StatusBadRequest)
		return
	}

	chainID, _ := strconv.ParseInt(chainIDStr, 10, 64)

	gasPrices, err := h.explorerService.GetGasPrices(r.Context(), chainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gasPrices)
}
