package search

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"sync"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/errors"
	"github.com/alphabill-org/alphabill-explorer-backend/util"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
)

type (
	Service struct {
		store            BlockStore
		partitionClients map[types.PartitionID]PartitionClient
	}

	PartitionClient interface {
		GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)
	}

	BlockStore interface {
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
		FindTxs(ctx context.Context, searchKey []byte, partitionIDs []types.PartitionID) ([]*domain.TxInfo, error)
	}

	Result struct {
		Blocks map[types.PartitionID]*domain.BlockInfo
		Txs    []*domain.TxInfo
		Units  map[types.PartitionID][]types.UnitID
	}
)

func NewSearchService(store BlockStore, partitionClients map[types.PartitionID]PartitionClient) (*Service, error) {
	if store == nil || partitionClients == nil {
		return nil, errors.ErrNilArgument
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
			return nil, errors.ErrNotFound
		}
		return &Result{Blocks: blockMap}, nil
	}

	searchKeyBytes, err := util.Decode(searchKey)
	if err != nil {
		return nil, errors.ErrFailedToDecodeHex
	}

	var (
		wg                 sync.WaitGroup
		partitionsToSearch []types.PartitionID
		txs                []*domain.TxInfo
		units              = make(map[types.PartitionID][]types.UnitID)
	)
	if len(partitionIDs) > 0 {
		for _, partitionID := range partitionIDs {
			if _, exists := s.partitionClients[partitionID]; exists {
				partitionsToSearch = append(partitionsToSearch, partitionID)
			} else {
				fmt.Printf("Warning: Partition %d does not exist in search service\n", partitionID)
			}
		}
	} else {
		for partitionID, _ := range s.partitionClients {
			partitionsToSearch = append(partitionsToSearch, partitionID)
		}
	}

	var pubKeyHash []byte
	if len(searchKeyBytes) == util.PubKeyBytesLength {
		hash := sha256.Sum256(searchKeyBytes)
		pubKeyHash = hash[:]
	} else if len(searchKeyBytes) == util.PubKeyHashBytesLength {
		pubKeyHash = searchKeyBytes
	}

	if pubKeyHash != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			units = s.searchUnitsByOwnerPubKey(ctx, pubKeyHash, partitionsToSearch)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var txsErr error
		txs, txsErr = s.store.FindTxs(ctx, searchKeyBytes, partitionIDs)
		if txsErr != nil && txsErr != errors.ErrNotFound {
			fmt.Println("Tx search failed: ", txsErr)
		}
	}()
	wg.Wait()

	return &Result{
		Txs:   txs,
		Units: units,
	}, nil
}

func (s *Service) searchUnitsByOwnerPubKey(ctx context.Context, pubKey []byte, partitionIDs []types.PartitionID) map[types.PartitionID][]types.UnitID {
	var (
		wg              sync.WaitGroup
		unitResultsChan = make(chan map[types.PartitionID][]types.UnitID, len(partitionIDs))
		units           = make(map[types.PartitionID][]types.UnitID)
	)

	if len(partitionIDs) == 0 {
		return units
	}

	wg.Add(len(partitionIDs))
	for _, partitionID := range partitionIDs {
		client, exists := s.partitionClients[partitionID]
		if !exists {
			fmt.Println("Skipping unknown partition: ", partitionID)
			continue
		}
		go func(pid types.PartitionID, c PartitionClient) {
			defer wg.Done()
			unitIDs, err := c.GetUnitsByOwnerID(ctx, pubKey)
			unitResultsChan <- map[types.PartitionID][]types.UnitID{pid: unitIDs}
			if err != nil {
				fmt.Println("Failed to get units by owner ID: ", err)
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
