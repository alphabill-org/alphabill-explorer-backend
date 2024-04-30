package restapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// @Summary Retrieve a transaction by hash
// @Description Retrieves transaction details using a transaction hash provided as a path parameter.
// @Tags Transactions
// @Accept json
// @Produce json
// @Param txHash path string true "The hash of the transaction to retrieve"
// @Success 200 {object} types.TxInfo "Successfully retrieved the transaction information"
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
	txInfo, err := api.Service.GetTxInfo(txHash)
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

// @Summary Retrieve transactions by block number
// @Description Retrieves a list of transactions for a given block number.
// @Tags Blocks
// @Accept json
// @Produce json
// @Param blockNumber path int true "The block number for which to retrieve transactions"
// @Success 200 {array} types.TxInfo "Successfully retrieved list of transactions for the block"
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

	txs, err := api.Service.GetBlockTxsByBlockNumber(blockNumber)
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

func (api *MoneyRestAPI) getTxsByUnitID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	unitID, ok := vars["unitID"]
	if !ok {
		http.Error(w, "Missing 'unitID' variable in the URL", http.StatusBadRequest)
		return
	}

	txs, err := api.Service.GetTxsByUnitID(unitID)
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
