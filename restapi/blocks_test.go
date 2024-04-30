package restapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGetBlock_Success(t *testing.T) {
	r := mux.NewRouter()
	restapi := &MoneyRestAPI{Service: &MockExplorerBackendService{
		getBlockFunc: func(blockNumber uint64) (*api.BlockInfo, error) {
			require.Equal(t, uint64(1), blockNumber)
			return &api.BlockInfo{TxHashes: []string{"FF"}}, nil
		},
	}}
	r.HandleFunc("/blocks/{blockNumber}", restapi.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/1", ts.URL))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	result := &api.BlockInfo{}
	require.NoError(t, json.Unmarshal(body, result))
	require.Equal(t, &api.BlockInfo{TxHashes: []string{"FF"}}, result)
}

func TestGetBlock_InvalidBlockNumber(t *testing.T) {
	r := mux.NewRouter()
	api := &MoneyRestAPI{Service: &MockExplorerBackendService{}}
	r.HandleFunc("/blocks/{blockNumber}", api.getBlock)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks/invalid", ts.URL))
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// Check the error message
	require.Contains(t, string(body), "invalid 'blockNumber' format")
}

func TestGetBlock_FailedToLoadBlock(t *testing.T) {
	r := mux.NewRouter()
	api := &MoneyRestAPI{Service: &MockExplorerBackendService{
		getBlockFunc: func(blockNumber uint64) (*api.BlockInfo, error) {
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
	require.Contains(t, string(body), "failed to load block with block number 1")
}

func TestGetBlocks_Success(t *testing.T) {
	r := mux.NewRouter()
	restapi := &MoneyRestAPI{Service: &MockExplorerBackendService{
		getBlocksFunc: func(dbStartBlock uint64, count int) (res []*api.BlockInfo, prevBlockNumber uint64, err error) {
			return []*api.BlockInfo{{TxHashes: []string{"AA"}}}, 0, nil
		},
		getLastBlockNumberFunc: func() (uint64, error) {
			return 0, nil
		},
	}}
	r.HandleFunc("/blocks", restapi.getBlocks)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/blocks?startBlock=1&limit=10", ts.URL))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	result := make([]*api.BlockInfo, 0)
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, []*api.BlockInfo{{TxHashes: []string{"AA"}}}, result)

	require.Contains(t, res.Header.Get("Link"), "offsetKey=0")
}
