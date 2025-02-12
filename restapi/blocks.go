package restapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/gorilla/mux"
)

// @Summary Retrieve a blockchain block by number, or the latest if unspecified
// @Description Retrieves a block for all given partitions by using the provided block number as a path parameter, or retrieves the latest block if no number is specified.
// @Tags Blocks
// @Accept json
// @Produce json
// @Param blockNumber path string true "Block number ('latest' or a specific number)"
// @Param partitionID query string false "List of partitions to get the blocks for. If not provided then get for all partitions"
// @Success 200 {object} BlockResponse "Block information successfully retrieved"
// @Failure 400 {object} string "invalid partitionID"
// @Failure 400 {object} string "Missing or invalid block number"
// @Failure 404 {object} string "No block found with the specified number"
// @Failure 500 {object} string "Internal server error, such as a failure to retrieve the block"
// @Router /blocks/{blockNumber} [get]
func (api *RestAPI) getBlock(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()
	var partitionIDs []types.PartitionID
	for _, pid := range qp[paramPartitionID] {
		id, err := strconv.ParseUint(pid, 10, 64)
		if err != nil {
			api.rw.WriteInvalidParamResponse(w, paramPartitionID)
			return
		}
		partitionIDs = append(partitionIDs, types.PartitionID(id))
	}

	var err error

	vars := mux.Vars(r)
	blockNumberStr, ok := vars[paramBlockNumber]
	if !ok {
		api.rw.WriteMissingParamResponse(w, paramBlockNumber)
		return
	}

	result := make(map[types.PartitionID]BlockInfo)
	if blockNumberStr == blockNumberLatest {
		blockMap, err := api.Service.GetLastBlocks(r.Context(), partitionIDs, 1, true)
		if err != nil {
			api.rw.WriteInternalErrorResponse(w, fmt.Errorf("failed to get latest blocks: %w", err))
			return
		}
		for partitionID, blocks := range blockMap {
			if len(blocks) > 0 {
				result[partitionID] = blockInfoResponse(blocks[0])
			}
		}
		api.rw.WriteResponse(w, result)
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		api.rw.WriteInvalidParamResponse(w, paramBlockNumber)
		return
	}

	blockMap, err := api.Service.GetBlock(r.Context(), blockNumber, partitionIDs)
	if err != nil {
		api.rw.WriteInternalErrorResponse(w, fmt.Errorf("failed to load block with block number %d: %w", blockNumber, err))
		return
	}

	if len(blockMap) == 0 {
		api.rw.WriteErrorResponse(w, fmt.Errorf("block with block number %d not found", blockNumber), http.StatusNotFound)
		return
	}

	for partitionID, block := range blockMap {
		result[partitionID] = blockInfoResponse(block)
	}

	api.rw.WriteResponse(w, result)
}

// @Summary Get blocks in a single partition, given a start block number and limit.
// @Description Get blocks in a single partition, given a start block number and limit.
// @produce	application/json
// @Param partitionID path string true "Partition ID to get the blocks for"
// @Param startBlock query string false "optionally specify the start block number"
// @Param limit query string false "optionally specify the number of blocks to return, defaults to 10"
// @Param includeEmpty query boolean false "whether to include blocks without transactions, defaults to true"
// @Success 200 {array} BlockInfo
// @Router /partitions/{partitionID}/blocks [get]
// @Tags Blocks
func (api *RestAPI) getBlocksInRange(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partitionIDStr, ok := vars[paramPartitionID]
	if !ok {
		api.rw.WriteMissingParamResponse(w, paramPartitionID)
		return
	}
	partitionIDUint, err := strconv.ParseUint(partitionIDStr, 10, 64)
	if err != nil {
		api.rw.WriteInvalidParamResponse(w, paramPartitionID)
		return
	}
	partitionID := types.PartitionID(partitionIDUint)

	qp := r.URL.Query()

	limitStr := qp.Get(paramLimit)
	limit := defaultBlocksPageLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			api.rw.WriteInvalidParamResponse(w, paramLimit)
			return
		}
	}

	includeEmptyStr := qp.Get(paramIncludeEmpty)
	includeEmpty := true
	if includeEmptyStr != "" {
		includeEmpty, err = strconv.ParseBool(includeEmptyStr)
		if err != nil {
			api.rw.WriteInvalidParamResponse(w, paramIncludeEmpty)
			return
		}
	}

	startBlockStr := qp.Get(paramStartBlock)
	var startBlock uint64
	if startBlockStr != "" {
		startBlock, err = strconv.ParseUint(startBlockStr, 10, 64)
		if err != nil {
			api.rw.WriteInvalidParamResponse(w, paramStartBlock)
			return
		}
	} else {
		lastBlocks, err := api.Service.GetLastBlocks(r.Context(), []types.PartitionID{partitionID}, limit, includeEmpty)
		if err != nil {
			api.rw.WriteInternalErrorResponse(w, err)
			return
		}
		var response = []BlockInfo{}
		for _, block := range lastBlocks[partitionID] {
			response = append(response, blockInfoResponse(block))
		}
		api.rw.WriteResponse(w, response)
		return
	}

	blocks, prevBlockNumber, err := api.Service.GetBlocksInRange(r.Context(), partitionID, startBlock, limit, includeEmpty)
	if err != nil {
		api.rw.WriteInternalErrorResponse(w, err)
		return
	}
	prevBlockNumberStr := strconv.FormatUint(prevBlockNumber, 10)
	setLinkHeader(r.URL, w, prevBlockNumberStr)

	var response = []BlockInfo{}
	for _, block := range blocks {
		response = append(response, blockInfoResponse(block))
	}
	api.rw.WriteResponse(w, response)
}

func blockInfoResponse(block *domain.BlockInfo) BlockInfo {
	return BlockInfo{
		PartitionID:        block.PartitionID,
		PartitionTypeID:    block.PartitionTypeID,
		ShardID:            block.ShardID,
		ProposerID:         block.ProposerID,
		PreviousBlockHash:  block.PreviousBlockHash,
		TxHashes:           block.TxHashes,
		UnicityCertificate: block.UnicityCertificate,
		BlockNumber:        block.BlockNumber,
	}
}
