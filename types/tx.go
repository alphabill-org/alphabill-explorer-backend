package types

import (
	"crypto"
	"encoding/hex"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
)

type TxInfo struct {
	_                struct{} `cbor:",toarray"`
	Hash             string
	BlockNumber      uint64
	Timeout          uint64
	PayloadType      string
	Status           *types.TxStatus
	TargetUnits      []types.UnitID
	TransactionOrder *types.TransactionOrder
	Fee              uint64
}

func NewTxInfo(blockNo uint64, txRecord *types.TransactionRecord) (*TxInfo, error) {
	if txRecord == nil {
		return nil, fmt.Errorf("transaction record is nil")
	}
	hashHex := hex.EncodeToString(txRecord.Hash(crypto.SHA256))
	txExplorer := &TxInfo{
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
