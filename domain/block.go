package domain

import (
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
)

type BlockInfo struct {
	PartitionID        types.PartitionID
	PartitionTypeID    types.PartitionTypeID
	ShardID            types.ShardID
	ProposerID         string
	PreviousBlockHash  hex.Bytes
	TxHashes           []TxHash
	UnicityCertificate types.TaggedCBOR
	BlockNumber        uint64
}

func NewBlockInfo(b *types.Block, partitionTypeID types.PartitionTypeID) (*BlockInfo, error) {
	if b == nil {
		return nil, fmt.Errorf("block is nil")
	}
	txHashes := make([]TxHash, 0, len(b.Transactions))

	for _, tx := range b.Transactions {
		hash, err := tx.Hash(crypto.SHA256)
		if err != nil {
			return nil, err
		}
		txHashes = append(txHashes, hash)
	}

	roundNumber, err := b.GetRoundNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to get round number from block: %w", err)
	}

	var (
		shardID           types.ShardID
		previousBlockHash hex.Bytes
	)
	if b.Header != nil {
		shardID = b.Header.ShardID
		previousBlockHash = b.Header.PreviousBlockHash
	}

	return &BlockInfo{
		PartitionID:        b.PartitionID(),
		PartitionTypeID:    partitionTypeID,
		ShardID:            shardID,
		ProposerID:         b.GetProposerID(),
		PreviousBlockHash:  previousBlockHash,
		TxHashes:           txHashes,
		UnicityCertificate: b.UnicityCertificate,
		BlockNumber:        roundNumber,
	}, nil
}
