package store

import (
	"crypto"
	"encoding/hex"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
	bolt "go.etcd.io/bbolt"
)

func CreateBuckets(update func(fn func(*bolt.Tx) error) error, buckets ...[]byte) error {
	return update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func EnsureSubBucket(tx *bolt.Tx, parentBucket []byte, bucket []byte, allowAbsent bool) (*bolt.Bucket, error) {
	pb := tx.Bucket(parentBucket)
	if pb == nil {
		return nil, fmt.Errorf("bucket %s not found", parentBucket)
	}
	b := pb.Bucket(bucket)
	if b == nil {
		if tx.Writable() {
			return pb.CreateBucket(bucket)
		}
		if allowAbsent {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to ensure bucket %s/%X", parentBucket, bucket)
	}
	return b, nil
}

func CreateBlockExplorer(b *types.Block) (*BlockExplorer, error) {
	if b == nil {
		return nil, fmt.Errorf("block is nil")
	}
	var txHashes []string

	for _, tx := range b.Transactions {
		hash := tx.Hash(crypto.SHA256) // crypto.SHA256?
		hashHex := hex.EncodeToString(hash)
		txHashes = append(txHashes, hashHex)
	}

	header := &HeaderExplorer{
		Timestamp:         b.UnicityCertificate.UnicitySeal.Timestamp,
		BlockHash:         b.UnicityCertificate.InputRecord.BlockHash,
		PreviousBlockHash: b.Header.PreviousBlockHash,
		ProposerID:        b.GetProposerID(),
	}
	blockExplorer := &BlockExplorer{
		SystemID:        &b.Header.SystemID,
		RoundNumber:     b.GetRoundNumber(),
		Header:          header,
		TxHashes:        txHashes,
		SummaryValue:    b.UnicityCertificate.InputRecord.SummaryValue,
		SumOfEarnedFees: b.UnicityCertificate.InputRecord.SumOfEarnedFees,
	}
	return blockExplorer, nil
}

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
