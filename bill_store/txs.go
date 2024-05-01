package bill_store

import (
	"encoding/json"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetTxInfo(txInfo *api.TxInfo) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		txInfoBytes, err := json.Marshal(txInfo)
		if err != nil {
			return err
		}
		txInfoBucket := tx.Bucket(txInfoBucket)
		hashBytes := txInfo.TxRecordHash
		err = txInfoBucket.Put(hashBytes, txInfoBytes)
		if err != nil {
			return err
		}
		for _, unitID := range txInfo.Transaction.ServerMetadata.TargetUnits {
			if err = s.addUnitTxHash(tx, unitID, txInfo.TxRecordHash); err != nil {
				return fmt.Errorf("failed to add unit tx hash: %w", err)
			}
		}
		return s.addTxHashMapping(tx, txInfo.TxOrderHash, txInfo.TxRecordHash)
	})
}

func (s *boltBillStore) addUnitTxHash(tx *bolt.Tx, unitID, txRecordHash []byte) error {
	bucket, err := EnsureSubBucket(tx, unitIDsToTxRecHashBucket, unitID, false)
	if err != nil {
		return fmt.Errorf("failed to ensure sub-bucket for unitID %s: %v", unitID, err)
	}
	err = bucket.Put(txRecordHash, nil)
	if err != nil {
		return fmt.Errorf("failed to set txRecordHash %X for unit %X: %v", txRecordHash, unitID, err)
	}
	return nil
}

func (s *boltBillStore) addTxHashMapping(tx *bolt.Tx, txOrderHash, txRecHash []byte) error {
	bucket := tx.Bucket(txOrderHashToTxRecHash)
	err := bucket.Put(txOrderHash, txRecHash)
	if err != nil {
		return fmt.Errorf("failed to put txRecHash %X for txOrderHash %X: %w", txRecHash, txOrderHash, err)
	}
	return nil
}

func (s *boltBillStore) GetTxInfo(txHash []byte) (*api.TxInfo, error) {
	var txInfo *api.TxInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		txInforBytes := tx.Bucket(txInfoBucket).Get(txHash)
		return json.Unmarshal(txInforBytes, &txInfo)
	})
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (s *boltBillStore) GetTxs(startTxRecHash api.TxRecordHash, count int) (res []*api.TxInfo, prevTxHash api.TxRecordHash, err error) {
	return res, prevTxHash, s.db.View(func(tx *bolt.Tx) error {
		return s.getTxs(tx, startTxRecHash, count, &res, &prevTxHash)
	})
}

func (s *boltBillStore) getTxs(tx *bolt.Tx, startTxRecHash api.TxRecordHash, count int, res *[]*api.TxInfo, prevTxHash *api.TxRecordHash) error {
	txInfoBucket := tx.Bucket(txInfoBucket)
	cursor := txInfoBucket.Cursor()
	if len(startTxRecHash) == 0 {
		startTxRecHash, _ = cursor.Last()
	}
	for k, v := cursor.Seek(startTxRecHash); k != nil && count > 0; k, v = cursor.Prev() {
		var txInfo api.TxInfo
		if err := json.Unmarshal(v, &txInfo); err != nil {
			return err
		}
		*res = append(*res, &txInfo)
		if count--; count == 0 {
			*prevTxHash, _ = cursor.Prev()
			break
		}
	}
	return nil
}

func (s *boltBillStore) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*api.TxInfo, err error) {
	return res, s.db.View(func(tx *bolt.Tx) error {
		var err error
		res, err = s.getBlockTxsByBlockNumber(tx, blockNumber)
		return err
	})
}

func (s *boltBillStore) getBlockTxsByBlockNumber(tx *bolt.Tx, blockNumber uint64) ([]*api.TxInfo, error) {
	var txs []*api.TxInfo
	blockNumberBytes := util.Uint64ToBytes(blockNumber)

	blockInfoBytes := tx.Bucket(blockInfoBucket).Get(blockNumberBytes)
	if blockInfoBytes == nil {
		return nil, fmt.Errorf("no block data found for block number %d", blockNumber)
	}

	var b api.BlockInfo
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

		t := &api.TxInfo{}
		if err := json.Unmarshal(txBytes, t); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}

	return txs, nil
}

func (s *boltBillStore) GetTxsByUnitID(unitID string) ([]*api.TxInfo, error) {
	var txs []*api.TxInfo

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(unitIDsToTxRecHashBucket)
		if unitIDsToTxRecHashBucket == nil {
			return fmt.Errorf("bucket %s not found", unitIDsToTxRecHashBucket)
		}

		subBucket := bucket.Bucket([]byte(unitID))
		if subBucket == nil {
			return fmt.Errorf("sub bucket %s not found", []byte(unitID))
		}

		txBucket := tx.Bucket(txInfoBucket)
		if txInfoBucket == nil {
			return fmt.Errorf("bucket %s not found", txInfoBucket)
		}

		err := subBucket.ForEach(func(txHash []byte, _ []byte) error {
			txBytes := txBucket.Get(txHash)
			if txBytes == nil {
				return fmt.Errorf("no transaction info found for txHash %s", txHash)
			}

			var txInfo *api.TxInfo
			if err := json.Unmarshal(txBytes, &txInfo); err != nil {
				return err
			}
			txs = append(txs, txInfo)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	return txs, err
}
