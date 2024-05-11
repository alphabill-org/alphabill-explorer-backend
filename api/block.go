package api

import (
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
)

type BlockInfo struct {
	_                  struct{} `cbor:",toarray"`
	Header             *types.Header
	TxHashes           []TxHash
	UnicityCertificate *types.UnicityCertificate
}

func NewBlockInfo(b *types.Block) (*BlockInfo, error) {
	if b == nil {
		return nil, fmt.Errorf("block is nil")
	}
	txHashes := make([]TxHash, 0, len(b.Transactions))

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash(crypto.SHA256))
	}

	return &BlockInfo{
		Header:             b.Header,
		TxHashes:           txHashes,
		UnicityCertificate: b.UnicityCertificate,
	}, nil
}
