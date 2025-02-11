package restapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/service"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
	wallettypes "github.com/alphabill-org/alphabill-wallet/client/types"
)

const (
	paramPartitionID  = "partitionID"
	paramBlockNumber  = "blockNumber"
	paramStartBlock   = "startBlock"
	paramLimit        = "limit"
	paramIncludeEmpty = "includeEmpty"
	paramTxHash       = "txHash"
	paramStartID      = "startID"
	paramUnitID       = "unitID"

	blockNumberLatest = "latest"

	defaultBlocksPageLimit = 10
	defaultTxsPageLimit    = 20
)

type (
	ExplorerBackendService interface {
		GetRoundNumber(ctx context.Context) ([]service.PartitionRoundInfo, error)

		//block
		GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error)
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
		GetBlocksInRange(
			ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool,
		) (res []*domain.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxByHash(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error)
		GetTxsPage(
			ctx context.Context, partitionID types.PartitionID, startID string, limit int,
		) (transactions []*domain.TxInfo, previousID string, err error)
		FindTxs(ctx context.Context, searchKey []byte) ([]*domain.TxInfo, error)

		//bill
		GetBillsByPubKey(ctx context.Context, ownerID hex.Bytes) (res []*wallettypes.Bill, err error)
	}

	RestAPI struct {
		Service ExplorerBackendService
		rw      *ResponseWriter
	}

	RoundNumberResponse []service.PartitionRoundInfo

	SearchResponse struct {
		Blocks map[types.PartitionID]BlockInfo
		Txs    []TxInfo
	}

	BlockResponse map[types.PartitionID]BlockInfo

	BlockInfo struct {
		PartitionID        types.PartitionID
		PartitionTypeID    types.PartitionTypeID
		ShardID            types.ShardID
		ProposerID         string
		PreviousBlockHash  hex.Bytes
		TxHashes           []domain.TxHash
		UnicityCertificate types.TaggedCBOR
		BlockNumber        uint64
	}

	TxInfo struct {
		TxRecordHash domain.TxHash
		TxOrderHash  domain.TxHash
		BlockNumber  uint64
		Transaction  *types.TransactionRecord
		PartitionID  types.PartitionID
	}
)

func (api *RestAPI) roundNumberFunc(w http.ResponseWriter, r *http.Request) {
	roundInfos, err := api.Service.GetRoundNumber(r.Context())
	if err != nil {
		println("GET /round-number error fetching round number", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		api.rw.WriteResponse(w, roundInfos)
	}
}

func (api *RestAPI) search(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()

	searchKey := qp.Get("q")
	if searchKey == "" {
		http.Error(w, "Empty search key", http.StatusBadRequest)
		return
	}

	var partitionIDs []types.PartitionID
	for _, pid := range qp[paramPartitionID] {
		id, err := strconv.ParseUint(pid, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid partitionID: %s", pid), http.StatusBadRequest)
			return
		}
		partitionIDs = append(partitionIDs, types.PartitionID(id))
	}

	result := SearchResponse{
		Blocks: map[types.PartitionID]BlockInfo{},
		Txs:    []TxInfo{},
	}

	blockNumber, err := strconv.ParseUint(searchKey, 10, 64)
	if err == nil {
		blockMap, err := api.Service.GetBlock(r.Context(), blockNumber, partitionIDs)
		if err == nil {
			if len(blockMap) == 0 {
				http.Error(w, fmt.Sprintf("no blocks found for number %d", blockNumber), http.StatusNotFound)
				return
			}
			for partitionID, block := range blockMap {
				result.Blocks[partitionID] = blockInfoResponse(block)
			}
			api.rw.WriteResponse(w, result)
			return
		} else {
			fmt.Printf("Error getting block by number (%d): %s\n", blockNumber, err)
		}
	}

	hashBytes, err := ParseHex[[]byte](searchKey, true)
	if err == nil {
		txs, err := api.Service.FindTxs(r.Context(), hashBytes)
		if err == nil {
			if len(txs) > 0 {
				for _, txInfo := range txs {
					result.Txs = append(result.Txs, TxInfo{
						TxRecordHash: txInfo.TxRecordHash,
						TxOrderHash:  txInfo.TxOrderHash,
						BlockNumber:  txInfo.BlockNumber,
						Transaction:  txInfo.Transaction,
						PartitionID:  txInfo.PartitionID,
					})
				}
				api.rw.WriteResponse(w, result)
				return
			}
		} else {
			fmt.Printf("Error finding transactions: %s\n", err)
		}
	}

	http.Error(w, fmt.Sprintf("no results found for '%s'", searchKey), http.StatusNotFound)
}
