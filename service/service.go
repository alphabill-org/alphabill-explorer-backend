package service

import (
	"context"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
)

type (
	BillStore interface {

		//block
		GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
		GetBlocksInRange(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(ctx context.Context, txHash api.TxHash) (*api.TxInfo, error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*api.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*api.TxInfo, error)
		GetTxsInRange(ctx context.Context, partitionID types.PartitionID, startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error)
	}

	ABClient interface {
		GetRoundNumber(ctx context.Context) (uint64, error)
		GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)
		//GetBill(ctx context.Context, unitID types.UnitID, includeStateProof bool) (*moneyApi.Bill, error)
	}

	ExplorerBackend struct {
		store BillStore
	}
)

func NewExplorerBackend(store BillStore) *ExplorerBackend {
	return &ExplorerBackend{
		store: store,
	}
}

// GetRoundNumber returns latest round number.
func (ex *ExplorerBackend) GetRoundNumber(ctx context.Context) (uint64, error) {
	panic("not implemented")
	//return ex.client.GetRoundNumber(ctx)
}

// block
// GetLastBlock returns last processed block for each partition
func (ex *ExplorerBackend) GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error) {
	return ex.store.GetLastBlock(ctx, partitionIDs)
}

// GetBlock returns block with given block number for each specified partition.
func (ex *ExplorerBackend) GetBlock(
	ctx context.Context,
	blockNumber uint64,
	partitionIDs []types.PartitionID,
) (map[types.PartitionID]*api.BlockInfo, error) {
	return ex.store.GetBlock(ctx, blockNumber, partitionIDs)
}

// GetBlocksInRange returns amount of blocks provided with count for given partition
func (ex *ExplorerBackend) GetBlocksInRange(
	ctx context.Context,
	partitionID types.PartitionID,
	dbStartBlock uint64,
	count int,
) (res []*api.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.GetBlocksInRange(ctx, partitionID, dbStartBlock, count)
}

// tx
func (ex *ExplorerBackend) GetTxInfo(ctx context.Context, txHash api.TxHash) (res *api.TxInfo, err error) {
	return ex.store.GetTxInfo(ctx, txHash)
}

func (ex *ExplorerBackend) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*api.TxInfo, error) {
	return ex.store.GetTxsByBlockNumber(ctx, blockNumber, partitionID)
}

func (ex *ExplorerBackend) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*api.TxInfo, error) {
	return ex.store.GetTxsByUnitID(ctx, unitID)
}

func (ex *ExplorerBackend) GetTxsInRange(ctx context.Context, partitionID types.PartitionID, startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error) {
	return ex.store.GetTxsInRange(ctx, partitionID, startSequenceNumber, count)
}

/*func (ex *ExplorerBackend) GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error) {
	return ex.client.GetUnitsByOwnerID(ctx, ownerID)
}

// bill
func (ex *ExplorerBackend) GetBillsByPubKey(ctx context.Context, ownerID hex.Bytes) (res []types.UnitID, err error) {
	unitIDs, err := ex.client.GetUnitsByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units by owner ID: %w", err)
	}
	// todo get bill data
	return unitIDs, nil
}
*/
