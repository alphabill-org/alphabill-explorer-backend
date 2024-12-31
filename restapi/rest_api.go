package restapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-go-base/types/hex"
)

const (
	paramIncludeDcBills = "includeDcBills"
	paramPubKey         = "pubkey"
)

type (
	ExplorerBackendService interface {
		GetRoundNumber(ctx context.Context) (uint64, error)
		GetUnitsByOwnerID(ctx context.Context, ownerID hex.Bytes) ([]types.UnitID, error)

		//block
		GetLastBlockNumber() (uint64, error)
		GetBlock(blockNumber uint64) (*api.BlockInfo, error)
		GetBlocks(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error)

		//tx
		GetTxInfo(txHash api.TxHash) (res *api.TxInfo, err error)
		GetBlockTxsByBlockNumber(blockNumber uint64) (res []*api.TxInfo, err error)
		GetTxsByUnitID(unitID types.UnitID) ([]*api.TxInfo, error)
		GetTxs(startSequenceNumber uint64, count int) (res []*api.TxInfo, prevSequenceNumber uint64, err error)

		//bill
		//GetBillsByPubKey(ctx context.Context, ownerID types.Bytes) (res []*moneyApi.Bill, err error)
	}

	MoneyRestAPI struct {
		Service            ExplorerBackendService
		ListBillsPageLimit int
		rw                 *ResponseWriter
		PartitionID        types.PartitionID
	}

	RoundNumberResponse struct {
		RoundNumber uint64 `json:"roundNumber,string"`
	}
)

func (api *MoneyRestAPI) roundNumberFunc(w http.ResponseWriter, r *http.Request) {
	lastRoundNumber, err := api.Service.GetRoundNumber(r.Context())
	if err != nil {
		println("GET /round-number error fetching round number", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		api.rw.WriteResponse(w, &RoundNumberResponse{RoundNumber: lastRoundNumber})
	}
}

func (api *MoneyRestAPI) getInfo(w http.ResponseWriter, _ *http.Request) {
	res := InfoResponse{
		PartitionID: api.PartitionID,
		Name:        "blocks backend",
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
