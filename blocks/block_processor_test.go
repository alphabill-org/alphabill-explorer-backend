package blocks

import (
	"context"
	"fmt"
	"testing"

	mocks "github.com/alphabill-org/alphabill-explorer-backend/internal/mocks/blocks"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBlockProcessor_Success(t *testing.T) {
	store := mocks.NewStore(t)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.EXPECT().GetBlockNumber(mock.Anything, partitionID).Return(uint64(1), nil)
	store.EXPECT().SetBlockNumber(mock.Anything, partitionID, uint64(2)).Return(nil)
	store.EXPECT().SetTxInfo(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().SetBlockInfo(mock.Anything, mock.Anything).Return(nil)

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
	store := mocks.NewStore(t)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.EXPECT().GetBlockNumber(mock.Anything, partitionID).Return(uint64(0), fmt.Errorf("some error"))

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
	store := mocks.NewStore(t)
	partitionID := types.PartitionID(1)
	partitionTypeID := types.PartitionTypeID(2)
	store.EXPECT().GetBlockNumber(mock.Anything, partitionID).Return(uint64(1), nil)
	store.EXPECT().SetBlockNumber(mock.Anything, partitionID, uint64(2)).Return(fmt.Errorf("some error"))
	store.EXPECT().SetTxInfo(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().SetBlockInfo(mock.Anything, mock.Anything).Return(nil)

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
