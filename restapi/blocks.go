package restapi

import (
	"fmt"
	exTypes "github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// @Summary Retrieve a blockchain block by number, or the latest if unspecified
// @Description Retrieves a block by using the provided block number as a path parameter, or retrieves the latest block if no number is specified.
// @Tags Blocks
// @Accept json
// @Produce json
// @Param blockNumber path string false "Block number ('latest' or a specific number)"
// @Success 200 {object} api.BlockInfo "Block information successfully retrieved"
// @Failure 400 {object} string "Missing or invalid block number"
// @Failure 404 {object} string "No block found with the specified number"
// @Failure 500 {object} string "Internal server error, such as a failure to retrieve the block"
// @Router /blocks/{blockNumber} [get]
func (api *MoneyRestAPI) getBlock(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()
	var partitionIDs []types.PartitionID
	for _, pid := range qp["partitionID"] {
		id, err := strconv.ParseUint(pid, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid partitionID: %s", pid), http.StatusBadRequest)
			return
		}
		partitionIDs = append(partitionIDs, types.PartitionID(id))
	}

	var err error

	vars := mux.Vars(r)
	blockNumberStr, ok := vars["blockNumber"]
	if !ok {
		http.Error(w, "Missing 'blockNumber' variable in the URL", http.StatusBadRequest)
		return
	}

	var blockMap map[types.PartitionID]*exTypes.BlockInfo
	if blockNumberStr == "latest" {
		blockMap, err = api.Service.GetLastBlock(r.Context(), partitionIDs)
		if err != nil {
			api.rw.ErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to get latest blocks: %w", err))
			return
		}
		api.rw.WriteResponse(w, blockMap)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid blockNumber: %s", blockNumberStr), http.StatusBadRequest)
		return
	}

	blockMap, err = api.Service.GetBlock(r.Context(), blockNumber, partitionIDs)
	if err != nil {
		api.rw.ErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to load block with block number %d: %w", blockNumber, err))
		return
	}

	if blockMap == nil || len(blockMap) == 0 {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("block with block number %d not found", blockNumber))
		return
	}

	api.rw.WriteResponse(w, blockMap)
}

// @Summary Get blocks, given a start block number and limit.
// @Description Get blocks, given a start block number and limit.
// @produce	application/json
// @Param startBlock query string false "optionally specify the start block number"
// @Param limit query string false "optionally specify the number of blocks to return, defaults to 10"
// @Success 200 {array} api.BlockInfo
// @Router /blocks [get]
// @Tags Blocks
func (api *MoneyRestAPI) getBlocksInRange(w http.ResponseWriter, r *http.Request) {
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

	qp := r.URL.Query()

	startBlockStr := qp.Get("startBlock")
	limitStr := qp.Get("limit")

	limit := 10
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid 'limit' format", http.StatusBadRequest)
			return
		}
	}

	var startBlock uint64
	if startBlockStr != "" {
		startBlock, err = strconv.ParseUint(startBlockStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid 'startBlock' format", http.StatusBadRequest)
			return
		}
	} else {
		lastBlock, err := api.Service.GetLastBlock(r.Context(), []types.PartitionID{types.PartitionID(partitionID)})
		if err != nil {
			http.Error(w, "unable to get last block number", http.StatusBadRequest)
			return
		}
		startBlock = lastBlock[types.PartitionID(partitionID)].BlockNumber
	}

	recs, prevBlockNumber, err := api.Service.GetBlocksInRange(r.Context(), types.PartitionID(partitionID), startBlock, limit)
	if err != nil {
		println("error on GET /blocks: ", err)
		api.rw.WriteErrorResponse(w, fmt.Errorf("unable to fetch blocks: %w", err))
		return
	}
	prevBlockNumberStr := strconv.FormatUint(prevBlockNumber, 10)
	setLinkHeader(r.URL, w, prevBlockNumberStr)
	api.rw.WriteResponse(w, recs)
}
