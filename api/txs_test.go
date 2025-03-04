package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	mocks "github.com/alphabill-org/alphabill-explorer-backend/internal/mocks/api"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTxByHash_Success(t *testing.T) {
	txHash := domain.TxHash([]byte{1, 2, 3, 4})
	r := mux.NewRouter()
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetTxByHash(mock.Anything, txHash).Return(&domain.TxInfo{TxRecordHash: txHash}, nil)
	restapi := &Controller{StorageService: mockStorage}
	r.HandleFunc("/txs/{txHash}", restapi.getTx)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/txs/0x%s", ts.URL, txHash.String()))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var result TxInfo
	require.NoError(t, json.Unmarshal(body, &result))
	require.NotNil(t, result)
	require.Equal(t, TxInfo{TxRecordHash: txHash}, result)
}

func TestGetTxs_Success(t *testing.T) {
	r := mux.NewRouter()
	mockStorage := mocks.NewStorageService(t)
	mockStorage.EXPECT().GetTxsPage(mock.Anything, partitionID1, "", defaultTxsPageLimit).
		Return([]*domain.TxInfo{
			{TxRecordHash: []byte{0x01}},
			{TxRecordHash: []byte{0x02}},
			{TxRecordHash: []byte{0x03}},
		}, "xxx", nil)
	restapi := &Controller{StorageService: mockStorage}
	r.HandleFunc("/partitions/{partitionID}/txs", restapi.getTxs)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/partitions/%d/txs", ts.URL, partitionID1))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var result []TxInfo
	require.NoError(t, json.Unmarshal(body, &result))
	require.NotNil(t, result)
	require.Len(t, result, 3)

	require.Contains(t, res.Header.Get("Link"), "offsetKey=xxx")
}
