package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	partitionID1 = types.PartitionID(1)
	partitionID2 = types.PartitionID(2)
)

func TestGetBlock_Success(t *testing.T) {
	r := mux.NewRouter()
	restapi := &RestAPI{Service: &MockExplorerBackendService{
		getBlockFunc: func(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
			require.EqualValues(t, 1, blockNumber)
			require.Len(t, partitionIDs, 1)
			require.EqualValues(t, partitionID1, partitionIDs[0])
			blockMap := make(map[types.PartitionID]*domain.BlockInfo)
			blockMap[1] = &domain.BlockInfo{TxHashes: []domain.TxHash{{0xFF}}}
			return blockMap, nil
		},
	}}
	r.HandleFunc("/blocks/{blockNumber}", restapi.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/1?partitionID=%d", ts.URL, partitionID1))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var result map[types.PartitionID]BlockInfo
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result, 1)
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0xFF}}}, result[partitionID1])
}

func TestGetBlock_Success_MultiplePartitions(t *testing.T) {
	r := mux.NewRouter()
	restapi := &RestAPI{Service: &MockExplorerBackendService{
		getBlockFunc: func(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
			require.EqualValues(t, 1, blockNumber)
			require.Len(t, partitionIDs, 2)
			require.EqualValues(t, partitionID1, partitionIDs[0])
			require.EqualValues(t, partitionID2, partitionIDs[1])
			blockMap := make(map[types.PartitionID]*domain.BlockInfo)
			blockMap[partitionID1] = &domain.BlockInfo{TxHashes: []domain.TxHash{{0xFF}}}
			blockMap[partitionID2] = &domain.BlockInfo{TxHashes: []domain.TxHash{{0xAA}}}
			return blockMap, nil
		},
	}}
	r.HandleFunc("/blocks/{blockNumber}", restapi.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/1?partitionID=%d&partitionID=%d", ts.URL, partitionID1, partitionID2))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var result map[types.PartitionID]BlockInfo
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result, 2)
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0xFF}}}, result[partitionID1])
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0xAA}}}, result[partitionID2])
}

func TestGetBlock_latest_Success(t *testing.T) {
	r := mux.NewRouter()
	restapi := &RestAPI{Service: &MockExplorerBackendService{
		getLastBlocksFunc: func(ctx context.Context, partitionIDs []types.PartitionID, count int, includeEmpty bool) (map[types.PartitionID][]*domain.BlockInfo, error) {
			blockMap := make(map[types.PartitionID][]*domain.BlockInfo)
			blockMap[partitionID1] = []*domain.BlockInfo{{TxHashes: []domain.TxHash{{0xFF}}}}
			return blockMap, nil
		},
	}}
	r.HandleFunc("/blocks/{blockNumber}", restapi.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/latest?partitionID=%d", ts.URL, partitionID1))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var result map[types.PartitionID]BlockInfo
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result, 1)
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0xFF}}}, result[partitionID1])
}

func TestGetBlock_InvalidBlockNumber(t *testing.T) {
	r := mux.NewRouter()
	api := &RestAPI{Service: &MockExplorerBackendService{}}
	r.HandleFunc("/blocks/{blockNumber}", api.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/invalid", ts.URL))
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// Check the error message
	require.Contains(t, string(body), "invalid 'blockNumber' parameter")
}

func TestGetBlock_FailedToLoadBlock(t *testing.T) {
	r := mux.NewRouter()
	api := &RestAPI{Service: &MockExplorerBackendService{
		getBlockFunc: func(ctx context.Context, blockNumber uint64, partitionIDs []types.PartitionID) (map[types.PartitionID]*domain.BlockInfo, error) {
			return nil, fmt.Errorf("failed to load block")
		},
	}}
	r.HandleFunc("/blocks/{blockNumber}", api.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/1", ts.URL))
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// Check the error message
	require.Contains(t, string(body), "internal error")
}

func TestGetBlocks_Success(t *testing.T) {
	r := mux.NewRouter()
	restapi := &RestAPI{Service: &MockExplorerBackendService{
		getBlocksInRangeFunc: func(
			ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool,
		) (res []*domain.BlockInfo, prevBlockNumber uint64, err error) {
			require.Equal(t, true, includeEmpty)
			return []*domain.BlockInfo{
				{}, // empty block
				{TxHashes: []domain.TxHash{{0x02}}},
				{TxHashes: []domain.TxHash{{0x03}}},
			}, 0, nil
		},
	}}
	r.HandleFunc("/partitions/{partitionID}/blocks", restapi.getBlocksInRange)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/partitions/%d/blocks?startBlock=1&limit=10", ts.URL, partitionID1))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	result := make([]BlockInfo, 3)
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result, 3)
	require.Equal(t, BlockInfo{}, result[0])
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0x02}}}, result[1])
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0x03}}}, result[2])

	require.Contains(t, res.Header.Get("Link"), "offsetKey=0")
}

func TestGetBlocks_Success_ExcludeEmpty(t *testing.T) {
	r := mux.NewRouter()
	restapi := &RestAPI{Service: &MockExplorerBackendService{
		getBlocksInRangeFunc: func(
			ctx context.Context, partitionID types.PartitionID, dbStartBlock uint64, count int, includeEmpty bool,
		) (res []*domain.BlockInfo, prevBlockNumber uint64, err error) {
			require.Equal(t, false, includeEmpty)
			return []*domain.BlockInfo{
				{TxHashes: []domain.TxHash{{0x02}}},
				{TxHashes: []domain.TxHash{{0x03}}},
			}, 0, nil
		},
	}}
	r.HandleFunc("/partitions/{partitionID}/blocks", restapi.getBlocksInRange)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/partitions/%d/blocks?startBlock=1&limit=10&includeEmpty=false", ts.URL, partitionID1))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	result := make([]BlockInfo, 2)
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result, 2)
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0x02}}}, result[0])
	require.Equal(t, BlockInfo{TxHashes: []domain.TxHash{{0x03}}}, result[1])

	require.Contains(t, res.Header.Get("Link"), "offsetKey=0")
}
