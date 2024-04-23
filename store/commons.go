package store

import (
	"crypto"
	"encoding/hex"
	"fmt"

	"github.com/alphabill-org/alphabill/types"

)

func CreateTxExplorer(blockNo uint64, txRecord *types.TransactionRecord) (*TxExplorer, error) {
	if txRecord == nil {
		return nil, fmt.Errorf("transaction record is nil")
	}
	hashHex := hex.EncodeToString(txRecord.Hash(crypto.SHA256))
	txExplorer := &TxExplorer{
		Hash:             hashHex,
		BlockNumber:      blockNo,
		Timeout:          txRecord.TransactionOrder.Timeout(),
		PayloadType:      txRecord.TransactionOrder.PayloadType(),
		Status:           &txRecord.ServerMetadata.SuccessIndicator,
		TargetUnits:      []types.UnitID{},
		TransactionOrder: txRecord.TransactionOrder,
		Fee:              txRecord.ServerMetadata.GetActualFee(),
	}
	txExplorer.TargetUnits = txRecord.ServerMetadata.TargetUnits
	return txExplorer, nil
}

func CreateBlockInfo(b *types.Block) (*BlockInfo, error) {
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
