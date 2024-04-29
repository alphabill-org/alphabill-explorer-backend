package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ainvaltin/httpsrv"
	"github.com/alphabill-org/alphabill-explorer-backend/api"
	bs "github.com/alphabill-org/alphabill-explorer-backend/bill_store"
	"github.com/alphabill-org/alphabill-explorer-backend/blocks"
	"github.com/alphabill-org/alphabill-explorer-backend/blocksync"
	ra "github.com/alphabill-org/alphabill-explorer-backend/restapi"
	exTypes "github.com/alphabill-org/alphabill-explorer-backend/types"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	abTypes "github.com/alphabill-org/alphabill/types"
	"golang.org/x/sync/errgroup"
)

type (
	BillStore interface {
		GetLastBlockNumber() (uint64, error)
		GetBlockInfo(blockNumber uint64) (*api.BlockInfo, error)
		GetBlocksInfo(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)
		GetBlockTxsByBlockNumber(blockNumber uint64) (res []*exTypes.TxInfo, err error)
		GetTxInfo(txHash string) (*exTypes.TxInfo, error)
	}

	ExplorerBackend struct {
		store  BillStore
		client *rpc.Client
	}

	Config struct {
		ABMoneySystemIdentifier abTypes.SystemID
		AlphabillUrl            string
		ServerAddr              string
		DbFile                  string
		ListBillsPageLimit      int
		BlockNumber             uint64
	}
)

func Run(ctx context.Context, config *Config) error {
	println("starting money partition explorer")
	store, err := bs.NewBoltBillStore(config.DbFile)
	if err != nil {
		return fmt.Errorf("failed to get storage: %w", err)
	}

	moneyClient, err := rpc.DialContext(ctx, args.BuildRpcUrl(config.AlphabillUrl))
	if err != nil {
		return fmt.Errorf("failed to dial rpc client: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		println("money backend REST server starting on ", config.ServerAddr)
		explorerBackend := &ExplorerBackend{store: store, client: moneyClient}
		defer moneyClient.Close()

		handler := &ra.MoneyRestAPI{
			Service:            explorerBackend,
			ListBillsPageLimit: config.ListBillsPageLimit,
			SystemID:           config.ABMoneySystemIdentifier,
		}

		return httpsrv.Run(
			ctx,
			http.Server{
				Addr:              config.ServerAddr,
				Handler:           handler.Router(),
				ReadTimeout:       3 * time.Second,
				ReadHeaderTimeout: time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       30 * time.Second,
			},
			httpsrv.ShutdownTimeout(5*time.Second))
	})

	g.Go(func() error {
		blockProcessor, err := blocks.NewBlockProcessor(store, config.ABMoneySystemIdentifier)
		if err != nil {
			return fmt.Errorf("failed to create block processor: %w", err)
		}
		getBlockNumber := func() (uint64, error) {
			storedBN, err := store.GetBlockNumber()
			println("stored block number: ", storedBN)
			if err != nil {
				return 0, fmt.Errorf("failed to read current block number: %w", err)
			}
			if config.BlockNumber > storedBN {
				return config.BlockNumber, nil
			}
			return storedBN, nil
		}
		// we act as if all errors returned by block sync are recoverable ie we
		// just retry in a loop until ctx is cancelled
		for {
			println("starting block sync")
			err := runBlockSync(ctx, moneyClient.GetBlock, getBlockNumber, 100, blockProcessor.ProcessBlock)
			if err != nil {
				println(fmt.Errorf("synchronizing blocks returned error: %w", err).Error())
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(rand.Int31n(10)+10) * time.Second):
			}
		}
	})

	return g.Wait()
}

func runBlockSync(ctx context.Context, getBlocks blocksync.BlockLoaderFunc, getBlockNumber func() (uint64, error), batchSize int, processor blocksync.BlockProcessorFunc) error {
	blockNumber, err := getBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to read current block number for a sync starting point: %w", err)
	}
	// on bootstrap storage returns 0 as current block and as block numbering
	// starts from 1 by adding 1 to it we start with the first block
	return blocksync.Run(ctx, getBlocks, blockNumber+1, 0, batchSize, processor)
}

// GetLastBlockNumber returns last processed block
func (ex *ExplorerBackend) GetLastBlockNumber() (uint64, error) {
	return ex.store.GetLastBlockNumber()
}

// GetBlock returns block with given block number.
func (ex *ExplorerBackend) GetBlock(blockNumber uint64) (*api.BlockInfo, error) {
	return ex.store.GetBlockInfo(blockNumber)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetBlocks(dbStartBlockNumber uint64, count int) (res []*api.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.GetBlocksInfo(dbStartBlockNumber, count)
}

func (ex *ExplorerBackend) GetTxInfo(txHash string) (res *exTypes.TxInfo, err error) {
	return ex.store.GetTxInfo(txHash)
}

// GetRoundNumber returns latest round number.
func (ex *ExplorerBackend) GetRoundNumber(ctx context.Context) (uint64, error) {
	return ex.client.GetRoundNumber(ctx)
}

func (ex *ExplorerBackend) GetBlockTxsByBlockNumber(blockNumber uint64) (res []*exTypes.TxInfo, err error) {
	return ex.store.GetBlockTxsByBlockNumber(blockNumber)
}
