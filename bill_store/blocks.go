package bill_store

import (
	"encoding/json"
	"fmt"

	st "github.com/alphabill-org/alphabill-explorer-backend/types"
	"github.com/alphabill-org/alphabill/types"
	"github.com/alphabill-org/alphabill/util"
	bolt "go.etcd.io/bbolt"
)

func (s *boltBillStore) SetBlockInfo(b *types.Block) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		blockInfoBucket := tx.Bucket(blockInfoBucket)
		blockNumber := b.UnicityCertificate.InputRecord.RoundNumber
		blockNumberBytes := util.Uint64ToBytes(blockNumber)

		blockInfo, err := st.NewBlockInfo(b)
		if err != nil {
			return err
		}

		blockInfoBytes, err := json.Marshal(blockInfo)

		if err != nil {
			return err
		}

		err = blockInfoBucket.Put(blockNumberBytes, blockInfoBytes)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *boltBillStore) GetLastBlockNumber() (uint64, error) {
	lastBlockNo := uint64(0)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blockInfoBucket)

		if b == nil {
			return fmt.Errorf("bucket %s not found", blockInfoBucket)
		}

		c := b.Cursor()
		key, _ := c.Last()
		if key == nil {
			return fmt.Errorf("no entries in the bucket  %s", blockInfoBucket)
		}
		lastBlockNo = util.BytesToUint64(key)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lastBlockNo, nil
}

func (s *boltBillStore) GetBlockInfo(blockNumber uint64) (*st.BlockInfo, error) {
	var b *st.BlockInfo
	blockNumberBytes := util.Uint64ToBytes(blockNumber)
	err := s.db.View(func(tx *bolt.Tx) error {
		blockInfoBytes := tx.Bucket(blockInfoBucket).Get(blockNumberBytes)
		return json.Unmarshal(blockInfoBytes, &b)
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *boltBillStore) GetBlocksInfo(dbStartBlock uint64, count int) (res []*st.BlockInfo, prevBlockNumber uint64, err error) {
	return res, prevBlockNumber, s.db.View(func(tx *bolt.Tx) error {
		var err error
		res, prevBlockNumber, err = s.getBlocksInfo(tx, dbStartBlock, count)
		return err
	})
}

func (s *boltBillStore) getBlocksInfo(tx *bolt.Tx, dbStartBlock uint64, count int) ([]*st.BlockInfo, uint64, error) {
	pb := tx.Bucket(blockInfoBucket)

	if pb == nil {
		return nil, 0, fmt.Errorf("bucket %s not found", blockInfoBucket)
	}

	dbStartKeyBytes := util.Uint64ToBytes(dbStartBlock)
	c := pb.Cursor()

	var res []*st.BlockInfo
	var prevBlockNumberBytes []byte
	var prevBlockNumber uint64

	for k, v := c.Seek(dbStartKeyBytes); k != nil && count > 0; k, v = c.Prev() {
		rec := &st.BlockInfo{}
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
