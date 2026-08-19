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
// has InitialHeight = 101.  It verifies that:
//
//   - The node starts without panicking (exercises all the InitialHeight paths
//     through Handshaker → ConsensusState.reconstructLastCommit →
//     BlockchainReactor.NewBlockchainReactor).
//   - The first committed block is at height 101, not 1.
func TestNodeBootWithInitialHeight(t *testing.T) {
	const initialHeight = int64(101)

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

	bs := n.BlockStore()
	require.NotNil(t, bs.LoadBlockMeta(initialHeight),
		"no block committed at InitialHeight (%d); block store height is %d", initialHeight, bs.Height())
	require.Nil(t, bs.LoadBlockMeta(initialHeight-1),
		"a block was committed below InitialHeight (%d); the node did not start from it", initialHeight)
}
