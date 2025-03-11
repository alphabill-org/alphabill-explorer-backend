package search

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/internal/log"
	"github.com/alphabill-org/alphabill-explorer-backend/util"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
	sdktypes "github.com/alphabill-org/alphabill-wallet/client/types"
)

type (
	Service struct {
		store            BlockStore
		partitionClients map[types.PartitionID]PartitionClient
	}

	PartitionClient interface {
		GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)
		GetUnit(ctx context.Context, unitID types.UnitID, includeStateProof bool) (*sdktypes.Unit[any], error)
	}

	BlockStore interface {
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
		FindTxs(ctx context.Context, searchKey []byte, partitionIDs []types.PartitionID) ([]*domain.TxInfo, error)
	}

	Result struct {
		Blocks  map[types.PartitionID]*domain.BlockInfo
		Txs     []*domain.TxInfo
		UnitIDs map[types.PartitionID][]types.UnitID
		Unit    *sdktypes.Unit[any]
	}
)

func NewSearchService(store BlockStore, partitionClients map[types.PartitionID]PartitionClient) (*Service, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if partitionClients == nil {
		return nil, errors.New("partitionClients is nil")
	}
	return &Service{
		store:            store,
		partitionClients: partitionClients,
	}, nil
}

func (s *Service) Search(ctx context.Context, searchKey string, partitionIDs []types.PartitionID) (*Result, error) {
	blockNumber, err := strconv.ParseUint(searchKey, 10, 64)
	if err == nil {
		blockMap, err := s.store.GetBlock(ctx, blockNumber, partitionIDs)
		if err != nil || len(blockMap) == 0 {
			return nil, domain.ErrNotFound
		}
		return &Result{Blocks: blockMap}, nil
	}

	searchKeyBytes, err := util.DecodeHex(searchKey)
	if err != nil {
		return nil, domain.ErrFailedToDecodeHex
	}

	var (
		wg                 sync.WaitGroup
		partitionsToSearch []types.PartitionID
		txs                []*domain.TxInfo
		units              = make(map[types.PartitionID][]types.UnitID)
		unit               *sdktypes.Unit[any]
	)
	if len(partitionIDs) > 0 {
		for _, partitionID := range partitionIDs {
			if _, exists := s.partitionClients[partitionID]; exists {
				partitionsToSearch = append(partitionsToSearch, partitionID)
			} else {
				log.Warn("Partition does not exist in search service", "id", partitionID)
			}
		}
	} else {
		for partitionID, _ := range s.partitionClients {
			partitionsToSearch = append(partitionsToSearch, partitionID)
		}
	}

	pubKeyHash, _ := util.PubKeyHash(searchKey)
	if pubKeyHash != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			units = s.findUnitsByOwnerPubKey(ctx, pubKeyHash, partitionsToSearch)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		unit = s.findUnit(ctx, searchKeyBytes, partitionsToSearch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var txsErr error
		txs, txsErr = s.store.FindTxs(ctx, searchKeyBytes, partitionIDs)
		if txsErr != nil && txsErr != domain.ErrNotFound {
			fmt.Println("Tx search failed: ", txsErr)
		}
	}()
	wg.Wait()

	return &Result{
		Txs:     txs,
		UnitIDs: units,
		Unit:    unit,
	}, nil
}

func (s *Service) findUnit(ctx context.Context, unitID types.UnitID, partitionIDs []types.PartitionID) *sdktypes.Unit[any] {
	if len(partitionIDs) == 0 {
		return nil
	}
	var (
		wg                    sync.WaitGroup
		unitResultChan        = make(chan *sdktypes.Unit[any], 1)
		ctxWithCancel, cancel = context.WithCancel(ctx)
	)
	defer cancel()

	for _, partitionID := range partitionIDs {
		client, exists := s.partitionClients[partitionID]
		if !exists {
			log.Info("Skipping unknown partition", "id", partitionID)
			continue
		}
		wg.Add(1)
		go func(c PartitionClient) {
			defer wg.Done()
			unit, err := c.GetUnit(ctxWithCancel, unitID, false)
			if unit != nil {
				cancel() // Cancel other goroutines if we found a result
				select {
				case unitResultChan <- unit:
				default:
					// channel full, do nothing
				}
			}
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Error("Failed to get unit", "err", err)
			}
		}(client)
	}
	wg.Wait()

	select {
	case unit := <-unitResultChan:
		return unit
	default:
		return nil
	}
}

func (s *Service) findUnitsByOwnerPubKey(ctx context.Context, pubKey []byte, partitionIDs []types.PartitionID) map[types.PartitionID][]types.UnitID {
	if len(partitionIDs) == 0 {
		return make(map[types.PartitionID][]types.UnitID)
	}
	var (
		wg              sync.WaitGroup
		unitResultsChan = make(chan map[types.PartitionID][]types.UnitID, len(partitionIDs))
		units           = make(map[types.PartitionID][]types.UnitID)
	)

	for _, partitionID := range partitionIDs {
		client, exists := s.partitionClients[partitionID]
		if !exists {
			log.Info("Skipping unknown partition", "id", partitionID)
			continue
		}
		wg.Add(1)
		go func(pid types.PartitionID, c PartitionClient) {
			defer wg.Done()
			unitIDs, err := c.GetUnitsByOwnerID(ctx, pubKey)
			unitResultsChan <- map[types.PartitionID][]types.UnitID{pid: unitIDs}
			if err != nil {
				log.Error("Failed to get units by owner ID", "err", err)
			}
		}(partitionID, client)
	}
	wg.Wait()
	close(unitResultsChan)

	for kv := range unitResultsChan {
		for partitionID, unitIDs := range kv {
			units[partitionID] = unitIDs
		}
	}

	return units
}

func (s *Service) AddPartitionClient(client PartitionClient, partitionID types.PartitionID) {
	s.partitionClients[partitionID] = client
}
