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

	// Assert on which heights the store holds, not on its current tip. Ready()
	// closes when the FIRST block arrives, and TestConfig sets
	// SkipTimeoutCommit with CreateEmptyBlocks on, so the tip moves off
	// InitialHeight about 20ms later: reading it here is a race the test loses
	// on a loaded runner. Which height the chain started at never changes.
	bs := n.BlockStore()
	require.NotNil(t, bs.LoadBlockMeta(initialHeight),
		"no block at InitialHeight (%d); store tip is %d", initialHeight, bs.Height())
	require.Nil(t, bs.LoadBlockMeta(1),
		"a block exists at height 1, so the chain ignored InitialHeight (%d)", initialHeight)
	require.Nil(t, bs.LoadBlockMeta(initialHeight-1),
		"a block exists below InitialHeight (%d), so it is not the first", initialHeight)
}
