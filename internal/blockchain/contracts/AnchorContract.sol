// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title AnchorContract
 * @dev Anchors hashes to Ethereum for immutable audit trail
 */
contract AnchorContract is Ownable {
    struct AnchorRecord {
        bytes32 hash;
        string metadata;
        uint256 timestamp;
        address anchorer;
    }

    // Mapping from hash to anchor record
    mapping(bytes32 => AnchorRecord) public anchors;

    // Array of all anchored hashes
    bytes32[] public anchoredHashes;

    // Events
    event Anchored(
        bytes32 indexed hash,
        string metadata,
        uint256 timestamp,
        address indexed anchorer
    );

    event BatchAnchored(
        bytes32[] hashes,
        uint256 timestamp,
        address indexed anchorer
    );

    constructor() Ownable(msg.sender) {}

    /**
     * @dev Anchors a single hash with metadata
     * @param hash The hash to anchor
     * @param metadata Additional metadata as JSON string
     */
    function anchor(bytes32 hash, string memory metadata) external {
        require(anchors[hash].timestamp == 0, "Hash already anchored");

        AnchorRecord memory record = AnchorRecord({
            hash: hash,
            metadata: metadata,
            timestamp: block.timestamp,
            anchorer: msg.sender
        });

        anchors[hash] = record;
        anchoredHashes.push(hash);

        emit Anchored(hash, metadata, block.timestamp, msg.sender);
    }

    /**
     * @dev Anchors multiple hashes in a single transaction
     * @param hashes Array of hashes to anchor
     * @param metadatas Array of metadata strings
     */
    function batchAnchor(bytes32[] memory hashes, string[] memory metadatas) external {
        require(hashes.length == metadatas.length, "Array length mismatch");
        require(hashes.length > 0, "Empty arrays");

        for (uint256 i = 0; i < hashes.length; i++) {
            require(anchors[hashes[i]].timestamp == 0, "Hash already anchored");

            AnchorRecord memory record = AnchorRecord({
                hash: hashes[i],
                metadata: metadatas[i],
                timestamp: block.timestamp,
                anchorer: msg.sender
            });

            anchors[hashes[i]] = record;
            anchoredHashes.push(hashes[i]);
        }

        emit BatchAnchored(hashes, block.timestamp, msg.sender);
    }

    /**
     * @dev Retrieves anchor information for a hash
     * @param hash The hash to lookup
     * @return metadata The metadata string
     * @return timestamp When the hash was anchored
     * @return anchorer Who anchored the hash
     */
    function getAnchor(bytes32 hash)
        external
        view
        returns (
            string memory metadata,
            uint256 timestamp,
            address anchorer
        )
    {
        AnchorRecord memory record = anchors[hash];
        require(record.timestamp != 0, "Hash not found");

        return (record.metadata, record.timestamp, record.anchorer);
    }

    /**
     * @dev Verifies if a hash is anchored
     * @param hash The hash to verify
     * @return True if anchored, false otherwise
     */
    function isAnchored(bytes32 hash) external view returns (bool) {
        return anchors[hash].timestamp != 0;
    }

    /**
     * @dev Gets total number of anchored hashes
     * @return The count of anchored hashes
     */
    function getAnchoredCount() external view returns (uint256) {
        return anchoredHashes.length;
    }

    /**
     * @dev Gets a range of anchored hashes
     * @param start Start index
     * @param count Number of hashes to retrieve
     * @return Array of hashes
     */
    function getAnchoredHashes(uint256 start, uint256 count)
        external
        view
        returns (bytes32[] memory)
    {
        require(start < anchoredHashes.length, "Start index out of bounds");

        uint256 end = start + count;
        if (end > anchoredHashes.length) {
            end = anchoredHashes.length;
        }

        bytes32[] memory result = new bytes32[](end - start);
        for (uint256 i = start; i < end; i++) {
            result[i - start] = anchoredHashes[i];
        }

        return result;
    }
}
