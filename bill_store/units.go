package bill_store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetUnitID(unitID string, txHash string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		unitIDBytes := []byte(unitID)

		bucket:= tx.Bucket(unitBucket)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", unitBucket)
		}

        oldValue := bucket.Get(unitIDBytes)
        var hashes []string
        
        if oldValue != nil {
            err := json.Unmarshal(oldValue, &hashes)
            if err != nil {
                return err
            }
        }
        
        hashes = append(hashes, txHash)
        
        newValue, err := json.Marshal(hashes)
        if err != nil {
            return err
        }

        return bucket.Put(unitIDBytes, newValue)
	})
}

