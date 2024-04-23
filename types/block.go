package types

import (
	"crypto"
	"encoding/hex"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
)

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
