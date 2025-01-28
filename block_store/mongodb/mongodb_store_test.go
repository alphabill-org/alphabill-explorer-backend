//go:build manual

package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	mongocontainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MongoBillStoreSuite struct {
	suite.Suite
	container *mongocontainer.MongoDBContainer
	store     *MongoBlockStore
	ctx       context.Context
	cancel    context.CancelFunc
}

const (
	mongoDBImage = "mongo:7.0"
	blockCount   = 5
	txsPerBlock  = 3
	partition1   = types.PartitionID(1)
	partition2   = types.PartitionID(2)
)

func TestMongoBillStoreSuite(t *testing.T) {
	suite.Run(t, new(MongoBillStoreSuite))
}

func (suite *MongoBillStoreSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	mongoContainer, err := mongocontainer.Run(suite.ctx, mongoDBImage, testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")))
	suite.Require().NoError(err, "failed to start MongoDB container")
	suite.container = mongoContainer

	connectionString, err := mongoContainer.ConnectionString(suite.ctx)
	suite.Require().NoError(err)

	suite.store, err = NewMongoBlockStore(suite.ctx, connectionString)
	suite.Require().NoError(err, "failed to initialize MongoBlockStore")
}

func (suite *MongoBillStoreSuite) TearDownSuite() {
	suite.Require().NoError(suite.container.Stop(suite.ctx, nil), "failed to stop MongoDB container")
	suite.cancel()
}

func (suite *MongoBillStoreSuite) SetupTest() {
	initTestDB(suite.T(), suite.ctx, suite.store)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlockNumber_ReturnForAllPartitions() {
	err := suite.store.SetBlockNumber(suite.ctx, partition1, 1)
	require.NoError(suite.T(), err)
	err = suite.store.SetBlockNumber(suite.ctx, partition2, 2)
	require.NoError(suite.T(), err)

	blockNumbers, err := suite.store.GetBlockNumbers(suite.ctx, nil)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blockNumbers)
	require.EqualValues(suite.T(), 1, blockNumbers[partition1])
	require.EqualValues(suite.T(), 2, blockNumbers[partition2])
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlockNumber_ReturnForGivenPartition() {
	err := suite.store.SetBlockNumber(suite.ctx, partition2, 2)
	require.NoError(suite.T(), err)

	blockNumbers, err := suite.store.GetBlockNumbers(suite.ctx, []types.PartitionID{partition2})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blockNumbers)
	require.EqualValues(suite.T(), 2, blockNumbers[partition2])
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlockNumber_ZeroIfMissing() {
	nonExistentPartition := types.PartitionID(100)
	blockNumbers, err := suite.store.GetBlockNumbers(suite.ctx, []types.PartitionID{nonExistentPartition})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blockNumbers)
	require.EqualValues(suite.T(), 0, blockNumbers[nonExistentPartition])
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlock_MultiplePartitions() {
	blockMap, err := suite.store.GetBlock(suite.ctx, 2, []types.PartitionID{partition1, partition2})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), blockMap, 2)
	require.EqualValues(suite.T(), partition1, blockMap[partition1].PartitionID)
	require.EqualValues(suite.T(), 2, blockMap[partition1].BlockNumber)
	require.EqualValues(suite.T(), partition2, blockMap[partition2].PartitionID)
	require.EqualValues(suite.T(), 2, blockMap[partition2].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlock_NoPartitionsSpecified() {
	blockMap, err := suite.store.GetBlock(suite.ctx, 2, nil)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), blockMap, 2)
	require.EqualValues(suite.T(), partition1, blockMap[partition1].PartitionID)
	require.EqualValues(suite.T(), 2, blockMap[partition1].BlockNumber)
	require.EqualValues(suite.T(), partition2, blockMap[partition2].PartitionID)
	require.EqualValues(suite.T(), 2, blockMap[partition2].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetBlockRange() {
	blocks, prevBlockNumber, err := suite.store.GetBlocksInRange(suite.ctx, partition1, 4, 2, true)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), blocks, 2)
	require.EqualValues(suite.T(), 2, prevBlockNumber)
	require.EqualValues(suite.T(), partition1, blocks[0].PartitionID)
	require.EqualValues(suite.T(), 4, blocks[0].BlockNumber)
	require.EqualValues(suite.T(), 3, blocks[1].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetLastBlocks() {
	err := suite.store.SetBlockInfo(suite.ctx, &domain.BlockInfo{
		TxHashes:    []domain.TxHash{},
		PartitionID: partition1,
		BlockNumber: blockCount + 1,
	})

	blocks, err := suite.store.GetLastBlocks(suite.ctx, []types.PartitionID{partition1, partition2}, 4, true)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blocks)
	require.Len(suite.T(), blocks[partition1], 4)
	require.EqualValues(suite.T(), blockCount+1, blocks[partition1][0].BlockNumber)
	require.Len(suite.T(), blocks[partition2], 4)
	require.EqualValues(suite.T(), blockCount, blocks[partition2][0].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetLastBlocks_NoPartitionsSpecified() {
	err := suite.store.SetBlockInfo(suite.ctx, &domain.BlockInfo{
		TxHashes:    []domain.TxHash{},
		PartitionID: partition1,
		BlockNumber: blockCount + 1,
	})

	require.NoError(suite.T(), suite.store.SetBlockNumber(suite.ctx, partition1, blockCount+1))
	require.NoError(suite.T(), suite.store.SetBlockNumber(suite.ctx, partition2, blockCount))

	blocks, err := suite.store.GetLastBlocks(suite.ctx, []types.PartitionID{}, 4, true)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blocks)
	require.Len(suite.T(), blocks[partition1], 4)
	require.EqualValues(suite.T(), blockCount+1, blocks[partition1][0].BlockNumber)
	require.Len(suite.T(), blocks[partition2], 4)
	require.EqualValues(suite.T(), blockCount, blocks[partition2][0].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetLastBlocks_NoEmptyBlocks() {
	// add empty blocks
	err := suite.store.SetBlockInfo(suite.ctx, &domain.BlockInfo{
		TxHashes:    []domain.TxHash{},
		PartitionID: partition1,
		BlockNumber: blockCount + 1,
	})
	require.NoError(suite.T(), err)
	err = suite.store.SetBlockInfo(suite.ctx, &domain.BlockInfo{
		TxHashes:    nil,
		PartitionID: partition1,
		BlockNumber: blockCount + 2,
	})
	require.NoError(suite.T(), err)

	blocks, err := suite.store.GetLastBlocks(suite.ctx, []types.PartitionID{partition1}, 4, false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), blocks)
	require.Len(suite.T(), blocks[partition1], 4)
	require.EqualValues(suite.T(), blockCount, blocks[partition1][0].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetTxByTxRecordHash() {
	txHash := testTxRecordHash(partition1, 3, 2)
	txInfo, err := suite.store.GetTxInfo(suite.ctx, txHash)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), partition1, txInfo.PartitionID)
	require.EqualValues(suite.T(), 3, txInfo.BlockNumber)
	require.EqualValues(suite.T(), txInfo.TxRecordHash, txHash)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetTxByTxOrderHash() {
	txHash := testTxOrderHash(partition1, 3, 2)
	txInfo, err := suite.store.GetTxInfo(suite.ctx, txHash)
	require.NoError(suite.T(), err)
	require.EqualValues(suite.T(), partition1, txInfo.PartitionID)
	require.EqualValues(suite.T(), 3, txInfo.BlockNumber)
	require.EqualValues(suite.T(), txInfo.TxOrderHash, txHash)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetTxsByBlockNumber() {
	txList, err := suite.store.GetTxsByBlockNumber(suite.ctx, 3, partition1)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, txsPerBlock)
	require.EqualValues(suite.T(), partition1, txList[0].PartitionID)
	require.EqualValues(suite.T(), 3, txList[0].BlockNumber)
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetTxsByUnitID() {
	unit3 := []byte("unit3")
	unit4 := []byte("unit4")
	unit5 := []byte("unit5")

	txList, err := suite.store.GetTxsByUnitID(suite.ctx, unit4)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 0)

	txInfoUnits34 := testTxInfo(partition1, testTxRecordHash(partition1, 3, 1), 3, []types.UnitID{unit3, unit4})
	txInfoUnits45 := testTxInfo(partition1, testTxRecordHash(partition1, 5, 2), 5, []types.UnitID{unit4, unit5})

	require.NoError(suite.T(), suite.store.SetTxInfo(suite.ctx, &txInfoUnits34))
	require.NoError(suite.T(), suite.store.SetTxInfo(suite.ctx, &txInfoUnits45))

	txList, err = suite.store.GetTxsByUnitID(suite.ctx, unit3)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 1)
	require.EqualValues(suite.T(), txList[0].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit3, unit4})

	txList, err = suite.store.GetTxsByUnitID(suite.ctx, unit4)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 2)
	require.EqualValues(suite.T(), txList[0].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit3, unit4})
	require.EqualValues(suite.T(), txList[1].Transaction.ServerMetadata.TargetUnits, []types.UnitID{unit4, unit5})
}

func (suite *MongoBillStoreSuite) TestMongoBillStore_GetTxsPage() {
	txList, previousID, err := suite.store.GetTxsPage(suite.ctx, partition1, "", 5)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 5)
	require.EqualValues(suite.T(), 5, txList[0].BlockNumber)
	require.EqualValues(suite.T(), 5, txList[1].BlockNumber)
	require.EqualValues(suite.T(), 5, txList[2].BlockNumber)
	require.EqualValues(suite.T(), 4, txList[3].BlockNumber)
	require.EqualValues(suite.T(), 4, txList[4].BlockNumber)

	txList, previousID, err = suite.store.GetTxsPage(suite.ctx, partition1, previousID, 2)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 2)
	require.EqualValues(suite.T(), 4, txList[0].BlockNumber)
	require.EqualValues(suite.T(), 3, txList[1].BlockNumber)

	txList, previousID, err = suite.store.GetTxsPage(suite.ctx, partition1, previousID, 200)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), txList, 8)
	require.EqualValues(suite.T(), 3, txList[0].BlockNumber)
	require.EqualValues(suite.T(), 1, txList[len(txList)-1].BlockNumber)
}

func initTestDB(t *testing.T, ctx context.Context, store *MongoBlockStore) {
	err := store.ResetCollections(ctx)
	require.NoError(t, err)

	for partition := 1; partition < 3; partition++ {
		for i := 1; i < blockCount+1; i++ {
			txRecordHashes := make([]domain.TxHash, 0, txsPerBlock)
			for j := 1; j < txsPerBlock+1; j++ {
				txRecordHashes = append(txRecordHashes, testTxRecordHash(types.PartitionID(partition), i, j))
			}
			txOrderHashes := make([]domain.TxHash, 0, txsPerBlock)
			for j := 1; j < txsPerBlock+1; j++ {
				txOrderHashes = append(txOrderHashes, testTxOrderHash(types.PartitionID(partition), i, j))
			}

			for k, txHash := range txRecordHashes {
				txInfo := &domain.TxInfo{
					TxRecordHash: txHash,
					TxOrderHash:  txOrderHashes[k],
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

			block := &domain.BlockInfo{
				Header:      nil,
				TxHashes:    txRecordHashes,
				PartitionID: types.PartitionID(partition),
				BlockNumber: uint64(i),
			}
			err = store.SetBlockInfo(ctx, block)
			require.NoError(t, err)
		}
	}
}

func testTxInfo(partitionID types.PartitionID, txHash domain.TxHash, blockNr uint64, targetUnits []types.UnitID) domain.TxInfo {
	return domain.TxInfo{
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

func testTxRecordHash(partitionID types.PartitionID, blockNumber, txNumber int) []byte {
	return []byte(fmt.Sprintf("p%db%dtx%d", partitionID, blockNumber, txNumber))
}

func testTxOrderHash(partitionID types.PartitionID, blockNumber, txNumber int) []byte {
	return []byte(fmt.Sprintf("p%db%dtx%d_order", partitionID, blockNumber, txNumber))
}
