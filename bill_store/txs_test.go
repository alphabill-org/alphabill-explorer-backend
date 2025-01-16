package bill_store

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestBoltBillStore_GetTxs(t *testing.T) {
	st, err := NewBoltBillStore(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)

	t.Run("empty", func(t *testing.T) {
		res, prev, err := st.GetTxs(0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 0, prev)
		require.Nil(t, res)
	})

	max := 10
	err = st.db.Update(func(tx *bolt.Tx) error {
		for i := 0; i < max; i++ {
			txInfoBucket := tx.Bucket(txInfoBucket)
			hash := []byte{byte(i)}
			txInfo := &domain.TxInfo{TxRecordHash: hash}
			bytes, err := json.Marshal(txInfo)
			require.NoError(t, err)
			require.NoError(t, txInfoBucket.Put(hash, bytes))
			require.NoError(t, st.addTxInOrder(tx, hash))
		}
		return nil
	})
	require.NoError(t, err)
	t.Run("check order from middle", func(t *testing.T) {
		res, prev, err := st.GetTxs(5, 20)
		require.NoError(t, err)
		require.EqualValues(t, 0, prev)
		require.Len(t, res, 5)
		for i, tx := range res {
			require.EqualValues(t, byte(4-i), tx.TxRecordHash[0])
		}
	})

	t.Run("check order with start nr not given", func(t *testing.T) {
		res, prev, err := st.GetTxs(0, 20)
		require.NoError(t, err)
		require.EqualValues(t, 0, prev)
		require.Len(t, res, 10)
		for i, tx := range res {
			require.EqualValues(t, byte(9-i), tx.TxRecordHash[0])
		}
	})

	t.Run("get 5 with start nr not given", func(t *testing.T) {
		res, prev, err := st.GetTxs(0, 5)
		require.NoError(t, err)
		require.Len(t, res, 5)
		for i, tx := range res {
			require.EqualValues(t, byte(9-i), tx.TxRecordHash[0])
		}
		require.EqualValues(t, 5, prev)
	})

	t.Run("check order from 7 to 5", func(t *testing.T) {
		res, prev, err := st.GetTxs(7, 3)
		require.NoError(t, err)
		require.Len(t, res, 3)
		for i, tx := range res {
			require.EqualValues(t, byte(6-i), tx.TxRecordHash[0])
		}
		require.EqualValues(t, 4, prev)
	})
}
