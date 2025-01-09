package api

import (
	"crypto"
	"fmt"
	"github.com/alphabill-org/alphabill-go-base/types"
)

type TxInfo struct {
	TxRecordHash TxHash
	TxOrderHash  TxHash
	BlockNumber  uint64
	Transaction  *types.TransactionRecord
}

func NewTxInfo(blockNo uint64, txRecord *types.TransactionRecord) (*TxInfo, error) {
	if txRecord == nil {
		return nil, fmt.Errorf("transaction record is nil")
	}
	txrHash, err := txRecord.Hash(crypto.SHA256)
	if err != nil {
		return nil, err
	}
	txOrder := types.TransactionOrder{}
	if err = txOrder.UnmarshalCBOR(txRecord.TransactionOrder); err != nil {
		return nil, err
	}

	txoHash, err := txOrder.Hash(crypto.SHA256)
	if err != nil {
		return nil, err
	}

	txInfo := &TxInfo{
		TxRecordHash: txrHash,
		TxOrderHash:  txoHash,
		BlockNumber:  blockNo,
		Transaction:  txRecord,
	}
	return txInfo, nil
}
