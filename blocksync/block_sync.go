package blocksync

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/alphabill-org/alphabill-wallet/wallet/money/api"
	"github.com/alphabill-org/alphabill/types"
	"golang.org/x/sync/errgroup"
)

type BlockLoaderFunc func(ctx context.Context, rn uint64) (*types.Block, error)
type BlockProcessorFunc func(context.Context, *types.Block) error

/*
Run loads blocks using "getBlocks" and processes them using "processor" until:
  - ctx is cancelled;
  - maxBlockNumber param is not zero and block with that number has been processed;
  - unrecoverable error is encountered.

Other parameters:
  - startingBlockNumber is the first block number to ask for (using getBlocks) must be > 0;
  - maxBlockNumber: when zero Run loads new blocks until ctx is cancelled, when not zero
    blocks are loaded until block with given number has been processed;
  - batchSize how big batches to use (used for getBlocks parameter);

Run returns non-nil error unless maxBlockNumber param is not zero and that block is
loaded and processed successfully.
*/
func Run(ctx context.Context, getBlock BlockLoaderFunc, startingBlockNumber, maxBlockNumber uint64, batchSize int, processor BlockProcessorFunc) error {
	if startingBlockNumber <= 0 {
		return fmt.Errorf("invalid sync condition: starting block number must be greater than zero, got %d", startingBlockNumber)
	}
	if batchSize <= 0 {
		return fmt.Errorf("invalid sync condition: batch size must be greater than zero, got %d", batchSize)
	}
	if maxBlockNumber != 0 {
		if maxBlockNumber < startingBlockNumber {
			return fmt.Errorf("invalid sync condition: starting block number %d is greater than max block number %d", startingBlockNumber, maxBlockNumber)
		}
		getBlock = loadUntilBlockNumber(maxBlockNumber, getBlock)
	}

	blocks := make(chan *types.Block, batchSize)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer close(blocks)
		err := fetchBlocks(ctx, getBlock, startingBlockNumber, blocks)
		if err != nil && errors.Is(err, errMaxBlockReached) {
			return nil
		}
		return err
	})

	g.Go(func() error {
		return processBlocks(ctx, blocks, processor)
	})

	return g.Wait()
}

func fetchBlocks(ctx context.Context, getBlock BlockLoaderFunc, blockNumber uint64, out chan<- *types.Block) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		block, err := getBlock(ctx, blockNumber)
		if err != nil && !errors.Is(err, api.ErrNotFound) {
			return fmt.Errorf("failed to fetch blocks [%d...]: %w", blockNumber, err)
		}
		if block != nil {
			out <- block
			blockNumber = block.GetRoundNumber() + 1
			continue
		}
		// we have reached to the last block the source currently has - wait a bit before asking for more
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(rand.Int31n(500)+500) * time.Millisecond):
		}
	}
}

func processBlocks(ctx context.Context, blocks <-chan *types.Block, processor BlockProcessorFunc) error {
	for b := range blocks {
		if err := processor(ctx, b); err != nil {
			return fmt.Errorf("failed to process block {%x : %d}: %w", b.SystemID(), b.GetRoundNumber(), err)
		}
	}
	return nil
}

func loadUntilBlockNumber(maxBN uint64, f BlockLoaderFunc) BlockLoaderFunc {
	return func(ctx context.Context, blockNumber uint64) (*types.Block, error) {
		if blockNumber > maxBN {
			return nil, errMaxBlockReached
		}
		//if blockNumber+batchSize > maxBN {
		//	batchSize = (maxBN - blockNumber) + 1
		//}
		return f(ctx, blockNumber)
	}
}

var errMaxBlockReached = fmt.Errorf("max block number has been reached")
