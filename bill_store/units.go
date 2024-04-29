package bill_store

import (
	"fmt"
	abtypes "github.com/alphabill-org/alphabill/types"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetUnit(unitID abtypes.UnitID, txHash []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists(unitID)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		existing := bucket.Get(unitID)
		if existing == nil {
			if err := bucket.Put(unitID, txHash); err != nil {
				return fmt.Errorf("put value: %s", err)
			}
		} else {
			newValue := append(existing, txHash...)
			if err := bucket.Put(unitID, newValue); err != nil {
				return fmt.Errorf("update value: %s", err)
			}
		}
		return nil
	})
}