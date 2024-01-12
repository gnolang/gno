package client_test

import (
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	nm "github.com/gnolang/gno/tm2/pkg/bft/node"
	rpctest "github.com/gnolang/gno/tm2/pkg/bft/rpc/test"
)

var node *nm.Node

func TestMain(m *testing.M) {
	// start a tendermint node (and kvstore) in the background to test against
	dir, err := os.MkdirTemp("/tmp", "rpc-client-test")
	if err != nil {
		panic(err)
	}
	app := kvstore.NewPersistentKVStoreApplication(dir)
	node = rpctest.StartTendermint(app)

	code := m.Run()

	// and shut down proper at the end
	rpctest.StopTendermint(node)
	os.Exit(code)
}
