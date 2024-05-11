package service

import (
	"context"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	moneyApi "github.com/alphabill-org/alphabill-wallet/wallet/money/api"
	txSysMoney "github.com/alphabill-org/alphabill/txsystem/money"
	"github.com/alphabill-org/alphabill/types"
)

type (
	BillStore interface {

		//block
		GetLastBlockNumber() (uint64, error)
		GetBlockInfo(blockNumber uint64) (*api.BlockInfo, error)
		GetBlocksInfo(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(txHash api.TxHash) (*api.TxInfo, error)
		GetBlockTxsByBlockNumber(blockNumber uint64) (res []*api.TxInfo, err error)
		GetTxsByUnitID(unitID types.UnitID) ([]*api.TxInfo, error)
		GetTxs(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error)
	}

	ABClient interface {
		GetRoundNumber(ctx context.Context) (uint64, error)
		GetUnitsByOwnerID(ctx context.Context, ownerID types.Bytes) ([]types.UnitID, error)
		GetBill(ctx context.Context, unitID types.UnitID, includeStateProof bool) (*moneyApi.Bill, error)
	}

	ExplorerBackend struct {
		store  BillStore
		client ABClient
	}
)

func NewExplorerBackend(store BillStore, client ABClient) *ExplorerBackend {
	return &ExplorerBackend{
		store:  store,
		client: client,
	}
}

// GetRoundNumber returns latest round number.
func (ex *ExplorerBackend) GetRoundNumber(ctx context.Context) (uint64, error) {
	return ex.client.GetRoundNumber(ctx)
}

// block
// GetLastBlockNumber returns last processed block
func (ex *ExplorerBackend) GetLastBlockNumber() (uint64, error) {
	return ex.store.GetLastBlockNumber()
}

// GetBlock returns block with given block number.
func (ex *ExplorerBackend) GetBlock(blockNumber uint64) (*api.BlockInfo, error) {
	return ex.store.GetBlockInfo(blockNumber)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetBlocks(dbStartBlockNumber uint64, count int) (res []*api.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.GetBlocksInfo(dbStartBlockNumber, count)
}

// tx
func (ex *ExplorerBackend) GetTxInfo(txHash api.TxHash) (res *api.TxInfo, err error) {
	return ex.store.GetTxInfo(txHash)
}

func (ex *ExplorerBackend) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*api.TxInfo, err error) {
	return ex.store.GetBlockTxsByBlockNumber(blockNumber)
}

func (ex *ExplorerBackend) GetTxsByUnitID(unitID types.UnitID) ([]*api.TxInfo, error) {
	return ex.store.GetTxsByUnitID(unitID)
}

func (ex *ExplorerBackend) GetTxs(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error) {
	return ex.store.GetTxs(startSequenceNumber, count)
}

// bill
func (ex *ExplorerBackend) GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error) {
	unitIDs, err := ex.client.GetUnitsByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units by owner ID: %w", err)
	}
	var bills []*moneyApi.Bill
	for _, unitID := range unitIDs {
		if !unitID.HasType(txSysMoney.BillUnitType) {
			continue
		}
		bill, err := ex.client.GetBill(ctx, unitID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch unit: %w", err)
		}
		bills = append(bills, bill)
	}
	return bills, nil
}
