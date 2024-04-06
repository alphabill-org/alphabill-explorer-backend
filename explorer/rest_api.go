package explorer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sdk "github.com/alphabill-org/alphabill-wallet/wallet"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	moneyapi "github.com/alphabill-org/alphabill-wallet/wallet/money/api"
	"github.com/alphabill-org/alphabill/txsystem/money"
	"github.com/alphabill-org/alphabill/types"
)

const (
	paramIncludeDcBills = "includeDcBills"
	paramPubKey         = "pubkey"
)

type (
	moneyRestAPI struct {
		Service            ExplorerBackendService
		ListBillsPageLimit int
		rw                 *ResponseWriter
		SystemID           types.SystemID
	}

	ListBillsResponse struct {
		Bills []*moneyapi.Bill `json:"bills"`
	}

	BalanceResponse struct {
		Balance uint64 `json:"balance,string"`
	}

	AddKeyRequest struct {
		Pubkey string `json:"pubkey"`
	}

	RoundNumberResponse struct {
		RoundNumber uint64 `json:"roundNumber,string"`
	}
)

var (
	errInvalidBillIDLength = errors.New("bill_id hex string must be 68 characters long (with 0x prefix)")
)

func (api *moneyRestAPI) Router() *mux.Router {
	// TODO add request/response headers middleware
	router := mux.NewRouter().StrictSlash(true)

	apiRouter := router.PathPrefix("/api").Subrouter()
	// add cors middleware
	// content-type needs to be explicitly defined without this content-type header is not allowed and cors filter is not applied
	// Link header is needed for pagination support.
	// OPTIONS method needs to be explicitly defined for each handler func
	apiRouter.Use(handlers.CORS(
		handlers.AllowedHeaders([]string{ContentType}),
		handlers.ExposedHeaders([]string{HeaderLink}),
	))

	// version v1 router
	apiV1 := apiRouter.PathPrefix("/v1").Subrouter()
	apiV1.HandleFunc("/blocks/{blockNumber}", api.getBlockByBlockNumber).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocks", api.getBlocks).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocksExp/{blockNumber}", api.getBlockExplorerByBlockNumber).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocksExp", api.getBlocksExplorer).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocksExp/{blockNumber}/txsExp/", api.getBlockExplorerTxsByBlockNumber).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/txsExp/{txHash}", api.getTxExplorerByTxHash).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/tx-history", api.getTxHistory).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/tx-history/{pubkey}", api.getTxHistoryByKey).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/units/{unitId}/transactions/{txHash}/proof", api.getTxProof).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/round-number", api.roundNumberFunc).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/info", api.getInfo).Methods("GET", "OPTIONS")

	return router
}

func (api *moneyRestAPI) getBlockByBlockNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockNumberStr, ok := vars["blockNumber"]
	if !ok {
		http.Error(w, "Missing 'blockNumber' variable in the URL", http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'blockNumber' format", http.StatusBadRequest)
		return
	}

	block, err := api.Service.GetBlockByBlockNumber(blockNumber)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load block with block number %d : %w", blockNumber, err))
		return
	}

	if block == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("block with block number %x not found", blockNumber))
		return
	}

	api.rw.WriteResponse(w, block)
}

func (api *moneyRestAPI) getBlocks(w http.ResponseWriter, r *http.Request) {

	qp := r.URL.Query()

	startBlockStr := qp.Get("startBlock")
	limitStr := qp.Get("limit")

	startBlock, err := strconv.ParseUint(startBlockStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'startBlock' format", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Invalid 'limit' format", http.StatusBadRequest)
		return
	}

	recs, prevBlockNumber, err := api.Service.GetBlocks(startBlock, limit)
	if err != nil {
		println("error on GET /blocks: ", err)
		api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch blocks: %w", err))
		return
	}
	prevBlockNumberStr := strconv.FormatUint(prevBlockNumber, 10)
	SetLinkHeader(r.URL, w, prevBlockNumberStr)
	api.rw.WriteResponse(w, recs)
}
func (api *moneyRestAPI) getBlockExplorerByBlockNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockNumberStr, ok := vars["blockNumber"]
	if !ok {
		http.Error(w, "Missing 'blockNumber' variable in the URL", http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'blockNumber' format", http.StatusBadRequest)
		return
	}

	block, err := api.Service.GetBlockExplorerByBlockNumber(blockNumber)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load block with block number %d : %w", blockNumber, err))
		return
	}

	if block == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("block with block number %d not found", blockNumber))
		return
	}

	api.rw.WriteResponse(w, block)
}
func (api *moneyRestAPI) getBlocksExplorer(w http.ResponseWriter, r *http.Request) {

	qp := r.URL.Query()

	startBlockStr := qp.Get("startBlock")
	limitStr := qp.Get("limit")

	startBlock, err := api.Service.GetLastBlockNumber()
	if err != nil {
		http.Error(w, "unable to get last block number", http.StatusBadRequest)
		return
	}

	if startBlockStr != "" {
		startBlock, err = strconv.ParseUint(startBlockStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid 'startBlock' format", http.StatusBadRequest)
			return
		}
	}

	limit := 10

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid 'limit' format", http.StatusBadRequest)
			return
		}
	}

	recs, prevBlockNumber, err := api.Service.GetBlocksExplorer(startBlock, limit)
	if err != nil {
		println("error on GET /blocks: ", err)
		api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch blocks: %w", err))
		return
	}
	prevBlockNumberStr := strconv.FormatUint(prevBlockNumber, 10)
	SetLinkHeader(r.URL, w, prevBlockNumberStr)
	api.rw.WriteResponse(w, recs)
}

func (api *moneyRestAPI) getTxExplorerByTxHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash, ok := vars["txHash"]
	if !ok {
		http.Error(w, "Missing 'txHash' variable in the URL", http.StatusBadRequest)
		return
	}
	txExplorer, err := api.Service.GetTxExplorerByTxHash(txHash)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load tx with txHash %s : %w", txHash, err))
		return
	}

	if txExplorer == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("tx with txHash %x not found", txHash))
		return
	}
	api.rw.WriteResponse(w, txExplorer)
}

func (api *moneyRestAPI) getBlockExplorerTxsByBlockNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockNumberStr, ok := vars["blockNumber"]
	if !ok {
		http.Error(w, "Missing 'blockNumber' variable in the URL", http.StatusBadRequest)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid 'blockNumber' format", http.StatusBadRequest)
		return
	}

	txsExplorer, err := api.Service.GetBlockExplorerTxsByBlockNumber(blockNumber)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load txs with blockNumber %d : %w", blockNumber, err))
		return
	}

	if txsExplorer == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("tx with txHash %d not found", blockNumber))
		return
	}
	api.rw.WriteResponse(w, txsExplorer)
}

func (api *moneyRestAPI) getTxHistory(w http.ResponseWriter, r *http.Request) {

	//qp := r.URL.Query()
	//startKey, err := ParseHex[[]byte](qp.Get(QueryParamOffsetKey), false)
	//if err != nil {
	//	api.rw.InvalidParamResponse(w, QueryParamOffsetKey, err)
	//	return
	//}
	//
	//limit, err := ParseMaxResponseItems(qp.Get(QueryParamLimit), api.ListBillsPageLimit)
	//if err != nil {
	//	api.rw.InvalidParamResponse(w, QueryParamLimit, err)
	//	return
	//}
	//recs, nextKey, err := api.Service.GetTxHistoryRecords(startKey, limit)
	//if err != nil {
	//	println("error on GET /tx-history: ", err)
	//	api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch tx history records: %w", err))
	//	return
	//}
	//// check if unconfirmed tx-s are now confirmed or failed
	//var roundNr uint64 = 0
	//for _, rec := range recs {
	//	// TODO: update db if stage changes to confirmed or failed
	//	if rec.State == UNCONFIRMED {
	//		proof, err := api.Service.GetTxProof(rec.UnitID, rec.TxHash)
	//		if err != nil {
	//			api.rw.WriteErrorResponse(w, fmt.Errorf("failed to fetch tx proof: %w", err))
	//		}
	//		if proof != nil {
	//			rec.State = CONFIRMED
	//		} else {
	//			if roundNr == 0 {
	//				roundNr, err = api.Service.GetRoundNumber(r.Context())
	//				if err != nil {
	//					api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch latest round number: %w", err))
	//				}
	//			}
	//			if roundNr > rec.Timeout {
	//				rec.State = FAILED
	//			}
	//		}
	//	}
	//}
	//SetLinkHeader(r.URL, w, EncodeHex(nextKey))
	//api.rw.WriteCborResponse(w, recs)
}

func (api *moneyRestAPI) getTxHistoryByKey(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//senderPubkey, err := DecodePubKeyHex(vars["pubkey"])
	//if err != nil {
	//	api.rw.InvalidParamResponse(w, "pubkey", fmt.Errorf("failed to parse sender pubkey: %w", err))
	//	return
	//}
	//qp := r.URL.Query()
	//startKey, err := ParseHex[[]byte](qp.Get(QueryParamOffsetKey), false)
	//if err != nil {
	//	api.rw.InvalidParamResponse(w, QueryParamOffsetKey, err)
	//	return
	//}
	//
	//limit, err := ParseMaxResponseItems(qp.Get(QueryParamLimit), api.ListBillsPageLimit)
	//if err != nil {
	//	api.rw.InvalidParamResponse(w, QueryParamLimit, err)
	//	return
	//}
	//recs, nextKey, err := api.Service.GetTxHistoryRecordsByKey(senderPubkey.Hash(), startKey, limit)
	//if err != nil {
	//	println("error on GET /tx-history: ", err)
	//	api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch tx history records: %w", err))
	//	return
	//}
	//// check if unconfirmed tx-s are now confirmed or failed
	//var roundNr uint64 = 0
	//for _, rec := range recs {
	//	// TODO: update db if stage changes to confirmed or failed
	//	if rec.State == UNCONFIRMED {
	//		proof, err := api.Service.GetTxProof(rec.UnitID, rec.TxHash)
	//		if err != nil {
	//			api.rw.WriteErrorResponse(w, fmt.Errorf("failed to fetch tx proof: %w", err))
	//		}
	//		if proof != nil {
	//			rec.State = CONFIRMED
	//		} else {
	//			if roundNr == 0 {
	//				roundNr, err = api.Service.GetRoundNumber(r.Context())
	//				if err != nil {
	//					api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch latest round number: %w", err))
	//				}
	//			}
	//			if roundNr > rec.Timeout {
	//				rec.State = FAILED
	//			}
	//		}
	//	}
	//}
	//SetLinkHeader(r.URL, w, EncodeHex(nextKey))
	//api.rw.WriteCborResponse(w, recs)
}

func (api *moneyRestAPI) getTxProof(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	unitID, err := ParseHex[types.UnitID](vars["unitId"], true)
	if err != nil {
		api.rw.InvalidParamResponse(w, "unitId", err)
		return
	}
	if len(unitID) != money.UnitIDLength {
		api.rw.ErrorResponse(w, http.StatusBadRequest, errInvalidBillIDLength)
		return
	}
	txHash, err := ParseHex[TxHash](vars["txHash"], true)
	if err != nil {
		api.rw.InvalidParamResponse(w, "txHash", err)
		return
	}

	proof, err := api.Service.GetTxProof(unitID, sdk.TxHash(txHash))
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load proof of tx 0x%X (unit 0x%s): %w", txHash, unitID, err))
		return
	}
	if proof == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("no proof found for tx 0x%X (unit 0x%s)", txHash, unitID))
		return
	}

	api.rw.WriteCborResponse(w, proof)
}

func (api *moneyRestAPI) roundNumberFunc(w http.ResponseWriter, r *http.Request) {
	lastRoundNumber, err := api.Service.GetRoundNumber(r.Context())
	if err != nil {
		println("GET /round-number error fetching round number", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		api.rw.WriteResponse(w, &RoundNumberResponse{RoundNumber: lastRoundNumber})
	}
}

func (api *moneyRestAPI) getInfo(w http.ResponseWriter, _ *http.Request) {
	res := InfoResponse{
		SystemID: api.SystemID,
		Name:     "explorer backend",
	}
	api.rw.WriteResponse(w, res)
}

func parsePubKeyQueryParam(r *http.Request) (PubKey, error) {
	return DecodePubKeyHex(r.URL.Query().Get(paramPubKey))
}

func parseIncludeDCBillsQueryParam(r *http.Request, defaultValue bool) (bool, error) {
	if r.URL.Query().Has(paramIncludeDcBills) {
		return strconv.ParseBool(r.URL.Query().Get(paramIncludeDcBills))
	}
	return defaultValue, nil
}
