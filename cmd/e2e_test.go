//go:build manual

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alphabill-org/alphabill-explorer-backend/bill_store"
	"github.com/alphabill-org/alphabill/txsystem/money"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	startTime := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	port := findFreePort(t)
	host := fmt.Sprintf("localhost:%s", port)
	_ = runService(t, ctx, host)
	awaitStartup(t, host)

	fmt.Printf("Started in %s\n", time.Since(startTime))

	// TODO: Add tests here

	fmt.Printf("Finished after %s\n", time.Since(startTime))
}

func runService(t *testing.T, ctx context.Context, host string) *sync.WaitGroup {
	os.Args = []string{t.TempDir()}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NotPanics(t, func() {
			err := Run(ctx, &Config{
				ABMoneySystemIdentifier: money.DefaultSystemIdentifier,
				AlphabillUrl:            "https://money-partition.testnet.alphabill.org",
				ServerAddr:              host,
				DbFile:                  filepath.Join(t.TempDir(), bill_store.BoltExplorerStoreFileName),
				BlockNumber:             0,
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
