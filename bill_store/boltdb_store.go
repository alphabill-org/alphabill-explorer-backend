package bill_store

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)

const BoltExplorerStoreFileName = "blocks.db"

var (
	blockInfoBucket          = []byte("blockInfoBucket")        // block_number => exTypes.BlockInfo
	txInfoBucket             = []byte("txInfoBucket")           // txRecHash => exTypes.TxInfo
	unitIDsToTxRecHashBucket = []byte("unitIDBucket")           // unitID => [txRecHash]
	txOrderHashToTxRecHash   = []byte("txOrderHashToTxRecHash") // txOrderHash => txRecHash mapping
	metaBucket               = []byte("metaBucket")             // block_number_key => block_number_val
)

var (
	blockNumberKey = []byte("blockNumberKey")
)

var (
	ErrOwnerPredicateIsNil = errors.New("unit owner predicate is nil")
)

type (
	boltBillStore struct {
		db *bolt.DB
	}
)

// NewBoltBillStore creates new on-disk persistent storage for bills and proofs using bolt db.
// If the file does not exist then it will be created, however, parent directories must exist beforehand.
func NewBoltBillStore(dbFile string) (*boltBillStore, error) {
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 3 * time.Second}) // -rw-------
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt DB: %w", err)
	}
	bbs := &boltBillStore{db: db}
	err = CreateBuckets(db.Update,
		blockInfoBucket,
		txInfoBucket,
		metaBucket,
		unitIDsToTxRecHashBucket,
		txOrderHashToTxRecHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create db buckets: %w", err)
	}
	err = bbs.initMetaData()
	if err != nil {
		return nil, fmt.Errorf("failed to init db metadata: %w", err)
	}
	return bbs, nil
}

func (s *boltBillStore) GetBlockNumber() (uint64, error) {
	blockNumber := uint64(0)
	err := s.db.Update(func(tx *bolt.Tx) error {
		blockNumberBytes := tx.Bucket(metaBucket).Get(blockNumberKey)
		blockNumber = util.BytesToUint64(blockNumberBytes)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func (s *boltBillStore) SetBlockNumber(blockNumber uint64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		blockNumberBytes := util.Uint64ToBytes(blockNumber)
		err := tx.Bucket(metaBucket).Put(blockNumberKey, blockNumberBytes)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *boltBillStore) initMetaData() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		val := tx.Bucket(metaBucket).Get(blockNumberKey)
		if val == nil {
			return tx.Bucket(metaBucket).Put(blockNumberKey, util.Uint64ToBytes(0))
		}
		return nil
	})
}

func setPosition(c *bolt.Cursor, key []byte) ([]byte, []byte) {
	if key != nil {
		k, v := c.Seek(key)
		if !bytes.Equal(k, key) {
			return nil, nil
		}
		return k, v
	}
	return c.First()
}
