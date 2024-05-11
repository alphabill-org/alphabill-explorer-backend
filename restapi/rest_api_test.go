package restapi

import (
	"context"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	moneyApi "github.com/alphabill-org/alphabill-wallet/wallet/money/api"
	types "github.com/alphabill-org/alphabill/types"
)

type MockExplorerBackendService struct {
	getLastBlockNumberFunc       func() (uint64, error)
	getBlockFunc                 func(blockNumber uint64) (*api.BlockInfo, error)
	getBlocksFunc                func(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)
	getTxInfoFunc                func(txHash api.TxHash) (res *api.TxInfo, err error)
	getBlockTxsByBlockNumberFunc func(blockNumber uint64) (res []*api.TxInfo, err error)
	getRoundNumberFunc           func(ctx context.Context) (uint64, error)
	getTxsByUnitID               func(unitID types.UnitID) ([]*api.TxInfo, error)
	getTxs                       func(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error)
	getBillsByPubKey             func(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
}

func (m *MockExplorerBackendService) GetLastBlockNumber() (uint64, error) {
	if m.getLastBlockNumberFunc != nil {
		return m.getLastBlockNumberFunc()
	}
	panic("GetLastBlockNumberFunc not implemented")
}

func (m *MockExplorerBackendService) GetBlock(blockNumber uint64) (*api.BlockInfo, error) {
	if m.getBlockFunc != nil {
		return m.getBlockFunc(blockNumber)
	}
	panic("GetBlockFunc not implemented")
}

func (m *MockExplorerBackendService) GetBlocks(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error) {
	if m.getBlocksFunc != nil {
		return m.getBlocksFunc(dbStartBlock, count)
	}
	panic("GetBlocksFunc not implemented")
}

func (m *MockExplorerBackendService) GetTxInfo(txHash api.TxHash) (res *api.TxInfo, err error) {
	if m.getTxInfoFunc != nil {
		return m.getTxInfoFunc(txHash)
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

func (m *MockExplorerBackendService) GetTxsByUnitID(unitID types.UnitID) ([]*api.TxInfo, error) {
	if m.getRoundNumberFunc != nil {
		return m.getTxsByUnitID(unitID)
	}
	panic("GetTxsByUnitIDFunc not implemented")
}

func (m *MockExplorerBackendService) GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error) {
	if m.getRoundNumberFunc != nil {
		return m.getBillsByPubKey(ctx, ownerID)
	}
	panic("GetBillsByPubKey not implemented")
}

func (m *MockExplorerBackendService) GetTxs(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error) {
	if m.getTxs != nil {
		return m.getTxs(startSequenceNumber, count)
	}
	panic("GetTxs not implemented")
}
