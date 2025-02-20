package domain

import (
	"crypto"
	"fmt"

	"github.com/alphabill-org/alphabill-go-base/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TxInfo struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's ObjectId field
	TxRecordHash TxHash
	TxOrderHash  TxHash
	BlockNumber  uint64
	Transaction  *types.TransactionRecord
	PartitionID  types.PartitionID
}

func NewTxInfo(PartitionID types.PartitionID, blockNo uint64, txRecord *types.TransactionRecord) (*TxInfo, error) {
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
		PartitionID:  PartitionID,
	}
	return txInfo, nil
}
