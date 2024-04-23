package explorer

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ainvaltin/httpsrv"
	"github.com/alphabill-org/alphabill-explorer-backend/blocksync"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	sdk "github.com/alphabill-org/alphabill-wallet/wallet"
	"github.com/alphabill-org/alphabill-wallet/wallet/account"
	"github.com/alphabill-org/alphabill/types"
	"golang.org/x/sync/errgroup"
	bs "github.com/alphabill-org/alphabill-explorer-backend/explorer/bill_store"
	st "github.com/alphabill-org/alphabill-explorer-backend/store"
	ra "github.com/alphabill-org/alphabill-explorer-backend/explorer/restapi"
)


type (

	ExplorerBackend struct {
		store  st.BillStore
		client *rpc.Client
	}

	Pubkey struct {
		Pubkey     []byte             `json:"pubkey"`
		PubkeyHash *account.KeyHashes `json:"pubkeyHash"`
	}

	p2pkhOwnerPredicates struct {
		sha256 []byte
		sha512 []byte
	}

	Config struct {
		ABMoneySystemIdentifier types.SystemID
		AlphabillUrl            string
		ServerAddr              string
		DbFile                  string
		ListBillsPageLimit      int
		BlockNumber             uint64
	}

	InitialBill struct {
		ID        []byte
		Value     uint64
		Predicate []byte
	}
)

func Run(ctx context.Context, config *Config) error {
	println("starting money backend")
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
		server := http.Server{
			Addr:              config.ServerAddr,
			Handler:           handler.Router(),
			ReadTimeout:       3 * time.Second,
			ReadHeaderTimeout: time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       30 * time.Second,
		}

		return httpsrv.Run(ctx, server, httpsrv.ShutdownTimeout(5*time.Second))
	})

	g.Go(func() error {
		blockProcessor, err := NewBlockProcessor(store, config.ABMoneySystemIdentifier)
		if err != nil {
			return fmt.Errorf("failed to create block processor: %w", err)
		}
		getBlockNumber := func() (uint64, error) {
			storedBN, err := store.Do().GetBlockNumber()
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

// GetBlockByBlockNumber returns block with given block number.
func (ex *ExplorerBackend) GetLastBlockNumber() (uint64, error) {
	return ex.store.Do().GetLastBlockNumber()
}

// GetBlockByBlockNumber returns block with given block number.
func (ex *ExplorerBackend) GetBlockByBlockNumber(blockNumber uint64) (*st.BlockInfo, error) {
	return ex.store.Do().GetBlockInfoByBlockNumber(blockNumber)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetBlocks(dbStartBlockNumber uint64, count int) (res []*st.BlockInfo, prevBlockNUmber uint64, err error) {
	return ex.store.Do().GetBlocksInfo(dbStartBlockNumber, count)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetTxExplorerByTxHash(txHash string) (res *st.TxExplorer, err error) {
	return ex.store.Do().GetTxExplorerByTxHash(txHash)
}

// GetBill returns most recently seen bill with given unit id.
func (ex *ExplorerBackend) GetBill(unitID []byte) (*st.Bill, error) {
	return ex.store.Do().GetBill(unitID)
}

func (ex *ExplorerBackend) GetTxProof(unitID types.UnitID, txHash sdk.TxHash) (*types.TxProof, error) {
	return ex.store.Do().GetTxProof(unitID, txHash)
}
// func (ex *ExplorerBackend) GetBlockExplorerTxsByBlockNumber(blockNumber uint64) (res []*st.TxExplorer, err error) {
// 	return ex.store.Do().GetBlockExplorerTxsByBlockNumber(blockNumber)
// }

// GetRoundNumber returns latest round number.
func (ex *ExplorerBackend) GetRoundNumber(ctx context.Context) (uint64, error) {
	return ex.client.GetRoundNumber(ctx)
}

//func (ex *ExplorerBackend) GetTxHistoryRecords(dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error) {
//	return ex.store.Do().GetTxHistoryRecords(dbStartKey, count)
//}
//
//func (ex *ExplorerBackend) GetTxHistoryRecordsByKey(hash sdk.PubKeyHash, dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error) {
//	return ex.store.Do().GetTxHistoryRecordsByKey(hash, dbStartKey, count)
//}

// func (b *s.Bill) getTxHash() []byte {
// 	if b != nil {
// 		return b.TxHash
// 	}
// 	return nil
// }

// func (b *s.Bill) getValue() uint64 {
// 	if b != nil {
// 		return b.Value
// 	}
// 	return 0
// }

// func (b *s.Bill) getLastAddFCTxHash() []byte {
// 	if b != nil {
// 		return b.LastAddFCTxHash
// 	}
// 	return nil
// }

// func (b *s.Bill) IsDCBill() bool {
// 	if b != nil {
// 		return len(b.DCTargetUnitID) > 0
// 	}
// 	return false
// }
