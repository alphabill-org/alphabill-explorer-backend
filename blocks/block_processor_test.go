package blocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetBlockNumber() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStore) SetBlockNumber(blockNumber uint64) error {
	args := m.Called(blockNumber)
	return args.Error(0)
}

func (m *MockStore) SetTxInfo(txExplorer *api.TxInfo) error {
	args := m.Called(txExplorer)
	return args.Error(0)
}

func (m *MockStore) SetBlockInfo(b *api.BlockInfo) error {
	args := m.Called(b)
	return args.Error(0)
}

func TestBlockProcessor_Success(t *testing.T) {
	store := new(MockStore)
	store.On("GetBlockNumber").Return(uint64(1), nil)
	store.On("SetBlockNumber", uint64(2)).Return(nil)
	store.On("SetTxInfo", mock.Anything).Return(nil)
	store.On("SetBlockInfo", mock.Anything).Return(nil)

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	unicityCertificate, err := (&types.UnicityCertificate{InputRecord: &types.InputRecord{RoundNumber: 2}}).MarshalCBOR()
	require.NoError(t, err)

	block := &types.Block{
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
		UnicityCertificate: unicityCertificate,
	}

	err = blockProcessor.ProcessBlock(context.Background(), block)
	require.NoError(t, err)

	store.AssertExpectations(t)
}

func TestBlockProcessor_FailOnGetBlockNumber(t *testing.T) {
	store := new(MockStore)
	store.On("GetBlockNumber").Return(uint64(0), fmt.Errorf("some error"))

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	block := &types.Block{
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
	}

	err = blockProcessor.ProcessBlock(context.Background(), block)
	require.Error(t, err)

	store.AssertExpectations(t)
}

func TestBlockProcessor_FailOnSetBlockNumber(t *testing.T) {
	store := new(MockStore)
	store.On("GetBlockNumber").Return(uint64(1), nil)
	store.On("SetBlockNumber", uint64(2)).Return(fmt.Errorf("some error"))
	store.On("SetTxInfo", mock.Anything).Return(nil)
	store.On("SetBlockInfo", mock.Anything).Return(nil)

	blockProcessor, err := NewBlockProcessor(store)
	require.NoError(t, err)

	txoBytes, err := (&types.TransactionOrder{}).MarshalCBOR()
	require.NoError(t, err)

	unicityCertificate, err := (&types.UnicityCertificate{InputRecord: &types.InputRecord{RoundNumber: 2}}).MarshalCBOR()
	require.NoError(t, err)

	block := &types.Block{
		Transactions: []*types.TransactionRecord{
			{
				TransactionOrder: txoBytes,
			},
		},
		UnicityCertificate: unicityCertificate,
	}

	err = blockProcessor.ProcessBlock(context.Background(), block)
	require.Error(t, err)

	store.AssertExpectations(t)
}
