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
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alphabill-org/alphabill-explorer-backend/api"
	"github.com/alphabill-org/alphabill-explorer-backend/bill_store"
	"github.com/alphabill-org/alphabill-explorer-backend/restapi"
	"github.com/alphabill-org/alphabill-wallet/cli/alphabill/cmd/wallet/args"
	"github.com/alphabill-org/alphabill-wallet/client/rpc"
	"github.com/alphabill-org/alphabill-wallet/wallet/account"
	"github.com/alphabill-org/alphabill-wallet/wallet/fees"
	wallet "github.com/alphabill-org/alphabill-wallet/wallet/money"
	"github.com/alphabill-org/alphabill/txsystem/money"
	"github.com/stretchr/testify/require"
)

const abMoneyRpcUrl = "https://money-partition.testnet.alphabill.org"

func TestE2E(t *testing.T) {
	startTime := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	port := findFreePort(t)
	host := fmt.Sprintf("localhost:%s", port)

	w := createMoneyWallet(t, ctx, t.TempDir())
	rn, err := w.GetRoundNumber(ctx)
	require.NoError(t, err)

	fmt.Printf("Round number: %d\n", rn)

	_ = runService(t, ctx, host, rn)
	awaitStartup(t, host)

	client := http.Client{Timeout: 5 * time.Second}

	fmt.Printf("Started in %s\n", time.Since(startTime))

	t.Run("Check first key's balance", func(t *testing.T) {
		// Check first key's balance
		balance, err := w.GetBalance(ctx, wallet.GetBalanceCmd{AccountIndex: 0})
		require.NoError(t, err)
		require.Greater(t, balance, uint64(0))
		fmt.Printf("Balance: %d\n", balance)
	})

	t.Run("ensure all tx records are indexed and returned by the explorer API", func(t *testing.T) {
		pk1, err := w.GetAccountManager().GetPublicKey(1)
		require.NoError(t, err)
		// send 1 unit to the second key
		proofs, err := w.Send(ctx, wallet.SendCmd{Receivers: []wallet.ReceiverData{{PubKey: pk1, Amount: 1}}, WaitForConfirmation: true, AccountIndex: 0})
		require.NoError(t, err)
		require.NotEmpty(t, proofs)

		for _, proof := range proofs {
			txRn := proof.TxProof.UnicityCertificate.GetRoundNumber()
			blockInfo := &api.BlockInfo{}
			require.Eventually(t, func() bool {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/blocks/%d", host, txRn))
				require.NoError(t, err)
				fmt.Printf("Checking block %d, status code: %d\n", txRn, resp.StatusCode)
				if resp.StatusCode == http.StatusOK {
					require.NoError(t, restapi.DecodeResponse(resp, http.StatusOK, blockInfo, false))
					return true
				}
				return false
			}, 10*time.Second, 100*time.Millisecond, "should index tx record")

			txrHash := proof.TxRecord.Hash(crypto.SHA256)

			t.Run("Check tx record hash is in the block info", func(t *testing.T) {
				require.Contains(t, blockInfo.TxHashes, api.TxRecordHash(txrHash))
			})

			t.Run("Check tx info is correct", func(t *testing.T) {
				resp, err := client.Get(fmt.Sprintf("http://%s/api/v1/txs/0x%X", host, txrHash))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				txInfo := &api.TxInfo{}
				err = restapi.DecodeResponse(resp, http.StatusOK, txInfo, false)
				require.NoError(t, err)
				require.Equal(t, txrHash, txInfo.TxRecordHash)
				require.Equal(t, txRn, txInfo.BlockNumber)
				require.Equal(t, proof.TxRecord, txInfo.Transaction)
				fmt.Printf("Tx record %X indexed, type: %s\n", txrHash, txInfo.Transaction.TransactionOrder.PayloadType())
			})
		}
	})

	fmt.Printf("Finished after %s\n", time.Since(startTime))
}

func createMoneyWallet(t *testing.T, ctx context.Context, walletDir string) *wallet.Wallet {
	am, err := account.NewManager(walletDir, "", true)
	require.NoError(t, err)

	err = wallet.CreateNewWallet(am, "prison tone orbit inside kitten clean page enrich plastic ring gather cross")
	require.NoError(t, err)

	feeManagerDB, err := fees.NewFeeManagerDB(walletDir)
	require.NoError(t, err)

	moneyClient, err := rpc.DialContext(ctx, args.BuildRpcUrl(abMoneyRpcUrl))
	require.NoError(t, err)

	w, err := wallet.LoadExistingWallet(am, feeManagerDB, moneyClient, slog.Default())
	require.NoError(t, err)

	_, _, err = w.GetAccountManager().AddAccount()
	require.NoError(t, err)

	return w
}

func runService(t *testing.T, ctx context.Context, host string, startFromBlock uint64) *sync.WaitGroup {
	os.Args = []string{t.TempDir()}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NotPanics(t, func() {
			err := Run(ctx, &Config{
				ABMoneySystemIdentifier: money.DefaultSystemIdentifier,
				AlphabillUrl:            abMoneyRpcUrl,
				ServerAddr:              host,
				DbFile:                  filepath.Join(t.TempDir(), bill_store.BoltExplorerStoreFileName),
				BlockNumber:             startFromBlock,
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
