package supervisor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

func init() {
	// A gnoland built with the goleveldb tag accepts db_backend = "goleveldb";
	// register the name so the config validates and OpenNode's own check is
	// the one that answers.
	dbm.InternalRegisterDBCreator(dbm.GoLevelDBBackend, func(string, string) (dbm.DB, error) {
		panic("not opened in tests")
	}, false)
}

func writeNodeConfig(t *testing.T, dataDir string, edit func(*config.Config)) {
	t.Helper()
	cfg := config.DefaultConfig()
	if edit != nil {
		edit(cfg)
	}
	require.NoError(t, os.MkdirAll(filepath.Join(dataDir, "config"), 0o755))
	require.NoError(t, config.WriteConfigFile(filepath.Join(dataDir, "config", "config.toml"), cfg))
}

func TestCommittedHeight(t *testing.T) {
	t.Parallel()
	dataDir := t.TempDir()
	writeNodeConfig(t, dataDir, nil)

	n, err := OpenNode(dataDir)
	require.NoError(t, err)
	h, err := n.CommittedHeight()
	require.NoError(t, err)
	assert.Equal(t, int64(0), h, "fresh node")
	_, err = os.Stat(filepath.Join(dataDir, "db"))
	assert.ErrorIs(t, err, os.ErrNotExist, "reading the height must not create the database")

	// Save the state the way tm2 does after committing a block.
	db, err := pebbledb.NewPebbleDB("state", filepath.Join(dataDir, "db"))
	require.NoError(t, err)
	sm.SaveState(db, sm.State{ChainID: testChainID, InitialHeight: 1, LastBlockHeight: 36300})
	require.NoError(t, db.Close())

	h, err = n.CommittedHeight()
	require.NoError(t, err)
	assert.Equal(t, int64(36300), h)

	require.NoError(t, n.CheckChain(testChainID))
	assert.ErrorContains(t, n.CheckChain("gnoland-1"), `is for chain "test-chain-1"`)
}

func TestCommittedHeightNoConfigYet(t *testing.T) {
	t.Parallel()
	// gnoland start --lazy writes the config on first start: tm2's defaults
	// apply until then.
	dataDir := t.TempDir()
	n, err := OpenNode(dataDir)
	require.NoError(t, err)
	h, err := n.CommittedHeight()
	require.NoError(t, err)
	assert.Equal(t, int64(0), h)
}

func TestOpenNodeRefuses(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		edit func(*config.Config)
		want string
	}{
		"config halt_height":  {func(c *config.Config) { c.HaltHeight = 1000 }, "halt_height is 1000"},
		"skip_upgrade_height": {func(c *config.Config) { c.SkipUpgradeHeight = 1183000 }, "skip_upgrade_height is 1183000"},
		"other db backend":    {func(c *config.Config) { c.DBBackend = "goleveldb" }, "pebbledb nodes only"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dataDir := t.TempDir()
			writeNodeConfig(t, dataDir, tc.edit)
			_, err := OpenNode(dataDir)
			assert.ErrorContains(t, err, tc.want)
		})
	}
}

func TestLock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f, err := Lock(dir)
	require.NoError(t, err)
	_, err = Lock(dir)
	assert.ErrorContains(t, err, "another supervisor")
	require.NoError(t, f.Close())
	f, err = Lock(dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
}

func TestLiveHeight(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"1651500"}}}`))
		case "/abci_query":
			require.Equal(t, `"params/node:p:halt_height"`, r.URL.Query().Get("path"))
			// What a node answers: the param as a JSON string, base64-encoded.
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"response":{"ResponseBase":{"Error":null,"Data":"IjE2NTE1MDAi"}}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	port := srv.URL[strings.LastIndex(srv.URL, ":")+1:]

	for laddr, ok := range map[string]bool{
		"tcp://127.0.0.1:" + port:  true,
		"tcp://0.0.0.0:" + port:    true, // listening everywhere: ask locally
		"unix:///tmp/gnoland.sock": false,
	} {
		dataDir := t.TempDir()
		writeNodeConfig(t, dataDir, func(c *config.Config) { c.RPC.ListenAddress = laddr })
		n, err := OpenNode(dataDir)
		require.NoError(t, err)
		live := n.LiveHeight()
		if !ok {
			assert.Nil(t, live, laddr)
			continue
		}
		require.NotNil(t, live, laddr)
		h, err := live(context.Background())
		require.NoError(t, err, laddr)
		assert.Equal(t, int64(1651500), h, laddr)
		hh, err := n.HaltParam()(context.Background())
		require.NoError(t, err, laddr)
		assert.Equal(t, int64(1651500), hh, laddr)
	}
}
