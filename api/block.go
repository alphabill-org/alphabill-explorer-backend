package api

import (
	"crypto"
	"encoding/hex"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
)

type BlockInfo struct {
	_                  struct{} `cbor:",toarray"`
	Header             *types.Header
	TxHashes           []string
	UnicityCertificate *types.UnicityCertificate
}

func NewBlockInfo(b *types.Block) (*BlockInfo, error) {
	if b == nil {
		return nil, fmt.Errorf("block is nil")
	}
	txHashes := make([]string, 0, len(b.Transactions))

	for _, tx := range b.Transactions {
		hash := tx.Hash(crypto.SHA256) // crypto.SHA256?
		hashHex := hex.EncodeToString(hash)
		txHashes = append(txHashes, hashHex)
	}

	return &BlockInfo{
		Header:             b.Header,
		TxHashes:           txHashes,
		UnicityCertificate: b.UnicityCertificate,
	}, nil
}
