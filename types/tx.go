package types

import (
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
)

type TxInfo struct {
	TxRecordHash []byte
	TxOrderHash  []byte
	BlockNumber  uint64
	Transaction  *types.TransactionRecord
}

func NewTxInfo(blockNo uint64, txRecord *types.TransactionRecord) (*TxInfo, error) {
	if txRecord == nil {
		return nil, fmt.Errorf("transaction record is nil")
	}
	txrHash := txRecord.Hash(crypto.SHA256)
	txoHash := txRecord.TransactionOrder.Hash(crypto.SHA256)

	txInfo := &TxInfo{
		TxRecordHash: txrHash,
		TxOrderHash:  txoHash,
		BlockNumber:  blockNo,
		Transaction:  txRecord,
	}
	return txInfo, nil
}
