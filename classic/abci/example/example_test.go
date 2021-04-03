package example

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/classic/libs/log"

	abcicli "github.com/tendermint/classic/abci/client"
	"github.com/tendermint/classic/abci/example/kvstore"
	abciserver "github.com/tendermint/classic/abci/server"
	abci "github.com/tendermint/classic/abci/types"
)

func TestKVStore(t *testing.T) {
	fmt.Println("### Testing KVStore")
	testStream(t, kvstore.NewKVStoreApplication())
}

func TestBaseApp(t *testing.T) {
	fmt.Println("### Testing BaseApp")
	testStream(t, abci.NewBaseApplication())
}

func testStream(t *testing.T, app abci.Application) {
	numDeliverTxs := 20000

	// Start the listener
	server := abciserver.NewSocketServer("unix://test.sock", app)
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	if err := server.Start(); err != nil {
		require.NoError(t, err, "Error starting socket server")
	}
	defer server.Stop()

	// Connect to the socket
	client := abcicli.NewSocketClient("unix://test.sock", false)
	client.SetLogger(log.TestingLogger().With("module", "abci-client"))
	if err := client.Start(); err != nil {
		t.Fatalf("Error starting socket client: %v", err.Error())
	}
	defer client.Stop()

	done := make(chan struct{})
	counter := 0
	client.SetResponseCallback(func(req abci.Request, res abci.Response) {
		// Process response
		switch res := res.(type) {
		case abci.ResponseDeliverTx:
			counter++
			if res.Error != nil {
				t.Error("DeliverTx failed with error", res.Error)
			}
			if counter > numDeliverTxs {
				t.Fatalf("Too many DeliverTx responses. Got %d, expected %d", counter, numDeliverTxs)
			}
			if counter == numDeliverTxs {
				go func() {
					time.Sleep(time.Second * 1) // Wait for a bit to allow counter overflow
					close(done)
				}()
				return
			}
		case abci.ResponseFlush:
			// ignore
		default:
			t.Error("Unexpected response type", reflect.TypeOf(res))
		}
	})

	// Write requests
	for counter := 0; counter < numDeliverTxs; counter++ {
		// Send request
		reqRes := client.DeliverTxAsync(abci.RequestDeliverTx{Tx: []byte("test")})
		_ = reqRes
		// check err ?

		// Sometimes send flush messages
		if counter%123 == 0 {
			client.FlushAsync()
			// check err ?
		}
	}

	// Send final flush message
	client.FlushAsync()

	<-done
}
