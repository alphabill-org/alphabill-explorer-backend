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
	s "github.com/alphabill-org/alphabill-explorer-backend/store"
)

type (
	ExplorerBackendService interface {
		GetLastBlockNumber() (uint64, error)
		GetBlockByBlockNumber(blockNumber uint64) (*types.Block, error)
		GetBlocks(dbStartBlock uint64, count int) (res []*types.Block, prevBlockNumber uint64, err error)
		GetBlockExplorerByBlockNumber(blockNumber uint64) (*s.BlockExplorer, error)
		GetBlocksExplorer(dbStartBlock uint64, count int) (res []*s.BlockExplorer, prevBlockNumber uint64, err error)
		GetTxExplorerByTxHash(txHash string) (*s.TxExplorer, error)
		GetBlockExplorerTxsByBlockNumber(blockNumber uint64) (res []*s.TxExplorer, err error)
		GetRoundNumber(ctx context.Context) (uint64, error)
		GetTxProof(unitID types.UnitID, txHash sdk.TxHash) (*types.TxProof, error)
		//GetTxHistoryRecords(dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
		//GetTxHistoryRecordsByKey(hash sdk.PubKeyHash, dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
	}

	ExplorerBackend struct {
		store  s.BillStore
		client *rpc.Client
	}

	// Bills struct {
	// 	Bills []*Bill `json:"bills"`
	// }

	// Bill struct {
	// 	Id                   []byte `json:"id"`
	// 	Value                uint64 `json:"value"`
	// 	TxHash               []byte `json:"txHash"`
	// 	DCTargetUnitID       []byte `json:"dcTargetUnitId,omitempty"`
	// 	DCTargetUnitBacklink []byte `json:"dcTargetUnitBacklink,omitempty"`
	// 	OwnerPredicate       []byte `json:"ownerPredicate"`

	// 	// fcb specific fields
	// 	// LastAddFCTxHash last add fee credit tx hash
	// 	LastAddFCTxHash []byte `json:"lastAddFcTxHash,omitempty"`
	// }

	Pubkey struct {
		Pubkey     []byte             `json:"pubkey"`
		PubkeyHash *account.KeyHashes `json:"pubkeyHash"`
	}

	// // BillStore type for creating BillStoreTx transactions
	// BillStore interface {
	// 	Do() BillStoreTx
	// 	WithTransaction(func(tx BillStoreTx) error) error
	// }

	// // BillStoreTx type for managing units by their ID and owner condition
	// BillStoreTx interface {
	// 	GetLastBlockNumber() (uint64, error)
	// 	GetBlockByBlockNumber(blockNumber uint64) (*types.Block, error)
	// 	GetBlocks(dbStartBlock uint64, count int) (res []*types.Block, prevBlockNumber uint64, err error)
	// 	SetBlock(b *types.Block) error
	// 	GetBlockExplorerByBlockNumber(blockNumber uint64) (*BlockExplorer, error)
	// 	GetBlocksExplorer(dbStartBlock uint64, count int) (res []*BlockExplorer, prevBlockNumber uint64, err error)
	// 	SetBlockExplorer(b *types.Block) error
	// 	GetBlockExplorerTxsByBlockNumber(blockNumber uint64) (res []*TxExplorer, err error)
	// 	GetBlockNumber() (uint64, error)
	// 	SetBlockNumber(blockNumber uint64) error
	// 	GetTxExplorerByTxHash(txHash string) (*TxExplorer, error)
	// 	SetTxExplorerToBucket(txExplorer *TxExplorer) error
	// 	GetBill(unitID []byte) (*Bill, error)
	// 	GetBills(ownerCondition []byte, includeDCBills bool, offsetKey []byte, limit int) ([]*Bill, []byte, error)
	// 	SetBill(bill *Bill, proof *types.TxProof) error
	// 	RemoveBill(unitID []byte) error
	// 	GetSystemDescriptionRecords() ([]*genesis.SystemDescriptionRecord, error)
	// 	SetSystemDescriptionRecords(sdrs []*genesis.SystemDescriptionRecord) error
	// 	GetTxProof(unitID types.UnitID, txHash sdk.TxHash) (*types.TxProof, error)
	// 	StoreTxProof(unitID types.UnitID, txHash sdk.TxHash, txProof *types.TxProof) error
	// 	//GetTxHistoryRecords(dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
	// 	//GetTxHistoryRecordsByKey(hash sdk.PubKeyHash, dbStartKey []byte, count int) ([]*sdk.TxHistoryRecord, []byte, error)
	// 	//StoreTxHistoryRecord(hash sdk.PubKeyHash, rec *sdk.TxHistoryRecord) error
	// }

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

		handler := &moneyRestAPI{
			Service:            explorerBackend,
			ListBillsPageLimit: config.ListBillsPageLimit,
			SystemID:           config.ABMoneySystemIdentifier,
			rw:                 &ResponseWriter{},
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
func (ex *ExplorerBackend) GetBlockByBlockNumber(blockNumber uint64) (*types.Block, error) {
	return ex.store.Do().GetBlockByBlockNumber(blockNumber)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetBlocks(dbStartBlockNumber uint64, count int) (res []*types.Block, prevBlockNUmber uint64, err error) {
	return ex.store.Do().GetBlocks(dbStartBlockNumber, count)
}

// GetBlockByBlockNumber returns block with given block number.
func (ex *ExplorerBackend) GetBlockExplorerByBlockNumber(blockNumber uint64) (*s.BlockExplorer, error) {
	return ex.store.Do().GetBlockExplorerByBlockNumber(blockNumber)
}

// GetBlocksExplorer return amount of blocks provided with count
func (ex *ExplorerBackend) GetBlocksExplorer(dbStartBlockNumber uint64, count int) (res []*s.BlockExplorer, prevBlockNUmber uint64, err error) {
	return ex.store.Do().GetBlocksExplorer(dbStartBlockNumber, count)
}

// GetBlocks return amount of blocks provided with count
func (ex *ExplorerBackend) GetTxExplorerByTxHash(txHash string) (res *s.TxExplorer, err error) {
	return ex.store.Do().GetTxExplorerByTxHash(txHash)
}

// GetBill returns most recently seen bill with given unit id.
func (ex *ExplorerBackend) GetBill(unitID []byte) (*s.Bill, error) {
	return ex.store.Do().GetBill(unitID)
}

func (ex *ExplorerBackend) GetTxProof(unitID types.UnitID, txHash sdk.TxHash) (*types.TxProof, error) {
	return ex.store.Do().GetTxProof(unitID, txHash)
}
func (ex *ExplorerBackend) GetBlockExplorerTxsByBlockNumber(blockNumber uint64) (res []*s.TxExplorer, err error) {
	return ex.store.Do().GetBlockExplorerTxsByBlockNumber(blockNumber)
}

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
