//go:build manual

package mongodb

import (
	"context"
	"fmt"
	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	// todo use testContainer
	connectionString = "mongodb://localhost:27017"
	blockCount       = 5
	txsPerBlock      = 3
	partition1       = types.PartitionID(1)
	partition2       = types.PartitionID(2)
)

func TestMongoBillStore_GetBlockNumber_ZeroIfMissing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	require.NoError(t, store.ResetCollections(ctx))
	nonExistentPartition := types.PartitionID(100)
	blockNumber, err := store.GetBlockNumber(ctx, nonExistentPartition)
	require.NoError(t, err)
	require.EqualValues(t, 0, blockNumber)
}

func TestMongoBillStore_GetBlock_MultiplePartitions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)

	blockMap, err := store.GetBlock(ctx, 2, []types.PartitionID{partition1, partition2})
	require.NoError(t, err)
	require.Len(t, blockMap, 2)
	require.EqualValues(t, partition1, blockMap[partition1].PartitionID)
	require.EqualValues(t, 2, blockMap[partition1].BlockNumber)
	require.EqualValues(t, partition2, blockMap[partition2].PartitionID)
	require.EqualValues(t, 2, blockMap[partition2].BlockNumber)
}

func TestMongoBillStore_GetBlock_NoPartitionsSpecified(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)

	blockMap, err := store.GetBlock(ctx, 2, nil)
	require.NoError(t, err)
	require.Len(t, blockMap, 2)
	require.EqualValues(t, partition1, blockMap[partition1].PartitionID)
	require.EqualValues(t, 2, blockMap[partition1].BlockNumber)
	require.EqualValues(t, partition2, blockMap[partition2].PartitionID)
	require.EqualValues(t, 2, blockMap[partition2].BlockNumber)
}

func TestMongoBillStore_GetBlockRange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	blocks, prevBlockNumber, err := store.GetBlocksInRange(ctx, partition1, 4, 2)
	require.NoError(t, err)
	require.Len(t, blocks, 2)
	require.EqualValues(t, 2, prevBlockNumber)
	require.EqualValues(t, partition1, blocks[0].PartitionID)
	require.EqualValues(t, 4, blocks[0].BlockNumber)
	require.EqualValues(t, 3, blocks[1].BlockNumber)
}

func TestMongoBillStore_GetTxByHash(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	txHash := testTxHash(partition1, 3, 2)
	txInfo, err := store.GetTxInfo(ctx, txHash)
	require.NoError(t, err)
	require.EqualValues(t, partition1, txInfo.PartitionID)
	require.EqualValues(t, 3, txInfo.BlockNumber)
	require.EqualValues(t, txInfo.TxRecordHash, txHash)
}

func TestMongoBillStore_GetTxsByBlockNumber(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	txList, err := store.GetTxsByBlockNumber(ctx, 3, partition1)
	require.NoError(t, err)
	require.Len(t, txList, txsPerBlock)
	require.EqualValues(t, partition1, txList[0].PartitionID)
	require.EqualValues(t, 3, txList[0].BlockNumber)
}

func TestMongoBillStore_GetTxsByUnitID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	unit3 := []byte("unit3")
	unit4 := []byte("unit4")
	unit5 := []byte("unit5")

	txList, err := store.GetTxsByUnitID(ctx, unit4)
	require.NoError(t, err)
	require.Len(t, txList, 0)

	txInfoUnits34 := testTxInfo(partition1, testTxHash(partition1, 3, 1), 3, []types.UnitID{unit3, unit4})
	txInfoUnits45 := testTxInfo(partition1, testTxHash(partition1, 5, 2), 5, []types.UnitID{unit4, unit5})

	require.NoError(t, store.SetTxInfo(ctx, &txInfoUnits34))
	require.NoError(t, store.SetTxInfo(ctx, &txInfoUnits45))

	txList, err = store.GetTxsByUnitID(ctx, unit3)
	require.NoError(t, err)
	require.Len(t, txList, 1)
	require.EqualValues(t, txList[0].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit3, unit4})

	txList, err = store.GetTxsByUnitID(ctx, unit4)
	require.NoError(t, err)
	require.Len(t, txList, 2)
	require.EqualValues(t, txList[0].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit3, unit4})
	require.EqualValues(t, txList[1].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit4, unit5})
}

func TestMongoBillStore_GetTxsPage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	txList, previousID, err := store.GetTxsPage(ctx, partition1, "", 5)
	require.NoError(t, err)
	require.Len(t, txList, 5)
	require.EqualValues(t, 5, txList[0].BlockNumber)
	require.EqualValues(t, 5, txList[1].BlockNumber)
	require.EqualValues(t, 5, txList[2].BlockNumber)
	require.EqualValues(t, 4, txList[3].BlockNumber)
	require.EqualValues(t, 4, txList[4].BlockNumber)

	txList, previousID, err = store.GetTxsPage(ctx, partition1, previousID, 2)
	require.NoError(t, err)
	require.Len(t, txList, 2)
	require.EqualValues(t, 4, txList[0].BlockNumber)
	require.EqualValues(t, 3, txList[1].BlockNumber)

	txList, previousID, err = store.GetTxsPage(ctx, partition1, previousID, 200)
	require.NoError(t, err)
	require.Len(t, txList, 8)
	require.EqualValues(t, 3, txList[0].BlockNumber)
	require.EqualValues(t, 1, txList[len(txList)-1].BlockNumber)
}

func TestMongoBillStore_GetLastBlocks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	store, err := NewMongoBillStore(ctx, connectionString)
	require.NoError(t, err)

	initTestDB(t, ctx, store)
	// add new block to partition 2
	block := &api.BlockInfo{PartitionID: partition2, BlockNumber: 6}
	err = store.SetBlockInfo(ctx, block)

	lastBlocks, err := store.GetLastBlock(ctx, []types.PartitionID{partition1, partition2})
	fmt.Printf("%v \n", lastBlocks[partition1].BlockNumber)
	fmt.Printf("%v \n", lastBlocks[partition2].BlockNumber)
	require.NoError(t, err)
	require.NotNil(t, lastBlocks)
	require.Len(t, lastBlocks, 2)
	require.NotNil(t, lastBlocks[partition1])
	require.EqualValues(t, 5, lastBlocks[partition1].BlockNumber)
	require.NotNil(t, lastBlocks[partition2])
	require.EqualValues(t, 6, lastBlocks[partition2].BlockNumber)
}

func initTestDB(t *testing.T, ctx context.Context, store *MongoBillStore) {
	err := store.ResetCollections(ctx)
	require.NoError(t, err)

	for partition := 1; partition < 3; partition++ {
		for i := 1; i < blockCount+1; i++ {
			txHashes := make([]api.TxHash, 0, txsPerBlock)
			for j := 1; j < txsPerBlock+1; j++ {
				txHashes = append(txHashes, testTxHash(types.PartitionID(partition), i, j))
			}

			for _, txHash := range txHashes {
				txInfo := &api.TxInfo{
					TxRecordHash: txHash,
					BlockNumber:  uint64(i),
					Transaction: &types.TransactionRecord{
						ServerMetadata: &types.ServerMetadata{
							TargetUnits: []types.UnitID{[]byte("unit1"), []byte("unit2")},
						},
					},
					PartitionID: types.PartitionID(partition),
				}
				err = store.SetTxInfo(ctx, txInfo)
				require.NoError(t, err)
			}

			block := &api.BlockInfo{
				Header:      nil,
				TxHashes:    txHashes,
				PartitionID: types.PartitionID(partition),
				BlockNumber: uint64(i),
			}
			err = store.SetBlockInfo(ctx, block)
			require.NoError(t, err)
		}
	}
}

func testTxInfo(partitionID types.PartitionID, txHash api.TxHash, blockNr uint64, targetUnits []types.UnitID) api.TxInfo {
	return api.TxInfo{
		TxRecordHash: txHash,
		BlockNumber:  blockNr,
		Transaction: &types.TransactionRecord{
			ServerMetadata: &types.ServerMetadata{
				TargetUnits: targetUnits,
			},
		},
		PartitionID: partitionID,
	}
}

func testTxHash(partitionID types.PartitionID, blockNumber, txNumber int) []byte {
	return []byte(fmt.Sprintf("p%db%dtx%d", partitionID, blockNumber, txNumber))
}
