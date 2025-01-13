package restapi

import (
	"context"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
)

type MockExplorerBackendService struct {
	getLastBlockFunc             func(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
	getBlockFunc                 func(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
	getBlocksInRangeFunc         func(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)
	getTxInfoFunc                func(ctx context.Context, txHash api.TxHash) (res *api.TxInfo, err error)
	getBlockTxsByBlockNumberFunc func(blockNumber uint64) (res []*api.TxInfo, err error)
	getRoundNumberFunc           func(ctx context.Context) (uint64, error)
	getTxsByUnitID               func(ctx context.Context, txHash api.TxHash) (res *api.TxInfo, err error)
	getTxs                       func(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error)
	//getBillsByPubKey             func(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
}

func (m *MockExplorerBackendService) GetTxsPage(ctx context.Context, partitionID types.PartitionID, startID string, limit int) (transactions []*api.TxInfo, previousID string, err error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error) {
	if m.getBlockFunc != nil {
		return m.getBlockFunc(ctx, blockNumber, partitionIDs)
	}
	panic("implement me")
}

func (m *MockExplorerBackendService) GetBlocksInRange(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error) {
	if m.getBlocksInRangeFunc != nil {
		return m.getBlocksInRangeFunc(ctx, partitionID, dbStartBlock, count)
	}
	panic("implement me")
}

func (m *MockExplorerBackendService) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*api.TxInfo, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*api.TxInfo, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error) {
	if m.getLastBlockFunc != nil {
		return m.getLastBlockFunc(ctx, partitionIDs)
	}
	panic("getLastBlockFunc not implemented")
}

func (m *MockExplorerBackendService) GetTxInfo(ctx context.Context, txHash api.TxHash) (res *api.TxInfo, err error) {
	if m.getTxInfoFunc != nil {
		return m.getTxInfoFunc(ctx, txHash)
	}
	panic("GetTxInfoFunc not implemented")
}

func (m *MockExplorerBackendService) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*api.TxInfo, err error) {
	if m.getBlockTxsByBlockNumberFunc != nil {
		return m.getBlockTxsByBlockNumberFunc(blockNumber)
	}
	panic("GetBlockTxsByBlockNumberFunc not implemented")
}

func (m *MockExplorerBackendService) GetRoundNumber(ctx context.Context) (uint64, error) {
	if m.getRoundNumberFunc != nil {
		return m.getRoundNumberFunc(ctx)
	}
	panic("GetRoundNumberFunc not implemented")
}

/*func (m *MockExplorerBackendService) GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error) {
	if m.getRoundNumberFunc != nil {
		return m.getBillsByPubKey(ctx, ownerID)
	}
	panic("GetBillsByPubKey not implemented")
}*/

func (m *MockExplorerBackendService) GetTxs(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error) {
	if m.getTxs != nil {
		return m.getTxs(startSequenceNumber, count)
	}
	panic("GetTxs not implemented")
}
