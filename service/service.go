package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
)

type (
	BlockStore interface {
		//block
		GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error)
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
		GetBlocksInRange(
			ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool,
		) (res []*domain.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(ctx context.Context, txHash domain.TxHash) (*domain.TxInfo, error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error)
		GetTxsPage(
			ctx context.Context,
			partitionID types.PartitionID,
			startID string,
			limit int,
		) (transactions []*domain.TxInfo, previousID string, err error)
	}

	ExplorerBackend struct {
		store            BlockStore
		partitionClients map[types.PartitionID]*PartitionClient
		sync.RWMutex
	}

	PartitionClient struct {
		*rpc.StateAPIClient
		partitionID     types.PartitionID
		partitionTypeID types.PartitionTypeID
	}

	PartitionRoundInfo struct {
		partitionID     types.PartitionID
		partitionTypeID types.PartitionTypeID
		RoundNumber     uint64
		EpochNumber     uint64
	}
)

func NewExplorerBackend(store BlockStore) *ExplorerBackend {
	return &ExplorerBackend{
		store:            store,
		partitionClients: map[types.PartitionID]*PartitionClient{},
	}
}

func (ex *ExplorerBackend) AddPartitionClient(
	client *rpc.StateAPIClient, partitionID types.PartitionID, partitionTypeID types.PartitionTypeID,
) {
	ex.Lock()
	defer ex.Unlock()
	ex.partitionClients[partitionID] = &PartitionClient{
		StateAPIClient:  client,
		partitionID:     partitionID,
		partitionTypeID: partitionTypeID,
	}
}

// GetRoundNumber returns the latest round and epoch number for all partitions
func (ex *ExplorerBackend) GetRoundNumber(ctx context.Context) ([]PartitionRoundInfo, error) {
	ex.RLock()
	defer ex.RUnlock()

	var result []PartitionRoundInfo
	for _, client := range ex.partitionClients {
		info, err := client.GetRoundInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get round info for partition %d: %w", client.partitionID, err)
		}
		result = append(result, PartitionRoundInfo{
			partitionID:     client.partitionID,
			partitionTypeID: client.partitionTypeID,
			RoundNumber:     info.RoundNumber,
			EpochNumber:     info.Epoch,
		})
	}
	return result, nil
}

// GetLastBlocks returns last processed blocks for each partition
func (ex *ExplorerBackend) GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error) {
	return ex.store.GetLastBlocks(ctx, partitionIDs, count, includeEmpty)
}

// GetBlock returns block with given block number for each specified partition.
func (ex *ExplorerBackend) GetBlock(
	ctx context.Context,
	blockNumber uint64,
	partitionIDs []types.PartitionID,
) (map[types.PartitionID]*domain.BlockInfo, error) {
	return ex.store.GetBlock(ctx, blockNumber, partitionIDs)
}

// GetBlocksInRange returns amount of blocks provided with count for given partition
func (ex *ExplorerBackend) GetBlocksInRange(
	ctx context.Context,
	partitionID types.PartitionID,
	dbStartBlock uint64,
	count int,
	includeEmpty bool,
) (res []*domain.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.GetBlocksInRange(ctx, partitionID, dbStartBlock, count, includeEmpty)
}

func (ex *ExplorerBackend) GetTxInfo(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error) {
	return ex.store.GetTxInfo(ctx, txHash)
}

func (ex *ExplorerBackend) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error) {
	return ex.store.GetTxsByBlockNumber(ctx, blockNumber, partitionID)
}

func (ex *ExplorerBackend) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error) {
	return ex.store.GetTxsByUnitID(ctx, unitID)
}

func (ex *ExplorerBackend) GetTxsPage(
	ctx context.Context,
	partitionID types.PartitionID,
	startID string,
	limit int,
) (transactions []*domain.TxInfo, previousID string, err error) {
	return ex.store.GetTxsPage(ctx, partitionID, startID, limit)
}

/*func (ex *ExplorerBackend) GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error) {
	return ex.client.GetUnitsByOwnerID(ctx, ownerID)
}

func (ex *ExplorerBackend) GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error) {
	return ex.GetUnitsByOwnerID(ctx, ownerID)
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
