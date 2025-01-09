package blocks

import (
	"context"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
)

type Store interface {
	GetBlockNumber(ctx context.Context, partitionID types.PartitionID) (uint64, error)
	SetBlockNumber(ctx context.Context, partitionID types.PartitionID, blockNumber uint64) error
	SetTxInfo(ctx context.Context, txInfo *api.TxInfo) error
	SetBlockInfo(ctx context.Context, blockInfo *api.BlockInfo) error
}

type BlockProcessor struct {
	store Store
}

func NewBlockProcessor(store Store) (*BlockProcessor, error) {
	return &BlockProcessor{store: store}, nil
}

func (p *BlockProcessor) ProcessBlock(ctx context.Context, b *types.Block) error {
	roundNumber, err := b.GetRoundNumber()
	if err != nil {
		return err
	}
	println("processing block: ", roundNumber)
	if len(b.Transactions) > 0 {
		fmt.Printf("Block number: %d has %d transactions\n", roundNumber, len(b.Transactions))
	}
	lastBlockNumber, err := p.store.GetBlockNumber(ctx, b.PartitionID())
	if err != nil {
		return err
	}
	if lastBlockNumber >= roundNumber {
		return fmt.Errorf("invalid block number. Received blockNumber %d current wallet blockNumber %d", roundNumber, lastBlockNumber)
	}
	for i, tx := range b.Transactions {
		if err := p.processTx(ctx, tx, b, i); err != nil {
			return fmt.Errorf("failed to process transaction: %w", err)
		}
	}
	err = p.saveBlock(ctx, b)
	if err != nil {
		return err
	}
	return p.store.SetBlockNumber(ctx, b.PartitionID(), roundNumber)
}

func (p *BlockProcessor) processTx(ctx context.Context, txr *types.TransactionRecord, b *types.Block, txIdx int) error {
	/*txo := txr.TransactionOrder
	txHash := txo.Hash(crypto.SHA256)
	_ = txHash
	proof, _, err := types.NewTxProof(b, txIdx, crypto.SHA256)
	if err != nil {
		return err
	}

	_ = proof // TODO: save proofs?*/

	roundNumber, err := b.GetRoundNumber()
	if err != nil {
		return err
	}

	txInfo, err := api.NewTxInfo(b.PartitionID(), roundNumber, txr)

	if err != nil {
		return fmt.Errorf("failed create new txInfo in ProcessBlock: %w", err)
	}

	err = p.saveTx(ctx, txInfo)
	if err != nil {
		return fmt.Errorf("failed to save tx in ProcessBlock: %w", err)
	}

	return nil
}

func (p *BlockProcessor) saveTx(ctx context.Context, txInfo *api.TxInfo) error {
	if txInfo == nil {
		return fmt.Errorf("transaction is nil")
	}
	err := p.store.SetTxInfo(ctx, txInfo)
	if err != nil {
		return err
	}
	return nil
}

func (p *BlockProcessor) saveBlock(ctx context.Context, b *types.Block) error {
	if b == nil {
		return fmt.Errorf("block is nil")
	}
	blockInfo, err := api.NewBlockInfo(b)
	if err != nil {
		return err
	}
	err = p.store.SetBlockInfo(ctx, blockInfo)
	if err != nil {
		return err
	}
	return nil
}
