package money

import (
	"context"
	"errors"
	"fmt"

	"github.com/alphabill-org/alphabill-go-base/types/hex"
	"github.com/alphabill-org/alphabill-wallet/client/types"
)

type (
	Service struct {
		moneyClient types.MoneyPartitionClient
	}
)

func NewMoneyService(moneyClient types.MoneyPartitionClient) *Service {
	return &Service{moneyClient: moneyClient}
}

func (m *Service) GetBillsByPubKeyHash(ctx context.Context, ownerID hex.Bytes) ([]*types.Bill, error) {
	if m.moneyClient == nil {
		return nil, errors.New("money partition not configured")
	}
	bills, err := m.moneyClient.GetBills(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bills by owner ID: %w", err)
	}
	return bills, nil
}
