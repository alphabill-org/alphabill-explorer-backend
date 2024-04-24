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
	TargetUnits      []string
	TransactionOrder *types.TransactionOrder
	Fee              uint64
}

func NewTxInfo(blockNo uint64, txRecord *types.TransactionRecord) (*TxInfo, error) {
	if txRecord == nil {
		return nil, fmt.Errorf("transaction record is nil")
	}
	hashHex := hex.EncodeToString(txRecord.Hash(crypto.SHA256))

	units := make([]string, 0, len(txRecord.ServerMetadata.TargetUnits))

	for _, unit := range txRecord.ServerMetadata.TargetUnits {
		unitString := unit.String()
		units = append(units, unitString)
	}

	txInfo := &TxInfo{
		Hash:             hashHex,
		BlockNumber:      blockNo,
		Timeout:          txRecord.TransactionOrder.Timeout(),
		PayloadType:      txRecord.TransactionOrder.PayloadType(),
		Status:           &txRecord.ServerMetadata.SuccessIndicator,
		TargetUnits:      units,
		TransactionOrder: txRecord.TransactionOrder,
		Fee:              txRecord.ServerMetadata.GetActualFee(),
	}
	return txInfo, nil
}
