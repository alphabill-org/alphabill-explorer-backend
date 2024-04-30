package bill_store

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetUnitID(unitID string, txHash string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		unitIDBytes := []byte(unitID)

		bucket := tx.Bucket(unitIDsToTxRecHashBucket)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", unitIDsToTxRecHashBucket)
		}

		subBucket, err := bucket.CreateBucketIfNotExists(unitIDBytes)
		if err != nil {
			return fmt.Errorf("failed to create or find sub-bucket for unitID %s: %v", unitID, err)
		}

		if err := subBucket.Put([]byte(txHash), nil); err != nil {
			return fmt.Errorf("failed to set txHash %s with nil in sub-bucket: %v", txHash, err)
		}

		return nil
	})
}
