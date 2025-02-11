package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
	wallettypes "github.com/alphabill-org/alphabill-wallet/client/types"
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
		GetTxByHash(ctx context.Context, txHash domain.TxHash) (*domain.TxInfo, error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error)
		GetTxsPage(
			ctx context.Context,
			partitionID types.PartitionID,
			startID string,
			limit int,
		) (transactions []*domain.TxInfo, previousID string, err error)
		FindTxs(ctx context.Context, searchKey []byte) ([]*domain.TxInfo, error)
	}

	ExplorerService struct {
		store            BlockStore
		partitionClients map[types.PartitionID]*PartitionClient
		moneyClient      wallettypes.MoneyPartitionClient
		sync.RWMutex
	}

	PartitionClient struct {
		RoundInfoClient
		partitionID     types.PartitionID
		partitionTypeID types.PartitionTypeID
	}

	RoundInfoClient interface {
		GetRoundInfo(ctx context.Context) (*wallettypes.RoundInfo, error)
	}

	PartitionRoundInfo struct {
		partitionID     types.PartitionID
		partitionTypeID types.PartitionTypeID
		RoundNumber     uint64
		EpochNumber     uint64
	}
)

func NewExplorerService(store BlockStore) *ExplorerService {
	return &ExplorerService{
		store:            store,
		partitionClients: map[types.PartitionID]*PartitionClient{},
	}
}

func (ex *ExplorerService) AddMoneyClient(client wallettypes.MoneyPartitionClient) {
	ex.moneyClient = client
}

func (ex *ExplorerService) AddPartitionClient(
	client RoundInfoClient, partitionID types.PartitionID, partitionTypeID types.PartitionTypeID,
) {
	ex.Lock()
	defer ex.Unlock()
	ex.partitionClients[partitionID] = &PartitionClient{
		RoundInfoClient: client,
		partitionID:     partitionID,
		partitionTypeID: partitionTypeID,
	}
}

// GetRoundNumber returns the latest round and epoch number for all partitions
func (ex *ExplorerService) GetRoundNumber(ctx context.Context) ([]PartitionRoundInfo, error) {
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
func (ex *ExplorerService) GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error) {
	return ex.store.GetLastBlocks(ctx, partitionIDs, count, includeEmpty)
}

// GetBlock returns block with given block number for each specified partition.
func (ex *ExplorerService) GetBlock(
	ctx context.Context,
	blockNumber uint64,
	partitionIDs []types.PartitionID,
) (map[types.PartitionID]*domain.BlockInfo, error) {
	return ex.store.GetBlock(ctx, blockNumber, partitionIDs)
}

// GetBlocksInRange returns amount of blocks provided with count for given partition
func (ex *ExplorerService) GetBlocksInRange(
	ctx context.Context,
	partitionID types.PartitionID,
	dbStartBlock uint64,
	count int,
	includeEmpty bool,
) (res []*domain.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.GetBlocksInRange(ctx, partitionID, dbStartBlock, count, includeEmpty)
}

func (ex *ExplorerService) GetTxByHash(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error) {
	return ex.store.GetTxByHash(ctx, txHash)
}

func (ex *ExplorerService) GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error) {
	return ex.store.GetTxsByBlockNumber(ctx, blockNumber, partitionID)
}

func (ex *ExplorerService) GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error) {
	return ex.store.GetTxsByUnitID(ctx, unitID)
}

func (ex *ExplorerService) GetTxsPage(
	ctx context.Context,
	partitionID types.PartitionID,
	startID string,
	limit int,
) (transactions []*domain.TxInfo, previousID string, err error) {
	return ex.store.GetTxsPage(ctx, partitionID, startID, limit)
}

func (ex *ExplorerService) FindTxs(ctx context.Context, searchKey []byte) ([]*domain.TxInfo, error) {
	return ex.store.FindTxs(ctx, searchKey)
}

func (ex *ExplorerService) GetBillsByPubKey(ctx context.Context, ownerID hex.Bytes) ([]*wallettypes.Bill, error) {
	if ex.moneyClient == nil {
		return nil, errors.New("bills partition not configured")
	}
	bills, err := ex.moneyClient.GetBills(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bills by owner ID: %w", err)
	}
	return bills, nil
}
