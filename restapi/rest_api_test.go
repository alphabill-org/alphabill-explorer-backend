package restapi

import (
	"context"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/service"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
	wallettypes "github.com/alphabill-org/alphabill-wallet/client/types"
)

type MockExplorerBackendService struct {
	getLastBlockFunc             func(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
	getLastBlocksFunc            func(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error)
	getBlockFunc                 func(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
	getBlocksInRangeFunc         func(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool) (res []*domain.BlockInfo, prevBlockNumber uint64, err error)
	getTxInfoFunc                func(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error)
	getBlockTxsByBlockNumberFunc func(blockNumber uint64) (res []*domain.TxInfo, err error)
	getRoundNumberFunc           func(ctx context.Context) ([]service.PartitionRoundInfo, error)
	getTxsByUnitID               func(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error)
	getTxsPageFunc               func(ctx context.Context, partitionID types.PartitionID, startID string, limit int) (transactions []*domain.TxInfo, previousID string, err error)
	//getBillsByPubKey             func(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
}

func (m *MockExplorerBackendService) FindTxs(ctx context.Context, searchKey []byte) ([]*domain.TxInfo, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetBillsByPubKey(ctx context.Context, ownerID hex.Bytes) (res []*wallettypes.Bill, err error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetTxsPage(ctx context.Context, partitionID types.PartitionID, startID string, limit int) (transactions []*domain.TxInfo, previousID string, err error) {
	if m.getTxsPageFunc != nil {
		return m.getTxsPageFunc(ctx, partitionID, startID, limit)
	}
	panic("implement me")
}

func (m *MockExplorerBackendService) GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
	if m.getBlockFunc != nil {
		return m.getBlockFunc(ctx, blockNumber, partitionIDs)
	}
	panic("implement me")
}

func (m *MockExplorerBackendService) GetBlocksInRange(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool) (res []*domain.BlockInfo, prevBlockNumber uint64, err error) {
	if m.getBlocksInRangeFunc != nil {
		return m.getBlocksInRangeFunc(ctx, partitionID, dbStartBlock, count, includeEmpty)
	}
	panic("implement me")
}

func (m *MockExplorerBackendService) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error) {
	panic("implement me")
}

func (m *MockExplorerBackendService) GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
	if m.getLastBlockFunc != nil {
		return m.getLastBlockFunc(ctx, partitionIDs)
	}
	panic("getLastBlockFunc not implemented")
}

func (m *MockExplorerBackendService) GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error) {
	if m.getLastBlocksFunc != nil {
		return m.getLastBlocksFunc(ctx, partitionIDs, count, includeEmpty)
	}
	panic("GetLastBlocks not implemented")
}

func (m *MockExplorerBackendService) GetTxByHash(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error) {
	if m.getTxInfoFunc != nil {
		return m.getTxInfoFunc(ctx, txHash)
	}
	panic("GetTxInfoFunc not implemented")
}

func (m *MockExplorerBackendService) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*domain.TxInfo, err error) {
	if m.getBlockTxsByBlockNumberFunc != nil {
		return m.getBlockTxsByBlockNumberFunc(blockNumber)
	}
	panic("GetBlockTxsByBlockNumberFunc not implemented")
}

func (m *MockExplorerBackendService) GetRoundNumber(ctx context.Context) ([]service.PartitionRoundInfo, error) {
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
