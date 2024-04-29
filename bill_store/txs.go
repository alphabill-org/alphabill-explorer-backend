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
		err = txInfoBucket.Put(hashBytes, txInfoBytes)
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

func (s *boltBillStore) GetTxsByUnitID(unitID string) ([]*exTypes.TxInfo, error) {
	var txs []*exTypes.TxInfo

	err := s.db.View(func(tx *bolt.Tx) error {
		unitB := tx.Bucket(unitBucket)
		if unitBucket == nil {
			return fmt.Errorf("bucket %s not found", unitBucket)
		}

		value := unitB.Get([]byte(unitID))
		if value == nil {
			return nil
		}

		var hashes []string
		if err := json.Unmarshal(value, &hashes); err != nil {
			return err
		}

		txInfoB := tx.Bucket(txInfoBucket)
		if txInfoBucket == nil {
			return fmt.Errorf("bucket %s not found", txInfoBucket)
		}

		for _, hash := range hashes {
			txInfo := &exTypes.TxInfo{}
			txBytes := txInfoB.Get([]byte(hash))
			if txBytes == nil {
				continue
			}
			if err := json.Unmarshal(txBytes, txInfo); err != nil {
				return err
			}
			txs = append(txs, txInfo)
		}
		return nil
	})

	return txs, err
}
