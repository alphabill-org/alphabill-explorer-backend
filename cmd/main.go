package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/ainvaltin/httpsrv"
	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-explorer-backend/block_store/mongodb"
	"github.com/alphabill-org/alphabill-explorer-backend/blocks"
	"github.com/alphabill-org/alphabill-explorer-backend/blocksync"
	internalrpc "github.com/alphabill-org/alphabill-explorer-backend/client/rpc"
	"github.com/alphabill-org/alphabill-explorer-backend/internal/log"
	moneyservice "github.com/alphabill-org/alphabill-explorer-backend/service/money"
	"github.com/alphabill-org/alphabill-explorer-backend/service/partition"
	"github.com/alphabill-org/alphabill-explorer-backend/service/search"
	"github.com/alphabill-org/alphabill-go-base/txsystem/money"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	sdktypes "github.com/alphabill-org/alphabill-wallet/client/types"
	"golang.org/x/sync/errgroup"
)

func main() {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
		log.Info("reading config", "path", configPath)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to read config: %w", err))
	}

	if err = log.SetupLogger(&log.Configuration{
		Level:      config.Log.Level,
		Format:     config.Log.Format,
		OutputPath: config.Log.OutputPath,
	}); err != nil {
		log.Warn("Failed to setup logger, using default settings", "err", err)
	}

	log.Info("loaded config: ", "config", config)

	err = Run(context.Background(), config)
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context, config *Config) error {
	log.Info("creating block store...")
	store, err := mongodb.NewMongoBlockStore(ctx, config.DB.URL)
	if err != nil {
		return fmt.Errorf("failed to get storage: %w", err)
	}
	log.Info("created store")

	g, ctx := errgroup.WithContext(ctx)
	var moneyClient sdktypes.MoneyPartitionClient
	partitionService, err := partition.NewPartitionService(make(map[types.PartitionID]*partition.Partition))
	if err != nil {
		return fmt.Errorf("failed to create partition service")
	}
	searchService, err := search.NewSearchService(store, make(map[types.PartitionID]search.PartitionClient))
	if err != nil {
		return fmt.Errorf("failed to create search service")
	}

	for _, node := range config.Nodes {
		partitionClient, nodeInfo, err := createPartitionClient(ctx, node)
		if err != nil {
			return fmt.Errorf("failed to create partition client: %w", err)
		}

		partitionService.AddPartition(partitionClient, nodeInfo.PartitionID, nodeInfo.PartitionTypeID)
		searchService.AddPartitionClient(partitionClient, nodeInfo.PartitionID)
		if nodeInfo.PartitionTypeID == money.PartitionTypeID {
			moneyClient, err = client.NewMoneyPartitionClient(ctx, args.BuildRpcUrl(node.URL))
			if err != nil {
				return fmt.Errorf("failed to create money partition client: %w", err)
			}
		}

		g.Go(func() error {
			blockProcessor, err := blocks.NewBlockProcessor(store)
			if err != nil {
				return fmt.Errorf("failed to create block processor: %w", err)
			}
			getBlockNumber := func(ctx context.Context, partitionID types.PartitionID) (uint64, error) {
				storedBN, err := store.GetBlockNumber(ctx, partitionID)
				log.Info("stored block number", "block number", storedBN)
				if err != nil {
					return 0, fmt.Errorf("failed to read current block number: %w", err)
				}
				if node.BlockNumber > storedBN {
					return node.BlockNumber, nil
				}
				return storedBN, nil
			}

			getRoundNumber := func(ctx context.Context) (uint64, error) {
				info, err := partitionClient.GetRoundInfo(ctx)
				if err != nil {
					return 0, err
				}
				return info.RoundNumber, nil
			}

			// we act as if all errors returned by block sync are recoverable ie we
			// just retry in a loop until ctx is cancelled
			for {
				log.Info("starting block sync")
				err := runBlockSync(ctx, partitionClient.GetBlock, getRoundNumber, getBlockNumber, 100,
					blockProcessor.ProcessBlock, nodeInfo.PartitionID, nodeInfo.PartitionTypeID)
				if err != nil {
					log.Error("synchronizing blocks returned error", "err", err)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(rand.Int31n(10)+10) * time.Second):
				}
			}
		})
	}

	g.Go(func() error {
		log.Info("block explorer REST server starting", "address", config.Server.Address)
		controller, err := api.NewController(store, partitionService, moneyservice.NewMoneyService(moneyClient), searchService)
		if err != nil {
			return fmt.Errorf("failed to create controller for rest API: %w", err)
		}

		return httpsrv.Run(
			ctx,
			&http.Server{
				Addr:              config.Server.Address,
				Handler:           controller.Router(),
				ReadTimeout:       3 * time.Second,
				ReadHeaderTimeout: time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       30 * time.Second,
			},
			httpsrv.ShutdownTimeout(5*time.Second))
	})

	return g.Wait()
}

func createPartitionClient(ctx context.Context, node Node) (*internalrpc.StateAPIClient, *sdktypes.NodeInfoResponse, error) {
	log.Info("getting node info", "url", node.URL)
	adminClient, err := rpc.NewAdminAPIClient(ctx, args.BuildRpcUrl(node.URL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial rpc client: %w", err)
	}

	nodeInfo, err := adminClient.GetNodeInfo(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get node info: %w", err)
	}
	log.Info("received partition ID", "id", nodeInfo.PartitionID)
	adminClient.Close()

	stateClient, err := internalrpc.NewStateAPIClient(ctx, args.BuildRpcUrl(node.URL))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial rpc client: %w", err)
	}
	roundInfo, err := stateClient.GetRoundInfo(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get round info: %w", err)
	}
	if node.BlockNumber > roundInfo.RoundNumber {
		return nil, nil, fmt.Errorf("current round number for partition %d (%d) is smaller than configured starting block number (%d)",
			nodeInfo.PartitionID, roundInfo.RoundNumber, node.BlockNumber)
	}
	return stateClient, nodeInfo, nil
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
