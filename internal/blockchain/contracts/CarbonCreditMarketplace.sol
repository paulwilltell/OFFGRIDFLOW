// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title CarbonCreditMarketplace
 * @dev Decentralized marketplace for trading carbon credit NFTs
 */
contract CarbonCreditMarketplace is Ownable, ReentrancyGuard, Pausable {
    // Marketplace fee (in basis points, e.g., 250 = 2.5%)
    uint256 public marketplaceFee = 250;
    uint256 public constant FEE_DENOMINATOR = 10000;

    // Carbon credit NFT contract
    IERC721 public carbonCreditNFT;

    // Listing structure
    struct Listing {
        uint256 tokenId;
        address seller;
        uint256 price; // in wei
        uint256 co2e; // for reference
        uint256 vintage;
        string projectId;
        bool active;
        uint256 listedAt;
    }

    // Offer structure
    struct Offer {
        uint256 tokenId;
        address buyer;
        uint256 price;
        uint256 expiresAt;
        bool active;
    }

    // Mappings
    mapping(uint256 => Listing) public listings;
    mapping(uint256 => Offer[]) public offers;
    uint256[] public activeListingIds;

    // Sales history
    struct Sale {
        uint256 tokenId;
        address seller;
        address buyer;
        uint256 price;
        uint256 timestamp;
    }
    Sale[] public salesHistory;

    // Events
    event Listed(
        uint256 indexed tokenId,
        address indexed seller,
        uint256 price,
        uint256 co2e
    );

    event Unlisted(uint256 indexed tokenId, address indexed seller);

    event PriceUpdated(
        uint256 indexed tokenId,
        uint256 oldPrice,
        uint256 newPrice
    );

    event Sold(
        uint256 indexed tokenId,
        address indexed seller,
        address indexed buyer,
        uint256 price
    );

    event OfferMade(
        uint256 indexed tokenId,
        address indexed buyer,
        uint256 price,
        uint256 expiresAt
    );

    event OfferAccepted(
        uint256 indexed tokenId,
        address indexed seller,
        address indexed buyer,
        uint256 price
    );

    event OfferCancelled(uint256 indexed tokenId, address indexed buyer);

    event FeeUpdated(uint256 oldFee, uint256 newFee);

    constructor(address _carbonCreditNFT) Ownable(msg.sender) {
        carbonCreditNFT = IERC721(_carbonCreditNFT);
    }

    /**
     * @dev Lists a carbon credit for sale
     * @param tokenId The NFT token ID
     * @param price Sale price in wei
     * @param co2e CO2 equivalent (for reference)
     * @param vintage Year of emission reduction
     * @param projectId Project identifier
     */
    function listCredit(
        uint256 tokenId,
        uint256 price,
        uint256 co2e,
        uint256 vintage,
        string memory projectId
    ) external whenNotPaused {
        require(
            carbonCreditNFT.ownerOf(tokenId) == msg.sender,
            "Not token owner"
        );
        require(price > 0, "Price must be positive");
        require(!listings[tokenId].active, "Already listed");

        // Transfer NFT to marketplace for escrow
        carbonCreditNFT.transferFrom(msg.sender, address(this), tokenId);

        Listing memory listing = Listing({
            tokenId: tokenId,
            seller: msg.sender,
            price: price,
            co2e: co2e,
            vintage: vintage,
            projectId: projectId,
            active: true,
            listedAt: block.timestamp
        });

        listings[tokenId] = listing;
        activeListingIds.push(tokenId);

        emit Listed(tokenId, msg.sender, price, co2e);
    }

    /**
     * @dev Cancels a listing
     * @param tokenId The NFT token ID
     */
    function unlistCredit(uint256 tokenId) external nonReentrant {
        Listing storage listing = listings[tokenId];
        require(listing.active, "Not listed");
        require(listing.seller == msg.sender, "Not seller");

        listing.active = false;

        // Return NFT to seller
        carbonCreditNFT.transferFrom(address(this), msg.sender, tokenId);

        // Remove from active listings
        _removeFromActiveListings(tokenId);

        emit Unlisted(tokenId, msg.sender);
    }

    /**
     * @dev Updates listing price
     * @param tokenId The NFT token ID
     * @param newPrice New price in wei
     */
    function updatePrice(uint256 tokenId, uint256 newPrice) external {
        Listing storage listing = listings[tokenId];
        require(listing.active, "Not listed");
        require(listing.seller == msg.sender, "Not seller");
        require(newPrice > 0, "Price must be positive");

        uint256 oldPrice = listing.price;
        listing.price = newPrice;

        emit PriceUpdated(tokenId, oldPrice, newPrice);
    }

    /**
     * @dev Buys a listed carbon credit
     * @param tokenId The NFT token ID
     */
    function buyCredit(uint256 tokenId)
        external
        payable
        nonReentrant
        whenNotPaused
    {
        Listing storage listing = listings[tokenId];
        require(listing.active, "Not listed");
        require(msg.value >= listing.price, "Insufficient payment");

        listing.active = false;
        address seller = listing.seller;
        uint256 price = listing.price;

        // Calculate fees
        uint256 fee = (price * marketplaceFee) / FEE_DENOMINATOR;
        uint256 sellerProceeds = price - fee;

        // Transfer NFT to buyer
        carbonCreditNFT.transferFrom(address(this), msg.sender, tokenId);

        // Transfer payment to seller
        (bool successSeller, ) = payable(seller).call{value: sellerProceeds}(
            ""
        );
        require(successSeller, "Seller payment failed");

        // Refund excess payment
        if (msg.value > price) {
            (bool successRefund, ) = payable(msg.sender).call{
                value: msg.value - price
            }("");
            require(successRefund, "Refund failed");
        }

        // Record sale
        salesHistory.push(
            Sale({
                tokenId: tokenId,
                seller: seller,
                buyer: msg.sender,
                price: price,
                timestamp: block.timestamp
            })
        );

        // Remove from active listings
        _removeFromActiveListings(tokenId);

        emit Sold(tokenId, seller, msg.sender, price);
    }

    /**
     * @dev Makes an offer on a listed credit
     * @param tokenId The NFT token ID
     * @param expiresAt Expiration timestamp
     */
    function makeOffer(uint256 tokenId, uint256 expiresAt)
        external
        payable
        whenNotPaused
    {
        require(listings[tokenId].active, "Not listed");
        require(msg.value > 0, "Offer must be positive");
        require(expiresAt > block.timestamp, "Invalid expiration");

        Offer memory offer = Offer({
            tokenId: tokenId,
            buyer: msg.sender,
            price: msg.value,
            expiresAt: expiresAt,
            active: true
        });

        offers[tokenId].push(offer);

        emit OfferMade(tokenId, msg.sender, msg.value, expiresAt);
    }

    /**
     * @dev Accepts an offer
     * @param tokenId The NFT token ID
     * @param offerIndex Index of the offer to accept
     */
    function acceptOffer(uint256 tokenId, uint256 offerIndex)
        external
        nonReentrant
    {
        Listing storage listing = listings[tokenId];
        require(listing.active, "Not listed");
        require(listing.seller == msg.sender, "Not seller");

        Offer storage offer = offers[tokenId][offerIndex];
        require(offer.active, "Offer not active");
        require(offer.expiresAt > block.timestamp, "Offer expired");

        listing.active = false;
        offer.active = false;

        address buyer = offer.buyer;
        uint256 price = offer.price;

        // Calculate fees
        uint256 fee = (price * marketplaceFee) / FEE_DENOMINATOR;
        uint256 sellerProceeds = price - fee;

        // Transfer NFT to buyer
        carbonCreditNFT.transferFrom(address(this), buyer, tokenId);

        // Transfer payment to seller
        (bool success, ) = payable(msg.sender).call{value: sellerProceeds}("");
        require(success, "Payment failed");

        // Record sale
        salesHistory.push(
            Sale({
                tokenId: tokenId,
                seller: msg.sender,
                buyer: buyer,
                price: price,
                timestamp: block.timestamp
            })
        );

        // Remove from active listings
        _removeFromActiveListings(tokenId);

        emit OfferAccepted(tokenId, msg.sender, buyer, price);
    }

    /**
     * @dev Cancels an offer
     * @param tokenId The NFT token ID
     * @param offerIndex Index of the offer to cancel
     */
    function cancelOffer(uint256 tokenId, uint256 offerIndex)
        external
        nonReentrant
    {
        Offer storage offer = offers[tokenId][offerIndex];
        require(offer.active, "Offer not active");
        require(offer.buyer == msg.sender, "Not offer maker");

        offer.active = false;

        // Refund offer amount
        (bool success, ) = payable(msg.sender).call{value: offer.price}("");
        require(success, "Refund failed");

        emit OfferCancelled(tokenId, msg.sender);
    }

    /**
     * @dev Gets all active listings
     * @return Array of active listing IDs
     */
    function getActiveListings() external view returns (uint256[] memory) {
        return activeListingIds;
    }

    /**
     * @dev Gets listings by project
     * @param projectId The project identifier
     * @return Array of token IDs
     */
    function getListingsByProject(string memory projectId)
        external
        view
        returns (uint256[] memory)
    {
        uint256 count = 0;
        for (uint256 i = 0; i < activeListingIds.length; i++) {
            if (
                listings[activeListingIds[i]].active &&
                keccak256(bytes(listings[activeListingIds[i]].projectId)) ==
                keccak256(bytes(projectId))
            ) {
                count++;
            }
        }

        uint256[] memory result = new uint256[](count);
        uint256 index = 0;
        for (uint256 i = 0; i < activeListingIds.length; i++) {
            if (
                listings[activeListingIds[i]].active &&
                keccak256(bytes(listings[activeListingIds[i]].projectId)) ==
                keccak256(bytes(projectId))
            ) {
                result[index] = activeListingIds[i];
                index++;
            }
        }

        return result;
    }

    /**
     * @dev Gets listings by vintage year
     * @param vintage The vintage year
     * @return Array of token IDs
     */
    function getListingsByVintage(uint256 vintage)
        external
        view
        returns (uint256[] memory)
    {
        uint256 count = 0;
        for (uint256 i = 0; i < activeListingIds.length; i++) {
            if (
                listings[activeListingIds[i]].active &&
                listings[activeListingIds[i]].vintage == vintage
            ) {
                count++;
            }
        }

        uint256[] memory result = new uint256[](count);
        uint256 index = 0;
        for (uint256 i = 0; i < activeListingIds.length; i++) {
            if (
                listings[activeListingIds[i]].active &&
                listings[activeListingIds[i]].vintage == vintage
            ) {
                result[index] = activeListingIds[i];
                index++;
            }
        }

        return result;
    }

    /**
     * @dev Gets all offers for a token
     * @param tokenId The NFT token ID
     * @return Array of offers
     */
    function getOffers(uint256 tokenId)
        external
        view
        returns (Offer[] memory)
    {
        return offers[tokenId];
    }

    /**
     * @dev Gets recent sales
     * @param count Number of recent sales to retrieve
     * @return Array of sales
     */
    function getRecentSales(uint256 count)
        external
        view
        returns (Sale[] memory)
    {
        if (count > salesHistory.length) {
            count = salesHistory.length;
        }

        Sale[] memory result = new Sale[](count);
        uint256 startIndex = salesHistory.length - count;

        for (uint256 i = 0; i < count; i++) {
            result[i] = salesHistory[startIndex + i];
        }

        return result;
    }

    /**
     * @dev Gets sales statistics
     * @return totalSales Total number of sales
     * @return totalVolume Total volume in wei
     * @return avgPrice Average sale price
     */
    function getSalesStats()
        external
        view
        returns (
            uint256 totalSales,
            uint256 totalVolume,
            uint256 avgPrice
        )
    {
        totalSales = salesHistory.length;
        totalVolume = 0;

        for (uint256 i = 0; i < salesHistory.length; i++) {
            totalVolume += salesHistory[i].price;
        }

        avgPrice = totalSales > 0 ? totalVolume / totalSales : 0;
    }

    /**
     * @dev Updates marketplace fee (only owner)
     * @param newFee New fee in basis points
     */
    function updateMarketplaceFee(uint256 newFee) external onlyOwner {
        require(newFee <= 1000, "Fee too high"); // Max 10%
        uint256 oldFee = marketplaceFee;
        marketplaceFee = newFee;
        emit FeeUpdated(oldFee, newFee);
    }

    /**
     * @dev Withdraws accumulated fees (only owner)
     */
    function withdrawFees() external onlyOwner nonReentrant {
        uint256 balance = address(this).balance;
        require(balance > 0, "No fees to withdraw");

        (bool success, ) = payable(owner()).call{value: balance}("");
        require(success, "Withdrawal failed");
    }

    /**
     * @dev Pauses marketplace
     */
    function pause() external onlyOwner {
        _pause();
    }

    /**
     * @dev Unpauses marketplace
     */
    function unpause() external onlyOwner {
        _unpause();
    }

    /**
     * @dev Removes token from active listings array
     * @param tokenId The token ID to remove
     */
    function _removeFromActiveListings(uint256 tokenId) private {
        for (uint256 i = 0; i < activeListingIds.length; i++) {
            if (activeListingIds[i] == tokenId) {
                activeListingIds[i] = activeListingIds[
                    activeListingIds.length - 1
                ];
                activeListingIds.pop();
                break;
            }
        }
    }

    // Fallback functions
    receive() external payable {}

    fallback() external payable {}
}
