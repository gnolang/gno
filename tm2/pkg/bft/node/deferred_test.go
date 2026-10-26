package node

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// deferringKVStore is a kvstore that reports when it is asked to defer.
type deferringKVStore struct {
	*kvstore.KVStoreApplication
	once     sync.Once
	deferred chan struct{}
}

var _ abci.DeferredCommitter = (*deferringKVStore)(nil)

func (app *deferringKVStore) DeferCommits(on bool) bool {
	if on {
		app.once.Do(func() { close(app.deferred) })
	}
	return on
}

func (app *deferringKVStore) WriteCommits() error   { return nil }
func (app *deferringKVStore) HeldBytes() int64      { return 0 }
func (app *deferringKVStore) NeedsFlushPoint() bool { return false }

// TestNodeDefersBlockSyncWrites pins that a node whose in-process app defers
// its commits holds the writes of its block store and state during block
// sync, and that it stops promptly while block sync has not caught up.
func TestNodeDefersBlockSyncWrites(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_deferred_test")
	defer os.RemoveAll(config.RootDir)

	// The only validator is someone else, so the node block-syncs.
	genDoc, err := types.GenesisDocFromFile(genesisFile)
	require.NoError(t, err)
	genDoc.Validators = []types.GenesisValidator{{PubKey: ed25519.GenPrivKey().PubKey(), Power: 10}}
	require.NoError(t, genDoc.SaveAs(genesisFile))

	app := &deferringKVStore{KVStoreApplication: kvstore.NewKVStoreApplication(), deferred: make(chan struct{})}
	config.LocalApp = app
	config.FastSyncMode = true
	n, err := DefaultNewNode(config, genesisFile, events.NewEventSwitch(), log.NewTestingLogger(t))
	require.NoError(t, err)
	require.NoError(t, n.Start())
	stopped := make(chan struct{})
	t.Cleanup(func() {
		if n.IsRunning() {
			n.Stop()
		}
	})

	select {
	case <-app.deferred:
	case <-time.After(10 * time.Second):
		t.Fatal("block sync did not defer the app's commits")
	}
	// Block sync asks the app first, and holds the node's databases right after.
	require.Eventually(t, func() bool {
		return n.blockStore.DB().(*dbm.HoldingDB).Holding() && n.stateDB.(*dbm.HoldingDB).Holding()
	}, 10*time.Second, 5*time.Millisecond, "the node's databases do not hold")

	go func() {
		n.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("the node did not stop while block-syncing")
	}
}
