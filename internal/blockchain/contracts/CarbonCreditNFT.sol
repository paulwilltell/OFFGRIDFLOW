// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

/**
 * @title CarbonCreditNFT
 * @dev Carbon credit NFT (ERC-721) representing verified carbon offsets
 */
contract CarbonCreditNFT is
    ERC721,
    ERC721Enumerable,
    ERC721URIStorage,
    AccessControl,
    Pausable
{
    using Counters for Counters.Counter;

    // Roles
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");

    // Counter for token IDs
    Counters.Counter private _tokenIdCounter;

    // Carbon Credit structure
    struct CarbonCredit {
        uint256 co2e; // CO2 equivalent in micro-tons (1e-6 tons)
        uint256 vintage; // Year of emission reduction
        string projectId; // Project identifier
        string standard; // "VCS", "GoldStandard", "CAR", etc.
        string methodology; // Methodology used
        string region; // Geographic region
        bool retired; // Whether the credit is retired
        uint256 mintedAt; // Timestamp when minted
        uint256 retiredAt; // Timestamp when retired
        address retiredBy; // Who retired the credit
        bytes32 chainAnchor; // Hash anchored to OffGridFlow chain
    }

    // Mapping from token ID to credit data
    mapping(uint256 => CarbonCredit) public credits;

    // Mapping from project ID to array of token IDs
    mapping(string => uint256[]) public projectCredits;

    // Total CO2e across all active credits
    uint256 public totalActiveCO2e;

    // Total CO2e retired
    uint256 public totalRetiredCO2e;

    // Events
    event CreditMinted(
        uint256 indexed tokenId,
        address indexed owner,
        uint256 co2e,
        uint256 vintage,
        string projectId
    );

    event CreditRetired(
        uint256 indexed tokenId,
        address indexed retiredBy,
        uint256 co2e,
        uint256 timestamp
    );

    event CreditTransferred(
        uint256 indexed tokenId,
        address indexed from,
        address indexed to,
        uint256 co2e
    );

    constructor() ERC721("Carbon Credit NFT", "CCNFT") {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(MINTER_ROLE, msg.sender);
        _grantRole(VERIFIER_ROLE, msg.sender);
    }

    /**
     * @dev Mints a new carbon credit NFT
     * @param to Address to mint to
     * @param co2e CO2 equivalent in micro-tons
     * @param vintage Year of emission reduction
     * @param projectId Project identifier
     * @param standard Carbon standard
     * @param methodology Methodology used
     * @param region Geographic region
     * @param chainAnchor Hash anchored to OffGridFlow chain
     * @param tokenURI Metadata URI
     * @return tokenId The minted token ID
     */
    function mint(
        address to,
        uint256 co2e,
        uint256 vintage,
        string memory projectId,
        string memory standard,
        string memory methodology,
        string memory region,
        bytes32 chainAnchor,
        string memory tokenURI
    ) external onlyRole(MINTER_ROLE) whenNotPaused returns (uint256) {
        require(to != address(0), "Cannot mint to zero address");
        require(co2e > 0, "CO2e must be positive");
        require(vintage >= 2000 && vintage <= 2100, "Invalid vintage year");
        require(bytes(projectId).length > 0, "Project ID required");

        uint256 tokenId = _tokenIdCounter.current();
        _tokenIdCounter.increment();

        _safeMint(to, tokenId);
        _setTokenURI(tokenId, tokenURI);

        CarbonCredit memory credit = CarbonCredit({
            co2e: co2e,
            vintage: vintage,
            projectId: projectId,
            standard: standard,
            methodology: methodology,
            region: region,
            retired: false,
            mintedAt: block.timestamp,
            retiredAt: 0,
            retiredBy: address(0),
            chainAnchor: chainAnchor
        });

        credits[tokenId] = credit;
        projectCredits[projectId].push(tokenId);
        totalActiveCO2e += co2e;

        emit CreditMinted(tokenId, to, co2e, vintage, projectId);

        return tokenId;
    }

    /**
     * @dev Batch mint multiple credits
     * @param to Address to mint to
     * @param creditData Array of credit data
     * @return tokenIds Array of minted token IDs
     */
    struct CreditData {
        uint256 co2e;
        uint256 vintage;
        string projectId;
        string standard;
        string methodology;
        string region;
        bytes32 chainAnchor;
        string tokenURI;
    }

    function batchMint(address to, CreditData[] memory creditData)
        external
        onlyRole(MINTER_ROLE)
        whenNotPaused
        returns (uint256[] memory)
    {
        require(creditData.length > 0, "Empty credit data");

        uint256[] memory tokenIds = new uint256[](creditData.length);

        for (uint256 i = 0; i < creditData.length; i++) {
            tokenIds[i] = this.mint(
                to,
                creditData[i].co2e,
                creditData[i].vintage,
                creditData[i].projectId,
                creditData[i].standard,
                creditData[i].methodology,
                creditData[i].region,
                creditData[i].chainAnchor,
                creditData[i].tokenURI
            );
        }

        return tokenIds;
    }

    /**
     * @dev Retires a carbon credit (permanent offset)
     * @param tokenId The token to retire
     */
    function retire(uint256 tokenId) external {
        require(_exists(tokenId), "Token does not exist");
        require(ownerOf(tokenId) == msg.sender, "Not token owner");
        require(!credits[tokenId].retired, "Already retired");

        CarbonCredit storage credit = credits[tokenId];
        credit.retired = true;
        credit.retiredAt = block.timestamp;
        credit.retiredBy = msg.sender;

        totalActiveCO2e -= credit.co2e;
        totalRetiredCO2e += credit.co2e;

        // Burn the NFT (optional - some protocols keep it for proof)
        // _burn(tokenId);

        emit CreditRetired(tokenId, msg.sender, credit.co2e, block.timestamp);
    }

    /**
     * @dev Batch retire multiple credits
     * @param tokenIds Array of token IDs to retire
     */
    function batchRetire(uint256[] memory tokenIds) external {
        for (uint256 i = 0; i < tokenIds.length; i++) {
            this.retire(tokenIds[i]);
        }
    }

    /**
     * @dev Gets credit information
     * @param tokenId The token ID
     * @return Credit data
     */
    function getCredit(uint256 tokenId)
        external
        view
        returns (CarbonCredit memory)
    {
        require(_exists(tokenId), "Token does not exist");
        return credits[tokenId];
    }

    /**
     * @dev Gets all credits for a project
     * @param projectId The project identifier
     * @return Array of token IDs
     */
    function getProjectCredits(string memory projectId)
        external
        view
        returns (uint256[] memory)
    {
        return projectCredits[projectId];
    }

    /**
     * @dev Gets all credits owned by an address
     * @param owner The owner address
     * @return Array of token IDs
     */
    function getCreditsByOwner(address owner)
        external
        view
        returns (uint256[] memory)
    {
        uint256 balance = balanceOf(owner);
        uint256[] memory result = new uint256[](balance);

        for (uint256 i = 0; i < balance; i++) {
            result[i] = tokenOfOwnerByIndex(owner, i);
        }

        return result;
    }

    /**
     * @dev Gets active (non-retired) credits for an owner
     * @param owner The owner address
     * @return Array of token IDs
     */
    function getActiveCredits(address owner)
        external
        view
        returns (uint256[] memory)
    {
        uint256 balance = balanceOf(owner);
        uint256[] memory temp = new uint256[](balance);
        uint256 activeCount = 0;

        for (uint256 i = 0; i < balance; i++) {
            uint256 tokenId = tokenOfOwnerByIndex(owner, i);
            if (!credits[tokenId].retired) {
                temp[activeCount] = tokenId;
                activeCount++;
            }
        }

        // Copy to result array with correct size
        uint256[] memory result = new uint256[](activeCount);
        for (uint256 i = 0; i < activeCount; i++) {
            result[i] = temp[i];
        }

        return result;
    }

    /**
     * @dev Gets total active CO2e for an owner
     * @param owner The owner address
     * @return Total CO2e in micro-tons
     */
    function getOwnerActiveCO2e(address owner) external view returns (uint256) {
        uint256 balance = balanceOf(owner);
        uint256 total = 0;

        for (uint256 i = 0; i < balance; i++) {
            uint256 tokenId = tokenOfOwnerByIndex(owner, i);
            if (!credits[tokenId].retired) {
                total += credits[tokenId].co2e;
            }
        }

        return total;
    }

    /**
     * @dev Pauses all token transfers
     */
    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpauses all token transfers
     */
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    // Override required functions

    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 tokenId,
        uint256 batchSize
    ) internal override(ERC721, ERC721Enumerable) whenNotPaused {
        require(!credits[tokenId].retired, "Cannot transfer retired credit");
        super._beforeTokenTransfer(from, to, tokenId, batchSize);

        if (from != address(0) && to != address(0)) {
            emit CreditTransferred(tokenId, from, to, credits[tokenId].co2e);
        }
    }

    function _burn(uint256 tokenId)
        internal
        override(ERC721, ERC721URIStorage)
    {
        super._burn(tokenId);
    }

    function tokenURI(uint256 tokenId)
        public
        view
        override(ERC721, ERC721URIStorage)
        returns (string memory)
    {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721, ERC721Enumerable, AccessControl)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
