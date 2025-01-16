package domain

import (
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill-go-base/types"
)

type BlockInfo struct {
	_                  struct{} `cbor:",toarray"`
	Header             *types.Header
	TxHashes           []TxHash
	UnicityCertificate types.TaggedCBOR
	PartitionID        types.PartitionID
	PartitionTypeID    types.PartitionTypeID
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

	return &BlockInfo{
		Header:             b.Header,
		TxHashes:           txHashes,
		UnicityCertificate: b.UnicityCertificate,
		PartitionID:        b.PartitionID(),
		PartitionTypeID:    partitionTypeID,
		BlockNumber:        roundNumber,
	}, nil
}
