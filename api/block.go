package api

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
}

func NewBlockInfo(b *types.Block) (*BlockInfo, error) {
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

	return &BlockInfo{
		Header:             b.Header,
		TxHashes:           txHashes,
		UnicityCertificate: b.UnicityCertificate,
	}, nil
}
