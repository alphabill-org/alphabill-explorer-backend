package restapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/service"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
)

const (
	paramIncludeDcBills = "includeDcBills"
	paramPubKey         = "pubkey"
	paramPartitionID    = "partitionID"
	paramBlockNumber    = "blockNumber"
	paramStartBlock     = "startBlock"
	paramLimit          = "limit"
	paramIncludeEmpty   = "includeEmpty"
	paramTxHash         = "txHash"
	paramStartID        = "startID"
	paramUnitID         = "unitID"

	blockNumberLatest = "latest"

	defaultBlocksPageLimit = 10
	defaultTxsPageLimit    = 20
)

type (
	ExplorerBackendService interface {
		GetRoundNumber(ctx context.Context) ([]service.PartitionRoundInfo, error)
		//GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)

		//block
		GetLastBlocks(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error)
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error)
		GetBlocksInRange(
			ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool,
		) (res []*domain.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(ctx context.Context, txHash domain.TxHash) (res *domain.TxInfo, err error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*domain.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*domain.TxInfo, error)
		GetTxsPage(
			ctx context.Context, partitionID types.PartitionID, startID string, limit int,
		) (transactions []*domain.TxInfo, previousID string, err error)

		//bill
		//GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
	}

	RestAPI struct {
		Service ExplorerBackendService
		rw      *ResponseWriter
	}

	RoundNumberResponse []service.PartitionRoundInfo

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

func (api *RestAPI) getInfo(w http.ResponseWriter, _ *http.Request) {
	res := InfoResponse{
		Name: "blocks backend",
	}
	api.rw.WriteResponse(w, res)
}

func parsePubKeyQueryParam(r *http.Request) (domain.PubKey, error) {
	return DecodePubKeyHex(r.URL.Query().Get(paramPubKey))
}

func parseIncludeDCBillsQueryParam(r *http.Request, defaultValue bool) (bool, error) {
	if r.URL.Query().Has(paramIncludeDcBills) {
		return strconv.ParseBool(r.URL.Query().Get(paramIncludeDcBills))
	}
	return defaultValue, nil
}
