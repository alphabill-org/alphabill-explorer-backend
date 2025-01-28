package blocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetBlockNumber(ctx context.Context, partitionID types.PartitionID) (uint64, error) {
	args := m.Called(ctx, partitionID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStore) SetBlockNumber(ctx context.Context, partitionID types.PartitionID, blockNumber uint64) error {
	args := m.Called(ctx, partitionID, blockNumber)
	return args.Error(0)
}

func (m *MockStore) SetTxInfo(ctx context.Context, txExplorer *domain.TxInfo) error {
	args := m.Called(ctx, txExplorer)
	return args.Error(0)
}

func (m *MockStore) SetBlockInfo(ctx context.Context, blockInfo *domain.BlockInfo) error {
	args := m.Called(ctx, blockInfo)
	return args.Error(0)
}

func TestBlockProcessor_Success(t *testing.T) {
	store := new(MockStore)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.On("GetBlockNumber", mock.Anything, partitionID).Return(uint64(1), nil)
	store.On("SetBlockNumber", mock.Anything, partitionID, uint64(2)).Return(nil)
	store.On("SetTxInfo", mock.Anything, mock.Anything).Return(nil)
	store.On("SetBlockInfo", mock.Anything, mock.Anything).Return(nil)

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	unicityCertificate, err := (&types.UnicityCertificate{InputRecord: &types.InputRecord{RoundNumber: 2}}).MarshalCBOR()
	require.NoError(t, err)

	block := &types.Block{
		Header: &types.Header{PartitionID: partitionID},
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
		UnicityCertificate: unicityCertificate,
	}

	err = blockProcessor.ProcessBlock(context.Background(), block, partitionTypeID)
	require.NoError(t, err)

	store.AssertExpectations(t)
}

func TestBlockProcessor_FailOnGetBlockNumber(t *testing.T) {
	store := new(MockStore)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.On("GetBlockNumber", mock.Anything, partitionID).Return(uint64(0), fmt.Errorf("some error"))

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	unicityCertificate, err := (&types.UnicityCertificate{}).MarshalCBOR()
	require.NoError(t, err)
	block := &types.Block{
		Header:             &types.Header{PartitionID: partitionID},
		UnicityCertificate: unicityCertificate,
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
	}

	err = blockProcessor.ProcessBlock(context.Background(), block, partitionTypeID)
	require.Error(t, err)

	store.AssertExpectations(t)
}

func TestBlockProcessor_FailOnSetBlockNumber(t *testing.T) {
	store := new(MockStore)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.On("GetBlockNumber", mock.Anything, partitionID).Return(uint64(1), nil)
	store.On("SetBlockNumber", mock.Anything, partitionID, uint64(2)).Return(fmt.Errorf("some error"))
	store.On("SetTxInfo", mock.Anything, mock.Anything).Return(nil)
	store.On("SetBlockInfo", mock.Anything, mock.Anything).Return(nil)

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	unicityCertificate, err := (&types.UnicityCertificate{InputRecord: &types.InputRecord{RoundNumber: 2}}).MarshalCBOR()
	require.NoError(t, err)

	block := &types.Block{
		Header: &types.Header{PartitionID: partitionID},
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
		UnicityCertificate: unicityCertificate,
	}

	err = blockProcessor.ProcessBlock(context.Background(), block, partitionTypeID)
	require.Error(t, err)

	store.AssertExpectations(t)
}
