package gnoland

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// TestNodeBootWithInitialHeight boots a full in-memory node whose genesis doc
// has InitialHeight = 100.  It verifies that:
//
//   - The node starts without panicking (exercises all the InitialHeight paths
//     through Handshaker → ConsensusState.reconstructLastCommit →
//     BlockchainReactor.NewBlockchainReactor).
//   - The first committed block is at height 100, not 1.
func TestNodeBootWithInitialHeight(t *testing.T) {
	const initialHeight = int64(100)

	td := t.TempDir()
	tmcfg := NewDefaultTMConfig(td)

	pv := bft.NewMockPV()
	pk := pv.PubKey()

	genesis := &bft.GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       tmcfg.ChainID(),
		InitialHeight: initialHeight,
		ConsensusParams: abci.ConsensusParams{
			Block: defaultBlockParams(),
		},
		Validators: []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "self",
			},
		},
		AppState: DefaultGenState(),
	}

	cfg := &InMemoryNodeConfig{
		PrivValidator: pv,
		Genesis:       genesis,
		TMConfig:      tmcfg,
		DB:            memdb.NewMemDB(),
		InitChainerConfig: InitChainerConfig{
			GenesisTxResultHandler: PanicOnFailingTxResultHandler,
			StdlibDir:              filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs"),
			CacheStdlibLoad:        true,
		},
	}

	n, err := NewInMemoryNode(log.NewTestingLogger(t), cfg)
	require.NoError(t, err)

	require.NoError(t, n.Start())
	t.Cleanup(func() { require.NoError(t, n.Stop()) })

	select {
	case <-n.Ready():
		// first block committed
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for node to produce first block")
	}

	// Assert on which heights exist in the block store, not on the current
	// tip: Ready() closes on the first EventNewBlock, delivered off the
	// consensus goroutine, so the node can already have committed the next
	// block by the time this runs.
	bs := n.BlockStore()
	require.GreaterOrEqual(t, bs.Height(), initialHeight)
	require.NotNil(t, bs.LoadBlock(initialHeight),
		"block at InitialHeight (%d) should be committed", initialHeight)
	require.Nil(t, bs.LoadBlock(initialHeight-1),
		"no block below InitialHeight (%d) should exist; the chain must not start at 1", initialHeight)
}

// AppState must be a GnoGenesisState value: loadAppState rejects a
// *GnoGenesisState as an invalid AppState.
func TestNewDefaultGenesisConfig_AppStateType(t *testing.T) {
	t.Parallel()

	genesis := NewDefaultGenesisConfig("test-chain", "gno.land")

	_, ok := genesis.AppState.(GnoGenesisState)
	require.True(t, ok, "AppState must be a GnoGenesisState value, got %T", genesis.AppState)
}
