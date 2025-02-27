package partition

import (
	"context"
	"fmt"
	"sync"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	wallettypes "github.com/alphabill-org/alphabill-wallet/client/types"
)

type (
	Service struct {
		partitions map[types.PartitionID]*Partition
		sync.RWMutex
	}

	Partition struct {
		RoundInfoClient
		partitionID     types.PartitionID
		partitionTypeID types.PartitionTypeID
	}

	RoundInfoClient interface {
		GetRoundInfo(ctx context.Context) (*wallettypes.RoundInfo, error)
	}

	RoundInfo struct {
		PartitionID     types.PartitionID
		PartitionTypeID types.PartitionTypeID
		RoundNumber     uint64
		EpochNumber     uint64
	}
)

func NewPartitionService(clients map[types.PartitionID]*Partition) (*Service, error) {
	if clients == nil {
		return nil, domain.ErrNilArgument
	}
	return &Service{
		partitions: clients,
		RWMutex:    sync.RWMutex{},
	}, nil
}

// GetRoundNumber returns the latest round and epoch number for all partitions
func (p *Service) GetRoundNumber(ctx context.Context) ([]RoundInfo, error) {
	p.RLock()
	defer p.RUnlock()

	var result []RoundInfo
	for _, client := range p.partitions {
		info, err := client.GetRoundInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get round info for partition %d: %w", client.partitionID, err)
		}
		result = append(result, RoundInfo{
			PartitionID:     client.partitionID,
			PartitionTypeID: client.partitionTypeID,
			RoundNumber:     info.RoundNumber,
			EpochNumber:     info.Epoch,
		})
	}
	return result, nil
}

func (p *Service) AddPartition(
	client RoundInfoClient, partitionID types.PartitionID, partitionTypeID types.PartitionTypeID,
) {
	p.Lock()
	defer p.Unlock()
	p.partitions[partitionID] = &Partition{
		RoundInfoClient: client,
		partitionID:     partitionID,
		partitionTypeID: partitionTypeID,
	}
}
