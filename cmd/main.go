package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/ainvaltin/httpsrv"
	"github.com/alphabill-org/alphabill-explorer-backend/block_store/mongodb"
	"github.com/alphabill-org/alphabill-explorer-backend/blocks"
	"github.com/alphabill-org/alphabill-explorer-backend/blocksync"
	ra "github.com/alphabill-org/alphabill-explorer-backend/restapi"
	"github.com/alphabill-org/alphabill-explorer-backend/service"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("Starting AB Explorer")

	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
		fmt.Printf("reading config from %s\n", configPath)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}
	fmt.Printf("config: %+v\n", config)

	err = Run(context.Background(), config)
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context, config *Config) error {
	println("creating block store...")
	store, err := mongodb.NewMongoBlockStore(ctx, config.DB.URL)
	if err != nil {
		return fmt.Errorf("failed to get storage: %w", err)
	}
	println("created store")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		println("block explorer backend REST server starting on ", config.Server.Address)
		explorerBackend := service.NewExplorerBackend(store)

		handler := &ra.RestAPI{
			Service: explorerBackend,
		}

		return httpsrv.Run(
			ctx,
			&http.Server{
				Addr:              config.Server.Address,
				Handler:           handler.Router(),
				ReadTimeout:       3 * time.Second,
				ReadHeaderTimeout: time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       30 * time.Second,
			},
			httpsrv.ShutdownTimeout(5*time.Second))
	})

	for _, node := range config.Nodes {
		println("getting node info for ", node.URL)
		adminClient, err := rpc.NewAdminAPIClient(ctx, args.BuildRpcUrl(node.URL))
		if err != nil {
			return fmt.Errorf("failed to dial rpc client: %w", err)
		}

		nodeInfo, err := adminClient.GetNodeInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get node info: %w", err)
		}
		println("partition ID: ", nodeInfo.PartitionID)

		stateClient, err := rpc.NewStateAPIClient(ctx, args.BuildRpcUrl(node.URL))
		if err != nil {
			return fmt.Errorf("failed to dial rpc client: %w", err)
		}

		roundInfo, err := stateClient.GetRoundInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get round info: %w", err)
		}
		if node.BlockNumber > roundInfo.RoundNumber {
			return fmt.Errorf("current round number for partition %d (%d) is smaller than configured starting block number (%d)",
				nodeInfo.PartitionID, roundInfo.RoundNumber, node.BlockNumber)
		}

		g.Go(func() error {
			blockProcessor, err := blocks.NewBlockProcessor(store)
			if err != nil {
				return fmt.Errorf("failed to create block processor: %w", err)
			}
			getBlockNumber := func(ctx context.Context, partitionID types.PartitionID) (uint64, error) {
				storedBN, err := store.GetBlockNumber(ctx, partitionID)
				println("stored block number: ", storedBN)
				if err != nil {
					return 0, fmt.Errorf("failed to read current block number: %w", err)
				}
				if node.BlockNumber > storedBN {
					return node.BlockNumber, nil
				}
				return storedBN, nil
			}

			getRoundNumber := func(ctx context.Context) (uint64, error) {
				info, err := stateClient.GetRoundInfo(ctx)
				if err != nil {
					return 0, err
				}
				return info.RoundNumber, nil
			}

			// we act as if all errors returned by block sync are recoverable ie we
			// just retry in a loop until ctx is cancelled
			for {
				println("starting block sync")
				err := runBlockSync(ctx, stateClient.GetBlock, getRoundNumber, getBlockNumber, 100,
					blockProcessor.ProcessBlock, nodeInfo.PartitionID, nodeInfo.PartitionTypeID)
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
	}

	return g.Wait()
}

func runBlockSync(
	ctx context.Context,
	getBlocks blocksync.BlockLoaderFunc,
	getRoundNumber blocksync.GetRoundNumberFunc,
	getBlockNumber func(ctx context.Context, partitionID types.PartitionID) (uint64, error),
	batchSize int,
	processor blocksync.BlockProcessorFunc,
	partitionID types.PartitionID,
	partitionTypeID types.PartitionTypeID,
) error {
	blockNumber, err := getBlockNumber(ctx, partitionID)
	if err != nil {
		return fmt.Errorf("failed to read current block number for a sync starting point: %w", err)
	}
	// on bootstrap storage returns 0 as current block and as block numbering
	// starts from 1 by adding 1 to it we start with the first block
	return blocksync.Run(ctx, getBlocks, getRoundNumber, blockNumber+1, 0, batchSize, processor, partitionTypeID)
}
