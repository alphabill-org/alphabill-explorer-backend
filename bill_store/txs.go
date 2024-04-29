package bill_store

import (
	"encoding/json"
	"fmt"

	exTypes "github.com/alphabill-org/alphabill-explorer-backend/types"
	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetTxInfo(txInfo *exTypes.TxInfo) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		txInfoBytes, err := json.Marshal(txInfo)
		if err != nil {
			return err
		}
		txInfoBucket := tx.Bucket(txInfoBucket)
		hashBytes := []byte(txInfo.Hash)
		err = txInfoBucket.Put(hashBytes, txInfoBytes,)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *boltBillStore) GetTxInfo(txHash string) (*exTypes.TxInfo, error) {
	var txInfo *exTypes.TxInfo
	hashBytes := []byte(txHash)
	err := s.db.View(func(tx *bolt.Tx) error {
		txInforBytes := tx.Bucket(txInfoBucket).Get(hashBytes)
		return json.Unmarshal(txInforBytes, &txInfo)
	})
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (s *boltBillStore) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*exTypes.TxInfo, err error) {
	return res, s.db.View(func(tx *bolt.Tx) error {
		var err error
		res, err = s.getBlockTxsByBlockNumber(tx, blockNumber)
		return err
	})
}

func (s *boltBillStore) getBlockTxsByBlockNumber(tx *bolt.Tx, blockNumber uint64) ([]*exTypes.TxInfo, error) {
	var txs []*exTypes.TxInfo
	blockNumberBytes := util.Uint64ToBytes(blockNumber)

	blockInfoBytes := tx.Bucket(blockInfoBucket).Get(blockNumberBytes)
	if blockInfoBytes == nil {
		return nil, fmt.Errorf("no block data found for block number %d", blockNumber)
	}

	var b exTypes.BlockInfo
	if err := json.Unmarshal(blockInfoBytes, &b); err != nil {
		return nil, err
	}

	txInfoBucket := tx.Bucket(txInfoBucket)
	for _, hash := range b.TxHashes {
		hashBytes := []byte(hash)
		txBytes := txInfoBucket.Get(hashBytes)
		if txBytes == nil {
			continue
		}

		t := &exTypes.TxInfo{}
		if err := json.Unmarshal(txBytes, t); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}

	return txs, nil
}
