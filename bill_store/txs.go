package bill_store

import (
	"encoding/json"

	exTypes "github.com/alphabill-org/alphabill-explorer-backend/types"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetTxInfo(txInfo *exTypes.TxInfo) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		txExplorerBytes, err := json.Marshal(txInfo)
		if err != nil {
			return err
		}
		txExplorerBucket := tx.Bucket(txInfoBucket)
		hashBytes := []byte(txInfo.Hash)
		err = txExplorerBucket.Put(hashBytes, txExplorerBytes)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *boltBillStore) GetTxInfo(txHash string) (*exTypes.TxInfo, error) {
	var txEx *exTypes.TxInfo
	hashBytes := []byte(txHash)
	err := s.db.Update(func(tx *bolt.Tx) error {
		txExplorerBytes := tx.Bucket(txInfoBucket).Get(hashBytes)
		return json.Unmarshal(txExplorerBytes, &txEx)
	})
	if err != nil {
		return nil, err
	}
	return txEx, nil
}