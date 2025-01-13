package restapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
)

const (
	paramIncludeDcBills = "includeDcBills"
	paramPubKey         = "pubkey"
)

type (
	ExplorerBackendService interface {
		//GetRoundNumber(ctx context.Context) (uint64, error)
		//GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)

		//block
		GetLastBlock(ctx context.Context, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
		GetBlock(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*api.BlockInfo, error)
		GetBlocksInRange(ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(ctx context.Context, txHash api.TxHash) (res *api.TxInfo, err error)
		GetTxsByBlockNumber(ctx context.Context, blockNumber uint64, partitionID types.PartitionID) ([]*api.TxInfo, error)
		GetTxsByUnitID(ctx context.Context, unitID types.UnitID) ([]*api.TxInfo, error)
		GetTxsPage(
			ctx context.Context,
			partitionID types.PartitionID,
			startID string,
			limit int,
		) (transactions []*api.TxInfo, previousID string, err error)

		//bill
		//GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
	}

	MoneyRestAPI struct {
		Service            ExplorerBackendService
		ListBillsPageLimit int
		rw                 *ResponseWriter
	}

	RoundNumberResponse struct {
		RoundNumber uint64 `json:"roundNumber,string"`
	}
)

/*func (api *MoneyRestAPI) roundNumberFunc(w http.ResponseWriter, r *http.Request) {
	lastRoundNumber, err := api.Service.GetRoundNumber(r.Context())
	if err != nil {
		println("GET /round-number error fetching round number", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		api.rw.WriteResponse(w, &RoundNumberResponse{RoundNumber: lastRoundNumber})
	}
}*/

func (api *MoneyRestAPI) getInfo(w http.ResponseWriter, _ *http.Request) {
	res := InfoResponse{
		Name: "blocks backend",
	}
	api.rw.WriteResponse(w, res)
}

func parsePubKeyQueryParam(r *http.Request) (api.PubKey, error) {
	return DecodePubKeyHex(r.URL.Query().Get(paramPubKey))
}

func parseIncludeDCBillsQueryParam(r *http.Request, defaultValue bool) (bool, error) {
	if r.URL.Query().Has(paramIncludeDcBills) {
		return strconv.ParseBool(r.URL.Query().Get(paramIncludeDcBills))
	}
	return defaultValue, nil
}
