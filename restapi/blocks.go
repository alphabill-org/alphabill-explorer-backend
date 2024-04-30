package restapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// @Summary Retrieve a block by block number or the latest block
// @Description Retrieves a block using the block number provided as a path parameter or retrieves the latest block if no block number is specified.
// @Tags Blocks
// @Accept json
// @Produce json
// @Param blockNumber path int false "The block number to retrieve (optional; if not provided, the latest block is returned)"
// @Success 200 {object} api.BlockInfo "Successfully retrieved the block"
// @Failure 400 {object} string "Invalid block number format or other client error"
// @Failure 404 {object} string "Block with the specified block number not found"
// @Failure 500 {object} string "Internal server error, such as failure to load the last block number or to load the block from the service"
// @Router /blocks/{blockNumber} [get]
func (api *MoneyRestAPI) getBlock(w http.ResponseWriter, r *http.Request) {

	var blockNumber uint64

	vars := mux.Vars(r)
	blockNumberStr, ok := vars["blockNumber"]

	if ok {
		var err error
		blockNumber, err = strconv.ParseUint(blockNumberStr, 10, 64)
		if err != nil {
			api.rw.ErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid 'blockNumber' format: %v", err))
			return
		}
	} else {
		var err error
		blockNumber, err = api.Service.GetLastBlockNumber()
		if err != nil {
			api.rw.ErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to load last block number: %v", err))
			return
		}
	}

	block, err := api.Service.GetBlock(blockNumber)
	if err != nil {
		api.rw.ErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to load block with block number %d: %v", blockNumber, err))
		return
	}

	if block == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("block with block number %d not found", blockNumber))
		return
	}

	api.rw.WriteResponse(w, block)
}

// @Summary Get blocks, given a start block number and limit.
// @Description Get blocks, given a start block number and limit.
// @produce	application/json
// @Param startBlock query string false "optionally specify the start block number"
// @Param limit query string false "optionally specify the number of blocks to return, defaults to 10"
// @Success 200 {array} api.BlockInfo
// @Router /blocks [get]
// @Tags Blocks
func (api *MoneyRestAPI) getBlocks(w http.ResponseWriter, r *http.Request) {

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
