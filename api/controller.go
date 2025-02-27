package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/service/partition"
	"github.com/alphabill-org/alphabill-explorer-backend/service/search"
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
	paramSearchKey    = "q"
	paramPubKey       = "pubKey"

	blockNumberLatest = "latest"

	defaultBlocksPageLimit = 10
	defaultTxsPageLimit    = 20
)

type (
	StorageService interface {
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
		FindTxs(ctx context.Context, searchKey []byte, partitionIDs []types.PartitionID) ([]*domain.TxInfo, error)
	}

	PartitionService interface {
		GetRoundNumber(ctx context.Context) ([]partition.RoundInfo, error)
	}

	MoneyService interface {
		GetBillsByPubKey(ctx context.Context, ownerID hex.Bytes) ([]*wallettypes.Bill, error)
	}

	SearchService interface {
		Search(ctx context.Context, searchKey string, partitionIDs []types.PartitionID) (*search.Result, error)
	}

	Controller struct {
		StorageService   StorageService
		PartitionService PartitionService
		MoneyService     MoneyService
		SearchService    SearchService
		rw               *ResponseWriter
	}

	RoundNumberResponse []partition.RoundInfo

	SearchResponse struct {
		Blocks map[types.PartitionID]BlockInfo
		Txs    []TxInfo
		Units  map[types.PartitionID][]types.UnitID
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

func NewController(
	StorageService StorageService,
	PartitionService PartitionService,
	MoneyService MoneyService,
	searchService SearchService,
) (*Controller, error) {
	if StorageService == nil {
		return nil, errors.New("storage service is nil")
	}
	if PartitionService == nil {
		return nil, errors.New("partition service is nil")
	}
	if MoneyService == nil {
		return nil, errors.New("money service is nil")
	}
	if searchService == nil {
		return nil, errors.New("search service is nil")
	}

	return &Controller{
		StorageService:   StorageService,
		PartitionService: PartitionService,
		MoneyService:     MoneyService,
		SearchService:    searchService,
		rw:               &ResponseWriter{},
	}, nil
}

// @Summary Retrieve round and epoch number for each partition
// @Description Retrieve round and epoch number for each partition
// @Tags Info
// @Produce json
// @Success 200 {array} partition.RoundInfo
// @Router /round-number [get]
func (c *Controller) roundNumber(w http.ResponseWriter, r *http.Request) {
	roundInfos, err := c.PartitionService.GetRoundNumber(r.Context())
	if err != nil {
		c.rw.WriteInternalErrorResponse(w, err)
		return
	}
	c.rw.WriteResponse(w, roundInfos)
}

// @Summary Retrieve blocks and transactions matching the search key
// @Description Retrieve blocks and transactions matching the search key
// @Tags Search
// @Produce json
// @Param q query string true "Search key"
// @Param partitionID query int false "Filter results by partition ID(s)"
// @Success 200 {object} SearchResponse "Block information successfully retrieved"
// @Failure 400 {object} string "Empty search key"
// @Failure 400 {object} string "invalid partitionID"
// @Failure 404 {object} string "no results found"
// @Router /search [get]
func (c *Controller) search(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()
	searchKey := qp.Get(paramSearchKey)
	if searchKey == "" {
		c.rw.WriteMissingParamResponse(w, paramSearchKey)
		return
	}

	var partitionIDs []types.PartitionID
	for _, pid := range qp[paramPartitionID] {
		id, err := strconv.ParseUint(pid, 10, 64)
		if err != nil {
			c.rw.WriteInvalidParamResponse(w, paramPartitionID)
			return
		}
		partitionIDs = append(partitionIDs, types.PartitionID(id))
	}

	result, err := c.SearchService.Search(r.Context(), searchKey, partitionIDs)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrFailedToDecodeHex) {
			c.rw.WriteErrorResponse(w, fmt.Errorf("no results found for '%s'", searchKey), http.StatusNotFound)
			return
		}
		c.rw.WriteInternalErrorResponse(w, err)
		return
	}

	if len(result.Txs) == 0 && len(result.Blocks) == 0 && len(result.Units) == 0 {
		c.rw.WriteErrorResponse(w, fmt.Errorf("no results found for '%s'", searchKey), http.StatusNotFound)
		return
	}

	c.rw.WriteResponse(w, formatSearchResponse(result))
}

func formatSearchResponse(result *search.Result) SearchResponse {
	response := SearchResponse{
		Blocks: make(map[types.PartitionID]BlockInfo),
		Txs:    []TxInfo{},
		Units:  make(map[types.PartitionID][]types.UnitID),
	}

	for partitionID, block := range result.Blocks {
		response.Blocks[partitionID] = blockInfoResponse(block)
	}
	for _, tx := range result.Txs {
		response.Txs = append(response.Txs, txInfoResponse(tx))
	}
	for partitionID, unitIDs := range result.Units {
		if len(unitIDs) > 0 {
			response.Units[partitionID] = unitIDs
		}
	}
	return response
}
