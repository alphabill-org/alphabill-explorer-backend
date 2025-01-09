package restapi

import (
	"fmt"
	"github.com/alphabill-org/alphabill-go-base/types"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/util"
	"github.com/gorilla/mux"
)

// @Summary Retrieve a transaction by hash
// @Description Retrieves transaction details using a transaction hash provided as a path parameter.
// @Tags Transactions
// @Accept json
// @Produce json
// @Param txHash path string true "The hash of the transaction to retrieve (HEX encoded)"
// @Success 200 {object} api.TxInfo "Successfully retrieved the transaction information"
// @Failure 400 {string} string "Missing 'txHash' variable in the URL"
// @Failure 404 {string} string "Transaction with the specified hash not found"
// @Failure 500 {string} string "Failed to load transaction details"
// @Router /txs/{txHash} [get]
func (api *MoneyRestAPI) getTx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash, ok := vars["txHash"]
	if !ok {
		http.Error(w, "Missing 'txHash' variable in the URL", http.StatusBadRequest)
		return
	}
	txHashBytes, err := ParseHex[[]byte](txHash, true)
	if err != nil {
		http.Error(w, "Invalid 'txHash' format", http.StatusBadRequest)
	}
	txInfo, err := api.Service.GetTxInfo(r.Context(), txHashBytes)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load tx with txHash %s : %w", txHash, err))
		return
	}

	if txInfo == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("tx with txHash %s not found", txHash))
		return
	}
	api.rw.WriteResponse(w, txInfo)
}

// @Summary Retrieve transactions, latest first.
// @Description Retrieves a list of transactions.
// @Tags Transactions
// @Produce json
// @Param startSeqNumber query string false "The sequence number of the transaction to start from, if not provided, the latest transactions are returned"
// @Param limit query int false "The maximum number of transactions to retrieve, default 20"
// @Success 200 {array} api.TxInfo "Successfully retrieved list of transactions"
// @Router /txs [get]
func (api *MoneyRestAPI) getTxs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partitionIDStr, ok := vars["partitionID"]
	if !ok {
		http.Error(w, "Missing 'partitionID' variable in the URL", http.StatusBadRequest)
		return
	}
	partitionID, err := strconv.ParseUint(partitionIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid partitionID: %s", partitionIDStr), http.StatusBadRequest)
		return
	}

	seqStr := r.URL.Query().Get("startSeqNumber")
	seq := uint64(0)
	if seqStr != "" {
		seq, err = strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			api.rw.InvalidParamResponse(w, "startSeqNumber", err)
		}
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		limit, err = ParseMaxResponseItems(limitStr, 100)
		if err != nil {
			http.Error(w, "Invalid 'limit' format", http.StatusBadRequest)
			return
		}
	}

	txs, prev, err := api.Service.GetTxsInRange(r.Context(), types.PartitionID(partitionID), seq, limit)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load txs with startSeqNumber %d and limit %d : %w", seq, limit, err))
		return
	}

	setLinkHeader(r.URL, w, fmt.Sprintf("%d", prev))
	api.rw.WriteResponse(w, txs)
}

// @Summary Retrieve transactions by block number
// @Description Retrieves a list of transactions for a given block number.
// @Tags Transactions
// @Accept json
// @Produce json
// @Param blockNumber path int true "The block number for which to retrieve transactions"
// @Success 200 {array} api.TxInfo "Successfully retrieved list of transactions for the block"
// @Failure 400 {string} string "Missing or invalid 'blockNumber' variable in the URL"
// @Failure 404 {string} string "No transactions found for the specified block number"
// @Router /blocks/{blockNumber}/txs [get]
func (api *MoneyRestAPI) getBlockTxsByBlockNumber(w http.ResponseWriter, r *http.Request) {
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
	partitionIDStr, ok := vars["partitionID"]
	if !ok {
		http.Error(w, "Missing 'partitionID' variable in the URL", http.StatusBadRequest)
		return
	}
	partitionID, err := strconv.ParseUint(partitionIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid partitionID: %s", partitionIDStr), http.StatusBadRequest)
		return
	}

	txs, err := api.Service.GetTxsByBlockNumber(r.Context(), blockNumber, types.PartitionID(partitionID))
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load txs with blockNumber %d : %w", blockNumber, err))
		return
	}

	if txs == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("tx with txHash %d not found", blockNumber))
		return
	}
	api.rw.WriteResponse(w, txs)
}

// @Summary Retrieve transactions by unit ID
// @Description Get transactions associated with a specific unit ID
// @Tags Transactions
// @Accept json
// @Produce json
// @Param unitID path string true "Unit ID (0xHEX encoded)"
// @Success 200 {array} api.TxInfo "List of transactions"
// @Failure 400 {object} ErrorResponse "Error: Missing 'unitID' variable in the URL"
// @Failure 404 {object} ErrorResponse "Error: Transaction with specified unit ID not found"
// @Router /units/{unitID}/txs [get]
func (api *MoneyRestAPI) getTxsByUnitID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	unitID, ok := vars["unitID"]
	if !ok {
		http.Error(w, "Missing 'unitID' variable in the URL", http.StatusBadRequest)
		return
	}

	uid, err := util.FromHex([]byte(unitID))
	if err != nil {
		http.Error(w, "Invalid 'unitID' format", http.StatusBadRequest)
		return
	}
	txs, err := api.Service.GetTxsByUnitID(r.Context(), uid)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load txs with unitID %s : %w", unitID, err))
		return
	}

	if txs == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("tx with unitID %s not found", unitID))
		return
	}
	api.rw.WriteResponse(w, txs)
}
