//go:build manual

package main

import (
	"context"
	"crypto"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/restapi"
	"github.com/alphabill-org/alphabill-explorer-backend/util"
	"github.com/alphabill-org/alphabill-go-base/types"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client"
	sdktypes "github.com/alphabill-org/alphabill-wallet/client/types"
	"github.com/alphabill-org/alphabill-wallet/wallet/account"
	"github.com/alphabill-org/alphabill-wallet/wallet/fees"
	wallet "github.com/alphabill-org/alphabill-wallet/wallet/money"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	mongocontainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	abMoneyRpcUrl        = "dev-ab-money.abdev1.guardtime.com/rpc"
	abMoneyArchiveRpcUrl = "dev-ab-money-archive.abdev1.guardtime.com/rpc"
	partitionID          = types.PartitionID(1)
	mongoDBImage         = "mongo:7.0"
	maxFee               = 10
)

func TestE2E(t *testing.T) {
	startTime := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	dbConnectionString := startDB(t, ctx)

	port := findFreePort(t)
	host := fmt.Sprintf("localhost:%s", port)

	fmt.Printf("Starting server on %s\n", host)

	w := createMoneyWallet(t, ctx, t.TempDir())
	rn, err := w.GetRoundNumber(ctx)
	require.NoError(t, err)

	fmt.Printf("Round number: %d\n", rn)

	_ = runService(t, ctx, host, dbConnectionString, rn-1)
	awaitStartup(t, host)

	client := http.Client{Timeout: 5 * time.Second}

	fmt.Printf("Started in %s\n", time.Since(startTime))

	t.Run("Check first key's balance", func(t *testing.T) {
		// Check first key's balance
		balance, err := w.GetBalance(ctx, wallet.GetBalanceCmd{AccountIndex: 0})
		require.NoError(t, err)
		fmt.Printf("Balance: %d\n", balance)
		require.Greater(t, balance, uint64(0))
	})

	t.Run("Check first key's fee credit record", func(t *testing.T) {
		var fcr *sdktypes.FeeCreditRecord
		fcr, err = w.GetFeeCredit(ctx, fees.GetFeeCreditCmd{AccountIndex: 0})
		require.NoError(t, err)
		if fcr == nil || fcr.Balance < maxFee*10 {
			_, err := w.AddFeeCredit(ctx, fees.AddFeeCmd{AccountIndex: 0, Amount: 100})
			require.NoError(t, err)
			fcr, err = w.GetFeeCredit(ctx, fees.GetFeeCreditCmd{AccountIndex: 0})
			require.NoError(t, err)
		}
		fmt.Printf("FCR balance: %d\n", fcr.Balance)
	})

	t.Run("ensure all tx records are indexed and returned by the explorer API", func(t *testing.T) {
		pk1, err := w.GetAccountManager().GetPublicKey(1)
		require.NoError(t, err)
		// send 1 unit to the second key
		proofs, err := w.Send(ctx, wallet.SendCmd{Receivers: []wallet.ReceiverData{{PubKey: pk1, Amount: 1}}, WaitForConfirmation: true, AccountIndex: 0, MaxFee: maxFee})
		require.NoError(t, err)
		require.NotEmpty(t, proofs)

		for _, proof := range proofs {
			unicityCertificate := types.UnicityCertificate{}
			err := unicityCertificate.UnmarshalCBOR(proof.TxProof.UnicityCertificate)
			require.NoError(t, err)
			txRn := unicityCertificate.GetRoundNumber()

			blockMap := make(map[types.PartitionID]restapi.BlockInfo)
			require.Eventually(t, func() bool {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/blocks/%d", host, txRn))
				require.NoError(t, err)
				fmt.Printf("Checking block %d, status code: %d\n", txRn, resp.StatusCode)
				if resp.StatusCode == http.StatusOK {
					require.NoError(t, restapi.DecodeResponse(resp, http.StatusOK, &blockMap, false))
					return true
				}
				return false
			}, 20*time.Second, 100*time.Millisecond, "should index tx record")

			txrHash, err := proof.TxRecord.Hash(crypto.SHA256)
			require.NoError(t, err)

			blockInfo := blockMap[partitionID]
			require.NotNil(t, blockInfo)
			require.Contains(t, blockInfo.TxHashes, domain.TxHash(txrHash))

			txInfo := &restapi.TxInfo{}
			t.Run("Check tx info is correct", func(t *testing.T) {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/txs/0x%X", host, txrHash))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				err = restapi.DecodeResponse(resp, http.StatusOK, txInfo, false)
				require.NoError(t, err)
				require.EqualValues(t, txrHash, txInfo.TxRecordHash)
				require.Equal(t, txRn, txInfo.BlockNumber)
				require.Equal(t, proof.TxRecord, txInfo.Transaction)
				txOrder := types.TransactionOrder{}
				require.NoError(t, txOrder.UnmarshalCBOR(txInfo.Transaction.TransactionOrder))
				fmt.Printf("Tx record %X indexed, type: %d\n", txrHash, txOrder.Payload.Type)
			})

			t.Run("check latest transactions to contain the tx", func(t *testing.T) {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/partitions/%s/txs", host, partitionID))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				txInfos := make([]restapi.TxInfo, 0)
				err = restapi.DecodeResponse(resp, http.StatusOK, &txInfos, false)
				require.NoError(t, err)
				require.Contains(t, txInfos, *txInfo)
			})

			t.Run("check tx returned in search", func(t *testing.T) {
				txHashHex := util.ToHex(txrHash)
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/search?q=%s", host, txHashHex))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				searchResponse := restapi.SearchResponse{}
				err = restapi.DecodeResponse(resp, http.StatusOK, &searchResponse, false)
				require.NoError(t, err)
				require.Contains(t, searchResponse.Txs, *txInfo)
			})

			t.Run("check block returned in search", func(t *testing.T) {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/search?q=%d", host, blockInfo.BlockNumber))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				searchResponse := restapi.SearchResponse{}
				err = restapi.DecodeResponse(resp, http.StatusOK, &searchResponse, false)
				require.NoError(t, err)
				require.NotNil(t, searchResponse.Blocks[partitionID])
				require.Equal(t, searchResponse.Blocks[partitionID].BlockNumber, blockInfo.BlockNumber)
			})
		}
	})

	fmt.Printf("Finished after %s\n", time.Since(startTime))
}

func createMoneyWallet(t *testing.T, ctx context.Context, walletDir string) *wallet.Wallet {
	am, err := account.NewManager(walletDir, "", true)
	require.NoError(t, err)

	feeManagerDB, err := fees.NewFeeManagerDB(walletDir)
	require.NoError(t, err)

	moneyClient, err := client.NewMoneyPartitionClient(ctx, args.BuildRpcUrl(abMoneyRpcUrl))
	require.NoError(t, err)

	w, err := wallet.NewWallet(ctx, am, feeManagerDB, moneyClient, maxFee, slog.Default())
	require.NoError(t, err)

	err = wallet.GenerateKeys(am, "prison tone orbit inside kitten clean page enrich plastic ring gather cross")
	require.NoError(t, err)

	_, _, err = am.AddAccount()
	require.NoError(t, err)

	keys, err := am.GetPublicKeys()
	require.NoError(t, err)

	for idx, key := range keys {
		fmt.Printf("Account #%d Pubkey: 0x%X\n", idx, key)
	}

	return w
}

func runService(t *testing.T, ctx context.Context, host string, dbConnectionString string, startFromBlock uint64) *sync.WaitGroup {
	os.Args = []string{t.TempDir()}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NotPanics(t, func() {
			err := Run(ctx, &Config{
				Nodes:  []Node{{URL: abMoneyArchiveRpcUrl, BlockNumber: startFromBlock}},
				Server: Server{Address: host},
				DB:     DB{URL: dbConnectionString},
			})
			require.NoError(t, err)
		}, "should not panic")
	}()

	return &wg
}

func awaitStartup(t *testing.T, host string) {
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", host))
		if err != nil {
			return false
		}
		if resp.StatusCode == http.StatusOK {
			return true
		}
		return false
	}, 60*time.Second, 100*time.Millisecond, "should start")
}

func findFreePort(t *testing.T) string {
	// Bind to port 0 to get a random available port
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	// Extract the port from the listener's address
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	return port
}

func startDB(t *testing.T, ctx context.Context) string {
	mongoContainer, err := mongocontainer.Run(ctx, mongoDBImage, testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")))
	require.NoError(t, err)
	connectionString, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)
	return connectionString
}
