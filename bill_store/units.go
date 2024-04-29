package bill_store

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetUnitID(unitID string, txHash string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		unitIDBytes := []byte(unitID)
		txHashBytes := []byte(txHash)

		bucket, err := tx.CreateBucketIfNotExists(unitIDBytes)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		existing := bucket.Get(unitIDBytes)
		if existing == nil {
			if err := bucket.Put(unitIDBytes, txHashBytes); err != nil {
				return fmt.Errorf("put value: %s", err)
			}
		} else {
			newValue := append(existing, txHashBytes...)
			if err := bucket.Put(unitIDBytes, newValue); err != nil {
				return fmt.Errorf("update value: %s", err)
			}
		}
		return nil
	})
}