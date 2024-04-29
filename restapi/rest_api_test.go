package restapi

import (
	"context"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	exTypes "github.com/alphabill-org/alphabill-explorer-backend/types"
)

type MockExplorerBackendService struct {
	getLastBlockNumberFunc       func() (uint64, error)
	getBlockFunc                 func(blockNumber uint64) (*api.BlockInfo, error)
	getBlocksFunc                func(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)
	getTxInfoFunc                func(txHash string) (res *exTypes.TxInfo, err error)
	getBlockTxsByBlockNumberFunc func(blockNumber uint64) (res []*exTypes.TxInfo, err error)
	getRoundNumberFunc           func(ctx context.Context) (uint64, error)
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

func (m *MockExplorerBackendService) GetTxInfo(txHash string) (res *exTypes.TxInfo, err error) {
	if m.getTxInfoFunc != nil {
		return m.getTxInfoFunc(txHash)
	}
	panic("GetTxInfoFunc not implemented")
}

func (m *MockExplorerBackendService) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*exTypes.TxInfo, err error) {
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
