package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	mocks "github.com/alphabill-org/alphabill-explorer-backend/internal/mocks/github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	partitionID1 = types.PartitionID(1)
	partitionID2 = types.PartitionID(2)
)

func TestGetBlock_Success(t *testing.T) {
	r := mux.NewRouter()
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetBlock(mock.Anything, uint64(1), []types.PartitionID{partitionID1}).
		Return(map[types.PartitionID]*domain.BlockInfo{
			1: {TxHashes: []domain.TxHash{{0xFF}}},
		}, nil)

	restapi := &Controller{StorageService: mockStorage}
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
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetBlock(mock.Anything, uint64(1), []types.PartitionID{partitionID1, partitionID2}).
		Return(map[types.PartitionID]*domain.BlockInfo{
			partitionID1: {TxHashes: []domain.TxHash{{0xFF}}},
			partitionID2: {TxHashes: []domain.TxHash{{0xAA}}},
		}, nil)

	restapi := &Controller{StorageService: mockStorage}
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
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetLastBlocks(mock.Anything, []types.PartitionID{partitionID1}, 1, true).
		Return(map[types.PartitionID][]*domain.BlockInfo{
			partitionID1: {{TxHashes: []domain.TxHash{{0xFF}}}},
		}, nil)

	restapi := &Controller{StorageService: mockStorage}
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
	api := &Controller{StorageService: mocks.NewStorageService(t)}
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
	mockStorage := mocks.NewStorageService(t)
	var partitionIDs []types.PartitionID
	mockStorage.EXPECT().GetBlock(mock.Anything, uint64(1), partitionIDs).
		Return(nil, fmt.Errorf("failed to load block"))

	api := &Controller{StorageService: mockStorage}
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
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetBlocksInRange(mock.Anything, partitionID1, uint64(1), 10, true).
		Return([]*domain.BlockInfo{
			{}, // empty block
			{TxHashes: []domain.TxHash{{0x02}}},
			{TxHashes: []domain.TxHash{{0x03}}},
		}, 0, nil)

	restapi := &Controller{StorageService: mockStorage}
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
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetBlocksInRange(mock.Anything, partitionID1, uint64(1), 10, false).
		Return([]*domain.BlockInfo{
			{TxHashes: []domain.TxHash{{0x02}}},
			{TxHashes: []domain.TxHash{{0x03}}},
		}, 0, nil)

	restapi := &Controller{StorageService: mockStorage}
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
