package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ainvaltin/httpsrv"
	bs "github.com/alphabill-org/alphabill-explorer-backend/bill_store"
	"github.com/alphabill-org/alphabill-explorer-backend/blocks"
	"github.com/alphabill-org/alphabill-explorer-backend/blocksync"
	ra "github.com/alphabill-org/alphabill-explorer-backend/restapi"
	"github.com/alphabill-org/alphabill-explorer-backend/service"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	"golang.org/x/sync/errgroup"
)

type (
	Config struct {
		AlphabillUrl       string
		ServerAddr         string
		DbFile             string
		ListBillsPageLimit int
		BlockNumber        uint64
	}
)

func main() {
	fmt.Println("Starting AB Explorer")
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Usage: blocks <AB Partition RPC url> <AB Explorer url> [<Block number>]")
		return
	}
	workDir := filepath.Dir(args[0]) //"/tmp/"
	fmt.Printf("filepath: %s\n", filepath.Dir(args[0]))
	fmt.Printf("AB Partition url: %s\n", args[1])
	fmt.Printf("AB Explorer url: %s\n", args[2])
	blockNumber := uint64(0)
	if len(args) > 3 {
		fmt.Printf("Block number: %s\n", args[3])
		blockNumber, _ = strconv.ParseUint(args[3], 10, 64)
	}
	err := Run(context.Background(), &Config{
		AlphabillUrl: args[1],
		ServerAddr:   args[2],
		DbFile:       filepath.Join(workDir, bs.BoltExplorerStoreFileName),
		BlockNumber:  blockNumber,
	})
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context, config *Config) error {
	println("starting money partition explorer")
	store, err := bs.NewBoltBillStore(config.DbFile)
	if err != nil {
		return fmt.Errorf("failed to get storage: %w", err)
	}

	adminClient, err := rpc.NewAdminAPIClient(ctx, args.BuildRpcUrl(config.AlphabillUrl))
	if err != nil {
		return fmt.Errorf("failed to dial rpc client: %w", err)
	}

	info, err := adminClient.GetNodeInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node info: %w", err)
	}

	println("partition ID: ", info.PartitionID)

	moneyClient, err := rpc.NewStateAPIClient(ctx, args.BuildRpcUrl(config.AlphabillUrl))
	if err != nil {
		return fmt.Errorf("failed to dial rpc client: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		println("money backend REST server starting on ", config.ServerAddr)
		explorerBackend := service.NewExplorerBackend(store, moneyClient)
		defer moneyClient.Close()

		handler := &ra.MoneyRestAPI{
			Service:            explorerBackend,
			ListBillsPageLimit: config.ListBillsPageLimit,
			PartitionID:        info.PartitionID,
		}

		return httpsrv.Run(
			ctx,
			&http.Server{
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
		blockProcessor, err := blocks.NewBlockProcessor(store)
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
