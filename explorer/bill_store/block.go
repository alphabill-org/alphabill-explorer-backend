package bill_store

import (
	"fmt"

	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)



func  (s boltBillStoreTx) GetLastBlockNumber() (uint64, error) {
	lastBlockNo := uint64(0)
	err := s.withTx(s.tx, func(tx *bolt.Tx) error {
		b := tx.Bucket(blockBucket)

		if b == nil {
			return fmt.Errorf("bucket %s not found", blockBucket)
		}

		c := b.Cursor()
		key, _ := c.Last()
		if key == nil {
			return fmt.Errorf("no entries in the bucket  %s", blockBucket)
		}
		lastBlockNo = util.BytesToUint64(key)
		return nil
	}, false)
	if err != nil {
		return 0, err
	}
	return lastBlockNo, nil
}
