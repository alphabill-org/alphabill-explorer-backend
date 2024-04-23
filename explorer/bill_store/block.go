package bill_store

import (
	"encoding/json"
	"fmt"

	"github.com/alphabill-org/alphabill/types"
	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)

func (s boltBillStoreTx) GetLastBlockNumber() (uint64, error) {
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
func (s *boltBillStoreTx) GetBlockByBlockNumber(blockNumber uint64) (*types.Block, error) {
	var b *types.Block
	blockNumberBytes := util.Uint64ToBytes(blockNumber)
	err := s.withTx(s.tx, func(tx *bolt.Tx) error {
		blockBytes := tx.Bucket(blockBucket).Get(blockNumberBytes)
		return json.Unmarshal(blockBytes, &b)
	}, false)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *boltBillStoreTx) SetBlock(b *types.Block) error {
	return s.withTx(s.tx, func(tx *bolt.Tx) error {
		blockNumber := b.UnicityCertificate.InputRecord.RoundNumber
		blockNumberBytes := util.Uint64ToBytes(blockNumber)
		blockBytes, err := json.Marshal(b)
		if err != nil {
			return err
		}
		err = tx.Bucket(blockBucket).Put(blockNumberBytes, blockBytes)
		if err != nil {
			return err
		}
		return nil
	}, true)
}

func (s *boltBillStoreTx) GetBlocks(dbStartBlock uint64, count int) (res []*types.Block, prevBlockNumber uint64, err error) {
	return res, prevBlockNumber, s.withTx(s.tx, func(tx *bolt.Tx) error {
		var err error
		res, prevBlockNumber, err = s.getBlocks(tx, dbStartBlock, count)
		return err
	}, false)
}
func (s *boltBillStoreTx) getBlocks(tx *bolt.Tx, dbStartBlock uint64, count int) ([]*types.Block, uint64, error) {
	pb := tx.Bucket(blockBucket)

	if pb == nil {
		return nil, 0, fmt.Errorf("bucket %s not found", blockBucket)
	}

	dbStartKeyBytes := util.Uint64ToBytes(dbStartBlock)
	c := pb.Cursor()

	var res []*types.Block
	var prevBlockNumberBytes []byte
	var prevBlockNumber uint64

	for k, v := c.Seek(dbStartKeyBytes); k != nil && count > 0; k, v = c.Prev() {
		rec := &types.Block{}
		if err := json.Unmarshal(v, rec); err != nil {
			return nil, 0, fmt.Errorf("failed to deserialize tx history record: %w", err)
		}
		res = append(res, rec)
		if count--; count == 0 {
			prevBlockNumberBytes, _ = c.Prev()
			break
		}
	}
	if len(prevBlockNumberBytes) != 0 {
		prevBlockNumber = util.BytesToUint64(prevBlockNumberBytes)
	} else {
		prevBlockNumber = 0
	}
	return res, prevBlockNumber, nil
}
